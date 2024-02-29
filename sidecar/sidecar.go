package sidecar

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"go.uber.org/zap"
)

var _ Sidecar = (*SidecarImpl)(nil)

// Sidecar defines the expected interface for a sidecar. It is consumed by the sidecar server.
type Sidecar interface {
	IsRunning() bool
	QueryBlockEvents(blockNumber *big.Int) ([]sidecartypes.Event, error)
	Start(ctx context.Context) error
	Stop()
}

// SidecarImpl is the core component responsible for fetching ethereum blocks and processing events.
type SidecarImpl struct {
	// --------------------- General Config --------------------- //
	logger *zap.Logger
	mu     sync.Mutex

	// --------------------- Ethereum Config --------------------- //
	client          *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	blocksMap       map[string]*EthereumBlock
	updateInterval  time.Duration

	// startQueryBlock is the block at which we started querying for events.
	startQueryBlock *big.Int
	// lastQueryBlock is last block we queried for events.
	lastQueryBlock *big.Int

	// running is the current status of the main sidecar process (running or not).
	running atomic.Bool
}

// NewSidecar creates a new Sidecar instance.
func NewSidecar(
	client *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
	startQueryBlock *big.Int,
	logger *zap.Logger,
) *SidecarImpl {
	return &SidecarImpl{
		logger:          logger,
		client:          client,
		contractAddress: contractAddress,
		contractABI:     contractAbi,
		startQueryBlock: startQueryBlock,
		blocksMap:       make(map[string]*EthereumBlock),
		updateInterval:  10 * time.Second,
	}
}

// Start begins the process of querying and storing events from the Ethereum blockchain.
func (s *SidecarImpl) Start(ctx context.Context) error {
	s.logger.Info("starting sidecar")

	// Initial check to verify Ethereum client connectivity and log fetching capability
	query := ethereum.FilterQuery{
		FromBlock: s.startQueryBlock,
		ToBlock:   s.startQueryBlock,
		Addresses: []common.Address{s.contractAddress},
	}

	// Attempt to fetch logs as a connectivity and configuration check
	_, err := s.client.FilterLogs(ctx, query)
	if err != nil {
		s.logger.Error("Failed to fetch logs for initial check", zap.Error(err))
		return err
	}

	s.running.Store(true)
	go s.queryAndStoreEvents(ctx)
	return nil
}

// Stop signals the sidecar to stop processing.
func (s *SidecarImpl) Stop() {
	s.logger.Info("stopping sidecar")
	s.running.Store(false)
}

// IsRunning checks if the sidecar process is currently running.
// It returns true if the sidecar is running, false otherwise.
func (s *SidecarImpl) IsRunning() bool {
	return s.running.Load()
}

// QueryBlockEvents queries the `blocksMap` for events associated with a specific block number.
func (s *SidecarImpl) QueryBlockEvents(blockNumber *big.Int) ([]sidecartypes.Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	blockNumberStr := blockNumber.String()
	block, exists := s.blocksMap[blockNumberStr]
	if !exists {

		// Check if the queried block is within the range of blocks that have been queried for events.
		if (s.startQueryBlock != nil && s.lastQueryBlock != nil) &&
			(blockNumber.Cmp(s.startQueryBlock) >= 0 && blockNumber.Cmp(s.lastQueryBlock) <= 0) {
			// The block was within the range we have queried, but no events were found, hence return an empty list.
			return []sidecartypes.Event{}, nil
		}

		// Otherwise this block was not yet processed
		return nil, fmt.Errorf("no events found for block number %s", blockNumber)
	}

	return block.Events, nil
}

// queryAndStoreEvents continuously fetches logs from the Ethereum blockchain and processes them.
func (s *SidecarImpl) queryAndStoreEvents(ctx context.Context) {

	s.lastQueryBlock = s.startQueryBlock

	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		s.logger.Debug("Processing block", zap.Int64("block", s.lastQueryBlock.Int64()))
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.IsRunning() {
				return
			}

			s.fetchAndProcessLogs(ctx)
		}
	}
}

