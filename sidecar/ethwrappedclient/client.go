package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/utils"
	"go.uber.org/zap"

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
}

// NewClient creates a new EthWrappedClient instance.
func NewClient(
	logger *zap.Logger,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	contractAbi abi.ABI,
) *EthWrappedClient {
	return &EthWrappedClient{
		logger:          logger,
		ethClient:       ethClient,
		contractAddress: contractAddress,
		contractABI:     contractAbi,
	}
}

// FilterLogs wraps the FilterLogs call to the Ethereum client.
func (ec *EthWrappedClient) FilterLogs(ctx context.Context, fromBlock, toBlock *big.Int) ([]ethereumtypes.Log, error) {

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

// ProcessLog processes a new log from Ethereum.
func (ec *EthWrappedClient) ProcessLog(log ethereumtypes.Log) (event *sidecartypes.Event, err error) {

	if log.Removed {
		ec.logger.Debug("skipping removed log", zap.Uint64("block", log.BlockNumber))
		return
	}

	// TODO: can we verify log order?

	event, err = utils.ExtractLogDataToEvent(log, ec.contractABI)
	if err != nil {
		ec.logger.Error("error processing log", zap.Error(err))
		return
	}

	// If the event is nil it means we've processed an unrecognized event, and we can skip it.
	if event == nil {
		ec.logger.Debug("skipping unrecognized event", zap.Uint64("block", log.BlockNumber))
		return
	}

	ec.logger.Debug("processed a log successfully",
		zap.Uint64("block", log.BlockNumber),
		zap.Uint("tx_index", log.TxIndex),
		zap.Uint("index", log.Index),
	)
	return
}
