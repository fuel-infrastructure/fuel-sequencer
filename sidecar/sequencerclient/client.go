package sequencerclient

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"google.golang.org/grpc"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SequencerClient wraps the sequencer bridge client, with extra functionality.
type SequencerClient struct {
	// client is the sequencer client.
	bridgeQueryClient bridgetypes.QueryClient

	// metrics is the set of all Prometheus metrics exposed by SequencerClient.
	metrics *Metrics
}

// NewClient creates a new Client instance.
func NewClient(
	grpcConn *grpc.ClientConn,
	metrics *Metrics,
) *SequencerClient {
	return &SequencerClient{
		bridgeQueryClient: bridgetypes.NewQueryClient(grpcConn),
		metrics:           metrics,
	}
}

// FetchLastEthereumBlockSynced fetches the last ethereum synced block from the bridge module.
func (sc *SequencerClient) FetchLastEthereumBlockSynced(ctx context.Context) (*big.Int, error) {
	queriedAt := time.Now()
	resp, err := sc.bridgeQueryClient.LastEthereumBlockSynced(
		ctx,
		&bridgetypes.QueryGetLastEthereumBlockSyncedRequest{},
	)
	sc.metrics.ObserveLEBSQueryDelay(queriedAt, time.Now())
	if err != nil {
		sc.metrics.LEBSQueryErrorCount.Add(1)
		return nil, fmt.Errorf("failed to fetch LastEthereumBlockSynced from Sequencer: %v", err)
	}

	lastSyncedBlock := new(big.Int)
	lastSyncedBlock.SetUint64(resp.Block)

	return lastSyncedBlock, nil
}