// fetchAndProcessLogs fetches the logs from the blockchain and processes them.
func (s *SidecarImpl) fetchAndProcessLogs(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	logs, err := s.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: s.lastQueryBlock,
		Addresses: []common.Address{s.contractAddress},
	})
	if err != nil {
		s.logger.Error("Error fetching logs", zap.Error(err))
		return
	}

	if len(logs) == 0 {
		return
	}

	if s.processLogs(logs) {
		// Update the last queried block to the block number of the last log + 1
		lastLogBlock := logs[len(logs)-1].BlockNumber
		s.lastQueryBlock = big.NewInt(0).SetUint64(lastLogBlock + 1)
	}
}

// processLogs processes each log in a sequential order stores it.
func (s *SidecarImpl) processLogs(logs []types.Log) bool {
	// Temporary structure to hold events per block
	tempBlocks := make(map[uint64][]sidecartypes.Event)

	lastBlockNumber := s.startQueryBlock.Uint64()
	lastTxIndex := int(-1)
	lastLogIndex := int(-1)

	for _, vLog := range logs {
		currentBlockNumber := vLog.BlockNumber

		if !s.isLogSequential(vLog, &lastBlockNumber, &lastTxIndex, &lastLogIndex) {
			return false
		}

		event, err := processLog(vLog, s.contractABI)
		if err != nil {
			s.logger.Error("Error processing log", zap.Error(err))
			return false
		}

		// If the event is nil it means we've processed an unknown event and we can skip it.
		if event == nil {
			s.logger.Debug("Processed unknown event, skipping it.", zap.Int64("block", int64(vLog.BlockNumber)))
			continue
		}

		// Add the event to the temporary block map
		tempBlocks[currentBlockNumber] = append(tempBlocks[currentBlockNumber], *event)
	}

	// All logs are sequential; move them from temporary to permanent storage
	for blockNum, events := range tempBlocks {
		blockNumStr := strconv.FormatUint(blockNum, 10)
		s.blocksMap[blockNumStr] = &EthereumBlock{
			BlockNumber: new(big.Int).SetUint64(blockNum),
			Events:      events,
		}
	}

	return true
}

// isLogSequential checks if the log is sequential based on TxIndex and LogIndex.
func (s *SidecarImpl) isLogSequential(vLog types.Log, lastBlockNumber *uint64, lastTxIndex, lastLogIndex *int) bool {
	currentBlockNumber := vLog.BlockNumber
	currentTxIndex := int(vLog.TxIndex)
	currentLogIndex := int(vLog.Index)

	// Verify that the block number is sequential.
	if currentBlockNumber != *lastBlockNumber {

		if currentBlockNumber < *lastBlockNumber {
			s.logger.Error(
				"Block is not sequential",
				zap.Uint64("currentBlockNumber", currentBlockNumber),
				zap.Uint64("lastBlockNumber", *lastBlockNumber),
			)
			return false
		}

		// Reset the indices for the new block
		*lastTxIndex = -1
		*lastLogIndex = -1
	}

	// For logs within the same block, check transaction and log indices
	if currentTxIndex <= *lastTxIndex || currentLogIndex <= *lastLogIndex {
		s.logger.Error(
			"Log is not sequential within the block",
			zap.Uint64("currentBlockNumber", currentBlockNumber),
			zap.Int("currentTxIndex", currentTxIndex),
			zap.Int("lastTxIndex", *lastTxIndex),
			zap.Int("currentLogIndex", currentLogIndex),
			zap.Int("lastLogIndex", *lastLogIndex),
		)
		return false
	}

	s.logger.Debug(
		"Processed log in order.",
		zap.Uint64("currentBlockNumber", currentBlockNumber),
		zap.Uint64("lastBlockNumber", *lastBlockNumber),
		zap.Int("currentTxIndex", currentTxIndex),
		zap.Int("lastTxIndex", *lastTxIndex),
		zap.Int("currentLogIndex", currentLogIndex),
		zap.Int("lastLogIndex", *lastLogIndex),
	)

	// Update the tracking variables
	*lastBlockNumber = currentBlockNumber
	*lastTxIndex = currentTxIndex
	*lastLogIndex = currentLogIndex

	return true
}
