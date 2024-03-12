package sidecar

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var _ Sidecar = (*SidecarImpl)(nil)

// Sidecar defines the expected interface for a sidecar. It is consumed by the sidecar server.
type Sidecar interface {
	IsRunning() bool
	QueryBlockEvents(ctx context.Context, blockNumber *big.Int) ([]sidecartypes.Event, error)
	Start(ctx context.Context) error
	Stop()
}

// SidecarImpl is the core component responsible for fetching ethereum blocks and processing events.
type SidecarImpl struct {
	// --------------------- General Config --------------------- //
	logger *zap.Logger
	mu     sync.Mutex

	// --------------------- Ethereum Config --------------------- //
	ethClient       *ethclient.Client
	contractAddress common.Address
	contractABI     abi.ABI
	blocksMap       map[string]*sidecartypes.EthereumBlock
	updateInterval  time.Duration

	// startQueryBlock is the block at which we started querying for events.
	startQueryBlock *big.Int
	// lastQueryBlock is last block we queried for events.
	lastQueryBlock *big.Int

	// --------------------- Cosmos Config ------------------------ //
	bridgeQueryClient bridgetypes.QueryClient

	// running is the current status of the main sidecar process (running or not).
	running atomic.Bool
}

// NewSidecar creates a new Sidecar instance.
func NewSidecar(
	ethClient *ethclient.Client,
	bridgeQueryClient bridgetypes.QueryClient,
	contractAddress common.Address,
	contractAbi abi.ABI,
	startQueryBlock *big.Int,
	logger *zap.Logger,
) *SidecarImpl {
	return &SidecarImpl{
		logger:            logger,
		ethClient:         ethClient,
		bridgeQueryClient: bridgeQueryClient,
		contractAddress:   contractAddress,
		contractABI:       contractAbi,
		startQueryBlock:   startQueryBlock,
		blocksMap:         make(map[string]*sidecartypes.EthereumBlock),
		updateInterval:    10 * time.Second,
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
	_, err := s.ethClient.FilterLogs(ctx, query)
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
func (s *SidecarImpl) QueryBlockEvents(ctx context.Context, blockNumber *big.Int) ([]sidecartypes.Event, error) {
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

		syncProgress, err := s.ethClient.SyncProgress(ctx)
		if err != nil {
			return nil, errors.New("could not get syncing status from Ethereum node")
		}

		ethHeight, err := s.ethClient.BlockNumber(ctx)
		if err != nil {
			return nil, errors.New("could not get latest height from Ethereum node")
		}

		// If the sidecar is synced with Ethereum, and it processed the current Ethereum height already, then it must
		// be the height being queried does not exist yet
		isNodeSynced := syncProgress == nil
		sidecarSyncedWithEthereum := s.lastQueryBlock.Cmp(new(big.Int).SetUint64(ethHeight)) == 0
		if isNodeSynced && sidecarSyncedWithEthereum {
			return nil, fmt.Errorf("%s %s", sidecartypes.ErrBlockDoesNotExist, blockNumber)
		}

		// Otherwise this block was not yet processed
		return nil, fmt.Errorf("block not yet processed %s", blockNumber)
	}

	return block.Events, nil
}

// fetchLastSyncedEthereumBlock fetches the last block that was processed on the sequencer chain.
func (s *SidecarImpl) fetchLastSyncedEthereumBlock(ctx context.Context) (*big.Int, error) {
	resp, err := s.bridgeQueryClient.LastEthereumBlockSynced(ctx, &bridgetypes.QueryGetLastEthereumBlockSyncedRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch last synced Ethereum block from fuel sequencer: %v", err)
	}

	lastSyncedBlock := new(big.Int)
	lastSyncedBlock.SetString(resp.Block, 10)

	return lastSyncedBlock, nil
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

			// Fetch the last synced Ethereum block before querying for new logs
			lastSyncedBlock, err := s.fetchLastSyncedEthereumBlock(ctx)
			s.logger.Debug("Querying LastSyncedEthereumBlock: ", zap.String("last synced block", lastSyncedBlock.String()))
			if err != nil {
				// Log the error if the last synced Ethereum block is not obtained.
				// Note; We should still attempt to process Ethereum blocks. Reason being is that if the processing
				// is skipped the Sequencer will not be able to produce the first block and the sidecar would not be
				// able to query the Sequencer, causing a deadlock.
				s.logger.Error("Error: ", zap.Error(err))
			}

			// Set the last queried block to the last synced block. If the value could not be obtained from the chain
			// it will default to s.startQueryBlock set in the beginning of this function.
			if lastSyncedBlock != nil && lastSyncedBlock.Cmp(s.lastQueryBlock) > 0 {
				s.lastQueryBlock = lastSyncedBlock
			}

			s.fetchAndProcessLogs(ctx)
		}
	}
}

