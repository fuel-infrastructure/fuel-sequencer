package ethwrappedclient

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/utils"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// EthRpcClient extends the EthWrappedClient with RPC calls.
type EthRpcClient struct {
	// The base struct.
	*EthWrappedClient

	// contractAddress is the address of the contract we'll be querying.
	contractAddress common.Address
	// contractABI is the ABI of the contract we're querying.
	contractABI abi.ABI

	// logsQueryLimiter limits how many queries for logs we can perform in a time interval.
	logsQueryLimiter *rate.Limiter
}

// NewEthRpcClient creates a new EthRpcClient instance.
func NewEthRpcClient(
	logger *zap.Logger,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
	minLogsQueryInterval time.Duration,
) *EthRpcClient {
	return &EthRpcClient{
		EthWrappedClient: &EthWrappedClient{
			logger:    logger,
			ethClient: ethClient,
		},
		contractAddress:  contractAddress,
		contractABI:      contractAbi,
		logsQueryLimiter: rate.NewLimiter(rate.Every(minLogsQueryInterval), 1), // max 1 request per interval
	}
}

// BlockNumber wraps the BlockNumber call to get the current block number.
func (ec *EthRpcClient) BlockNumber(ctx context.Context) (uint64, error) {
	return ec.ethClient.BlockNumber(ctx)
}

// FinalizedBlockNumber contains logic for querying the block number of the latest finalized block.
// NOTE: This was copied from https://github.com/ethereum/go-ethereum/blob/7f131dcbc9ffe986f91a1f51025bcfdcc0aa8f0e/ethclient/ethclient.go#L128
func (ec *EthRpcClient) FinalizedBlockNumber(ctx context.Context) (*big.Int, error) {
	var raw json.RawMessage
	err := ec.ethClient.Client().CallContext(ctx, &raw, "eth_getBlockByNumber", "finalized", true)
	if err != nil {
		return nil, err
	}

	// Decode header and transactions.
	var head *ethereumtypes.Header
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, err
	}
	// When the block is not found, the API returns JSON null.
	if head == nil {
		return nil, ethereum.NotFound
	}

	return head.Number, nil
}

// FilterLogs wraps and rate-limits the FilterLogs call to the Ethereum client.
func (ec *EthRpcClient) FilterLogs(ctx context.Context, fromBlock, toBlock *big.Int) ([]ethereumtypes.Log, error) {

	// Wait for the rate limiter to let us through.
	if !ec.logsQueryLimiter.Allow() {
		maxWaitSeconds := 1 / float64(ec.logsQueryLimiter.Limit())
		ec.logger.Debug("waiting for logs rate limiter", zap.Float64("max_wait_seconds", maxWaitSeconds))
		if err := ec.logsQueryLimiter.Wait(ctx); err != nil {
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

// FetchAndProcessLogs fetches the logs from the blockchain and processes them. It returns the events and the last block
// that it synced up to. Upon failure, no events and no last block synced are returned.
func (ec *EthRpcClient) FetchAndProcessLogs(
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

	// Regardless of whether logs were found, update the last queried block to the current block number, since we have
	// now queried up to this block.
	ec.logger.Info("processed logs from range",
		zap.String("from_block", fromBlock.String()),
		zap.String("to_block", toBlock.String()),
		zap.Int("num_events", len(logs)),
	)

	// Return the fetched events and the block that we fetched up to.
	return eventsMap, toBlock, nil
}

// processLogs processes each log in a sequential order and stores it.
func (ec *EthRpcClient) processLogs(
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
	lastTxIndex := -1
	lastLogIndex := -1

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
			log, err := vLog.MarshalJSON()
			ec.logger.Debug("skipping unrecognized event",
				zap.Uint64("block", vLog.BlockNumber),
				zap.String("log", string(log)),
				zap.NamedError("json_marshal_err", err),
			)
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
