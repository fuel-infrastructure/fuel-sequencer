package ethclient

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/utils"
	"go.uber.org/zap"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// EthWrappedClient wraps the Ethereum client, with extra functionality.
type EthWrappedClient struct {

	// ethClient is the direct ethereum client, we'll be interacting with.
	ethClient *ethclient.Client
	// contractAddress is the address of the contract we'll be querying.
	contractAddress common.Address
	// contractABI is the contract ABI of the contract address we're querying.
	contractABI abi.ABI
}

// NewClient creates a new EthWrappedClient instance.
func NewClient(
	ethClient *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
) *EthWrappedClient {
	return &EthWrappedClient{
		ethClient:       ethClient,
		contractAddress: contractAddress,
		contractABI:     contractAbi,
	}
}

// FilterLogs wraps the FilterLogs call to the Ethereum client.
func (ec *EthWrappedClient) FilterLogs(ctx context.Context, fromBlock, toBlock *big.Int) ([]types.Log, error) {

	// Create the ethereum query for the set contract.
	query := ethereum.FilterQuery{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Addresses: []common.Address{ec.contractAddress},
	}

	return ec.ethClient.FilterLogs(ctx, query)
}

// BlockNumber wraps the BlockNumber call to get the current block number.
func (ec *EthWrappedClient) BlockNumber(ctx context.Context) (uint64, error) {
	return ec.ethClient.BlockNumber(ctx)
}

// CheckEthereumNodeSync checks if the Ethereum node is synced.
func (ec *EthWrappedClient) CheckEthereumNodeSync(ctx context.Context) (bool, error) {
	progress, err := ec.ethClient.SyncProgress(ctx)
	if err != nil {
		return false, errors.New("could not get syncing status from Ethereum node")
	}

	// nil progress means it's synced
	return progress == nil, nil
}

// FetchAndProcessLogs fetches the logs from the blockchain and processes them.
func (ec *EthWrappedClient) FetchAndProcessLogs(
	ctx context.Context,
	logger *zap.Logger,
	nextQueryBlock, maxQueryRange *big.Int,
) (*map[uint64][]sidecartypes.Event, *big.Int) {

	// Determine the range of blocks to query.
	currentBlockNumber, err := ec.ethClient.BlockNumber(ctx)
	if err != nil {
		logger.Error("Error fetching current Ethereum block number", zap.Error(err))
		return nil, nextQueryBlock
	}

	// Return if there is no update for the ETH block height.
	if currentBlockNumber < nextQueryBlock.Uint64() {
		logger.Info("No new blocks", zap.Uint64("current_block_number", currentBlockNumber))
		return nil, nextQueryBlock
	}

	// Cap the range of blocks to query. We subtract 1 from maxQueryRange since otherwise we query an extra block.
	// Example: if nextQueryBlock is 100 and maxQueryRange is 1, then the fromBlock and toBlock should both be 100.
	toBlock := new(big.Int).SetUint64(currentBlockNumber)
	maxToBlock := new(big.Int).Add(nextQueryBlock, new(big.Int).Sub(maxQueryRange, big.NewInt(1)))
	if toBlock.Cmp(maxToBlock) > 0 {
		toBlock = maxToBlock
	}

	// Filter the logs from the next query block to the to block.
	logs, err := ec.FilterLogs(ctx, nextQueryBlock, toBlock)
	if err != nil {
		logger.Error("Error fetching logs", zap.Error(err))
		return nil, nextQueryBlock
	}

	// Process the logs if any are found.
	eventsMap, err := ec.processLogs(logger, logs, nextQueryBlock)
	if err != nil {
		logger.Error("Failed to process logs", zap.Error(err))
		return nil, nextQueryBlock
	}

	// Regardless of whether logs were found, update the last queried block to the current block number,
	// since we have now queried up to this block.
	logger.Info("Processed logs from range of blocks",
		zap.String("from_block", nextQueryBlock.String()),
		zap.String("to_block", toBlock.String()),
		zap.Int("no_of_events", len(logs)),
	)

	// Next block to query will be the one after toBlock
	return eventsMap, new(big.Int).Add(toBlock, big.NewInt(1))
}

// processLogs processes each log in a sequential order and stores it.
func (ec *EthWrappedClient) processLogs(
	logger *zap.Logger,
	logs []types.Log,
	nextQueryBlock *big.Int,
) (*map[uint64][]sidecartypes.Event, error) {

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
			logger.Debug("Processed a removed log, skipping it.", zap.Int64("block", int64(vLog.BlockNumber)))
			continue
		}

		if err := utils.ValidateIsLogSequential(vLog, &lastBlockNumber, &lastTxIndex, &lastLogIndex); err != nil {
			logger.Error("Failed sequential validation", zap.Error(err))
			return nil, err
		}

		event, err := utils.ExtractLogDataToEvent(vLog, ec.contractABI)
		if err != nil {
			logger.Error("Error processing log", zap.Error(err))
			return nil, fmt.Errorf("error processing log %s", err)
		}

		// If the event is nil it means we've processed an unknown event and we can skip it.
		if event == nil {
			logger.Debug("Processed unknown event, skipping it.", zap.Int64("block", int64(vLog.BlockNumber)))
			continue
		}

		// Add the event to the temporary block map
		tempBlocks[currentBlockNumber] = append(tempBlocks[currentBlockNumber], *event)
		logger.Debug("Processed a log successfully",
			zap.Uint64("block", vLog.BlockNumber),
			zap.Uint("tx_index", vLog.TxIndex),
			zap.Uint("index", vLog.Index),
		)
	}

	return &tempBlocks, nil
}
