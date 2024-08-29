package ethwrappedclient

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// EthWrappedClient wraps the standard Ethereum client library. This only acts as a base struct for EthWsClient and
// EthRpcClient.
type EthWrappedClient struct {
	logger *zap.Logger

	// ethClient is the direct ethereum client we'll be interacting with.
	//
	// NOTE:
	// 1. An http URL should be dialed if we'll use RPC calls.
	// 2. A websocket URL should be dialed if we'll use subscriptions.
	ethClient *ethclient.Client

	// metrics is the set of all Prometheus metrics exposed by EthWrappedClient.
	metrics *Metrics
}

// Close closes the underlying RPC connection.
func (ec *EthWrappedClient) Close() {
	ec.ethClient.Close()
}