// fetchAndProcessLogs fetches the logs from the blockchain and processes them.
func (s *SidecarImpl) fetchAndProcessLogs(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Determine the range of blocks to query.
	currentBlockNumber, err := s.ethClient.BlockNumber(ctx)
	if err != nil {
		s.logger.Error("Error fetching current block number", zap.Error(err))
		return
	}

	// Return if there is no update for the ETH block height.
	if currentBlockNumber <= s.lastQueryBlock.Uint64() {
		s.logger.Debug(
			"Current block number already processed.",
			zap.String("current block number", s.lastQueryBlock.String()),
		)
		return
	}

	logs, err := s.ethClient.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: s.lastQueryBlock,
		ToBlock:   new(big.Int).SetUint64(currentBlockNumber),
		Addresses: []common.Address{s.contractAddress},
	})
	if err != nil {
		s.logger.Error("Error fetching logs", zap.Error(err))
		return
	}

	// Process the logs if any are found.
	if len(logs) > 0 {
		if err := s.processLogs(logs); err != nil {
			s.logger.Error("Failed to process logs", zap.Error(err))
			return
		}
	}

	// Regardless of whether logs were found, update the last queried block to the current block number,
	// since we have now queried up to this block.
	s.lastQueryBlock = new(big.Int).SetUint64(currentBlockNumber)
}

// processLogs processes each log in a sequential order and stores it.
func (s *SidecarImpl) processLogs(logs []types.Log) error {
	// Temporary structure to hold events per block
	tempBlocks := make(map[uint64][]sidecartypes.Event)

	lastBlockNumber := s.startQueryBlock.Uint64()
	lastTxIndex := int(-1)
	lastLogIndex := int(-1)

	for _, vLog := range logs {
		currentBlockNumber := vLog.BlockNumber

		if vLog.Removed {
			s.logger.Debug("Processed a removed log, skipping it.", zap.Int64("block", int64(vLog.BlockNumber)))
			continue
		}

		if err := s.validateIsLogSequential(vLog, &lastBlockNumber, &lastTxIndex, &lastLogIndex); err != nil {
			s.logger.Error("Failed sequential validation", zap.Error(err))
			return err
		}

		event, err := processLog(vLog, s.contractABI)
		if err != nil {
			s.logger.Error("Error processing log", zap.Error(err))
			return fmt.Errorf("error processing log %s", err)
		}

		// If the event is nil it means we've processed an unknown event and we can skip it.
		if event == nil {
			s.logger.Debug("Processed unknown event, skipping it.", zap.Int64("block", int64(vLog.BlockNumber)))
			continue
		}

		// Add the event to the temporary block map
		tempBlocks[currentBlockNumber] = append(tempBlocks[currentBlockNumber], *event)
		s.logger.Debug("Processed a log successfully.", zap.Int64("block", int64(vLog.BlockNumber)))
	}

	// All logs are sequential; move them from temporary to permanent storage
	for blockNum, events := range tempBlocks {
		blockNumStr := strconv.FormatUint(blockNum, 10)
		s.blocksMap[blockNumStr] = &sidecartypes.EthereumBlock{
			BlockNumber: new(big.Int).SetUint64(blockNum),
			Events:      events,
		}
	}

	return nil
}

// validateIsLogSequential checks if the log is sequential based on TxIndex and LogIndex.
func (s *SidecarImpl) validateIsLogSequential(vLog types.Log, lastBlockNumber *uint64, lastTxIndex, lastLogIndex *int) error {
	currentBlockNumber := vLog.BlockNumber
	currentTxIndex := int(vLog.TxIndex)
	currentLogIndex := int(vLog.Index)

	// Initial verification to ascertain that the current block's number sequentially follows the last processed block's number.
	if currentBlockNumber != *lastBlockNumber {
		if currentBlockNumber < *lastBlockNumber {
			return fmt.Errorf(
				"non-sequential block detected: current block number %d precedes last processed block number %d",
				currentBlockNumber, *lastBlockNumber,
			)
		}

		// Resetting indices for the new block, acknowledging the transition to a subsequent block in the sequence.
		*lastTxIndex = -1
		*lastLogIndex = -1
	}

	// Ensuring within-block log sequentiality by comparing the current log's indices against the last processed log's indices.
	if currentTxIndex <= *lastTxIndex || currentLogIndex <= *lastLogIndex {
		return fmt.Errorf(
			"log sequentiality violation within block %d: currentTxIndex=%d, lastTxIndex=%d, currentLogIndex=%d, lastLogIndex=%d",
			currentBlockNumber, currentTxIndex, *lastTxIndex, currentLogIndex, *lastLogIndex,
		)
	}

	// Upon successful validation, updating tracking variables to reflect the most recent log's indices.
	*lastBlockNumber = currentBlockNumber
	*lastTxIndex = currentTxIndex
	*lastLogIndex = currentLogIndex

	return nil
}
