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
	client                *ethclient.Client
	contractAddress       common.Address
	contractABI           abi.ABI
	startBlock            *big.Int
	blocksMap             map[string]*EthereumBlock
	updateInterval        time.Duration
	latestBlockWithEvents *big.Int

	// running is the current status of the main sidecar process (running or not).
	running atomic.Bool
}

// NewSidecar creates a new Sidecar instance.
func NewSidecar(
	client *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
	startBlock *big.Int,
	logger *zap.Logger,
) *SidecarImpl {
	return &SidecarImpl{
		logger:          logger,
		client:          client,
		contractAddress: contractAddress,
		contractABI:     contractAbi,
		startBlock:      startBlock,
		blocksMap:       make(map[string]*EthereumBlock),
		updateInterval:  10 * time.Second,
	}
}

// Start begins the process of querying and storing events from the Ethereum blockchain.
func (s *SidecarImpl) Start(ctx context.Context) error {
	s.logger.Info("starting sidecar")

	// Initial check to verify Ethereum client connectivity and log fetching capability
	query := ethereum.FilterQuery{
		FromBlock: s.startBlock,
		ToBlock:   s.startBlock,
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

		// If the queried block is within the range of processed blocks but not found,
		// it means there were no events for this block, hence return an empty list.
		if s.latestBlockWithEvents != nil && blockNumber.Cmp(s.latestBlockWithEvents) <= 0 {
			return []sidecartypes.Event{}, nil
		}

		// Otherwise this block was not yet processed
		return nil, fmt.Errorf("no events found for block number %s", blockNumber)
	}

	return block.Events, nil
}

// queryAndStoreEvents continuously fetches logs from the Ethereum blockchain and processes them.
func (s *SidecarImpl) queryAndStoreEvents(ctx context.Context) {
	lastQueriedBlock := s.startBlock

	ticker := time.NewTicker(s.updateInterval)
	defer ticker.Stop()

	for {
		s.logger.Debug("Processing block", zap.Int64("block", lastQueriedBlock.Int64()))
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !s.IsRunning() {
				return
			}

			s.fetchAndProcessLogs(ctx, &lastQueriedBlock)
		}
	}
}

// fetchAndProcessLogs fetches the logs from the blockchain and processes them.
func (s *SidecarImpl) fetchAndProcessLogs(ctx context.Context, lastQueriedBlock **big.Int) {
	logs, err := s.client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: *lastQueriedBlock,
		Addresses: []common.Address{s.contractAddress},
	})
	if err != nil {
		s.logger.Error("Error fetching logs", zap.Error(err))
		return
	}

	if len(logs) == 0 {
		return
	}

	success := s.processLogs(logs)
	if !success {
		return
	}

	// Update the last queried block to the block number of the last log + 1
	lastLogBlock := logs[len(logs)-1].BlockNumber
	*lastQueriedBlock = big.NewInt(0).SetUint64(lastLogBlock + 1)
}

// processLogs processes each log in a sequential order stores it.
func (s *SidecarImpl) processLogs(logs []types.Log) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	lastTxIndex := -1
	lastLogIndex := -1

	for _, vLog := range logs {
		if !s.isLogSequential(vLog, &lastTxIndex, &lastLogIndex) {
			return false
		}

		s.storeLog(vLog)
	}
	return true
}

// isLogSequential checks if the log is sequential based on TxIndex and LogIndex.
func (s *SidecarImpl) isLogSequential(vLog types.Log, lastTxIndex, lastLogIndex *int) bool {
	currentTxIndex := int(vLog.TxIndex)
	currentLogIndex := int(vLog.Index)

	if currentTxIndex <= *lastTxIndex || currentLogIndex <= *lastLogIndex {
		s.logger.Error(
			"Log is not sequential",
			zap.Int("currentTxIndex", currentTxIndex),
			zap.Int("lastTxIndex", *lastTxIndex),
			zap.Int("currentLogIndex", currentLogIndex),
			zap.Int("lastLogIndex", *lastLogIndex),
		)
		return false
	}

	*lastTxIndex = currentTxIndex
	*lastLogIndex = currentLogIndex
	return true
}

// storeLog processes and stores a single log.
func (s *SidecarImpl) storeLog(vLog types.Log) {
	blockNumStr := strconv.FormatUint(vLog.BlockNumber, 10)
	if _, exists := s.blocksMap[blockNumStr]; !exists {
		s.blocksMap[blockNumStr] = &EthereumBlock{
			BlockNumber: new(big.Int).SetUint64(vLog.BlockNumber),
			Events:      make([]sidecartypes.Event, 0),
		}
	}

	event, err := processLog(vLog, s.contractABI)
	if err != nil {
		s.logger.Error("Error processing log", zap.Error(err))
		return
	}

	s.blocksMap[blockNumStr].Events = append(s.blocksMap[blockNumStr].Events, event)
}
