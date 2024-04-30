package ethclient

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/utils"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// EthWrappedClient wraps the Ethereum client, with extra functionality.
type EthWrappedClient struct {
	logger *zap.Logger

	// ethClient is the direct ethereum client, we'll be interacting with.
	ethClient *ethclient.Client
	// contractAddress is the address of the contract we'll be querying.
	contractAddress common.Address
	// contractABI is the contract ABI of the contract address we're querying.
	contractABI abi.ABI

	// logsQueryRateLimiter limits how many queries for logs we can perform in a time interval.
	logsQueryRateLimiter *rate.Limiter
}

// NewClient creates a new EthWrappedClient instance.
func NewClient(
	logger *zap.Logger,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
	logsQueryInterval time.Duration,
) *EthWrappedClient {
	return &EthWrappedClient{
		logger:               logger,
		ethClient:            ethClient,
		contractAddress:      contractAddress,
		contractABI:          contractAbi,
		logsQueryRateLimiter: rate.NewLimiter(rate.Every(logsQueryInterval), 1), // max 1 request per interval
	}
}

// FilterLogs wraps and rate-limits the FilterLogs call to the Ethereum client.
func (ec *EthWrappedClient) FilterLogs(ctx context.Context, fromBlock, toBlock *big.Int) ([]ethereumtypes.Log, error) {

	// Wait for the rate limiter to let us through.
	if !ec.logsQueryRateLimiter.Allow() {
		ec.logger.Debug("waiting for logs rate limiter", zap.Float64("limit", float64(ec.logsQueryRateLimiter.Limit())))
		if err := ec.logsQueryRateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("error when waiting for logs rate limiter: %s", err.Error())
		}
	}

	// Create the ethereum query for the set contract.
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{ec.contractAddress},
	}

	return ec.ethClient.FilterLogs(ctx, query)
}

// SubscribeNewHead wraps the SubscribeNewHead call to the Ethereum client.
func (ec *EthWrappedClient) SubscribeNewHead(
	ctx context.Context, ch chan<- *ethereumtypes.Header,
) (ethereum.Subscription, error) {
	return ec.ethClient.SubscribeNewHead(ctx, ch)
}

// BlockNumber wraps the BlockNumber call to get the current block number.
func (ec *EthWrappedClient) BlockNumber(ctx context.Context) (uint64, error) {
	return ec.ethClient.BlockNumber(ctx)
}

// SyncProgress checks if the Ethereum node is synced.
func (ec *EthWrappedClient) SyncProgress(ctx context.Context) (*ethereum.SyncProgress, error) {
	return ec.ethClient.SyncProgress(ctx)
}

// PeerCount checks if the peer count of the Ethereum node.
func (ec *EthWrappedClient) PeerCount(ctx context.Context) (uint64, error) {
	return ec.ethClient.PeerCount(ctx)
}

// FetchAndProcessLogs fetches the logs from the blockchain and processes them.
// It returns the events and the last block that it synced up to.
// Upon failure, no events and no last block synced are returned.
func (ec *EthWrappedClient) FetchAndProcessLogs(
	ctx context.Context,
	fromBlock *big.Int,
	toBlock *big.Int,
	maxQueryRange *big.Int,
) (events map[uint64][]sidecartypes.Event, lastSyncedBlock *big.Int, err error) {

	// Cap the range of blocks to query. We subtract 1 from maxQueryRange since otherwise we query an extra block.
	// Example: if fromBlock is 100 and maxQueryRange is 1, then the fromBlock and toBlock should both be 100.
	maxToBlock := new(big.Int).Add(fromBlock, new(big.Int).Sub(maxQueryRange, big.NewInt(1)))
	if toBlock.Cmp(maxToBlock) > 0 {
		toBlock = maxToBlock
	}

	// Filter the logs from the next query block to the to block (note: this is rate-limited under the hood).
	logs, err := ec.FilterLogs(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("logs query failed: %s", err.Error())
	}

	// Process the logs if any are found.
	eventsMap, err := ec.processLogs(logs, fromBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("logs processing failed: %s", err.Error())
	}

	// Regardless of whether logs were found, update the last queried block to the current block number,
	// since we have now queried up to this block.
	ec.logger.Info("processed logs from range",
		zap.String("from_block", fromBlock.String()),
		zap.String("to_block", toBlock.String()),
		zap.Int("num_events", len(logs)),
	)

	// Return the fetched events and the block that we fetched up to.
	return eventsMap, toBlock, nil
}

// processLogs processes each log in a sequential order and stores it.
func (ec *EthWrappedClient) processLogs(
	logs []ethereumtypes.Log,
	nextQueryBlock *big.Int,
) (map[uint64][]sidecartypes.Event, error) {

	// If there are no logs to process return.
	if len(logs) == 0 {
		return nil, nil
	}

	// Temporary structure to hold events per block
	tempBlocks := make(map[uint64][]sidecartypes.Event)

	lastBlockNumber := nextQueryBlock.Uint64()
	lastTxIndex := int(-1)
	lastLogIndex := int(-1)

	for _, vLog := range logs {
		currentBlockNumber := vLog.BlockNumber

		if vLog.Removed {
			ec.logger.Debug("skipping removed log", zap.Uint64("block", vLog.BlockNumber))
			continue
		}

		if err := utils.ValidateIsLogSequential(vLog, &lastBlockNumber, &lastTxIndex, &lastLogIndex); err != nil {
			ec.logger.Error("failed sequential validation", zap.Error(err))
			return nil, err
		}

		event, err := utils.ExtractLogDataToEvent(vLog, ec.contractABI)
		if err != nil {
			ec.logger.Error("error processing log", zap.Error(err))
			return nil, fmt.Errorf("error processing log %s", err)
		}

		// If the event is nil it means we've processed an unrecognized event, and we can skip it.
		if event == nil {
			ec.logger.Debug("skipping unrecognized event", zap.Uint64("block", vLog.BlockNumber))
			continue
		}

		// Add the event to the temporary block map
		tempBlocks[currentBlockNumber] = append(tempBlocks[currentBlockNumber], *event)
		ec.logger.Debug("processed log from block",
			zap.Uint64("block", vLog.BlockNumber),
			zap.Uint("tx_index", vLog.TxIndex),
			zap.Uint("index", vLog.Index),
		)
	}

	return tempBlocks, nil
}
