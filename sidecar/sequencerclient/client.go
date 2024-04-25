package sequencerclient

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/grpc"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SequencerClient wraps the sequencer bridge client, with extra functionality.
type SequencerClient struct {
	// client is the sequencer client.
	bridgeQueryClient bridgetypes.QueryClient
}

// NewClient creates a new Client instance.
func NewClient(
	grpcConn *grpc.ClientConn,
) *SequencerClient {

	return &SequencerClient{
		bridgeQueryClient: bridgetypes.NewQueryClient(grpcConn),
	}
}

// FetchLastEthereumBlockSynced fetches the last ethereum synced block from the bridge module.
func (sc *SequencerClient) FetchLastEthereumBlockSynced(ctx context.Context) (*big.Int, error) {
	resp, err := sc.bridgeQueryClient.LastEthereumBlockSynced(
		ctx,
		&bridgetypes.QueryGetLastEthereumBlockSyncedRequest{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch LastEthereumBlockSynced from Sequencer: %v", err)
	}

	lastSyncedBlock := new(big.Int)
	lastSyncedBlock.SetUint64(resp.Block)

	return lastSyncedBlock, nil
}
