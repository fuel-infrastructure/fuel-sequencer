package ethwrappedclient

import (
	"context"

	"github.com/ethereum/go-ethereum"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

// EthWsClient extends the EthWrappedClient with websocket functionality.
type EthWsClient struct {
	// The base struct.
	*EthWrappedClient
}

// NewEthWsClient creates a new NewEthWsClient instance.
func NewEthWsClient(logger *zap.Logger, ethClient *ethclient.Client) *EthWsClient {
	return &EthWsClient{
		EthWrappedClient: &EthWrappedClient{
			logger:    logger,
			ethClient: ethClient,
		},
	}
}

// SubscribeNewHead wraps the SubscribeNewHead call to the Ethereum client.
func (ec *EthWsClient) SubscribeNewHead(
	ctx context.Context, ch chan<- *ethereumtypes.Header,
) (ethereum.Subscription, error) {
	return ec.ethClient.SubscribeNewHead(ctx, ch)
}
