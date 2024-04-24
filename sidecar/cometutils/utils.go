package comet_utils

import (
	"context"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	sidecarutils "github.com/fuel-infrastructure/fuel-sequencer/sidecar/utils"
)

// QueryCometGenesisForLastEthereumBlockSynced queries the comet genesis through an rpc client and parses it to extract
// the last ethereum block synced.
func QueryCometGenesisForLastEthereumBlockSynced(ctx context.Context, tendermintNodeRPC string) (uint64, error) {

	httpClient, err := libclient.DefaultHTTPClient(tendermintNodeRPC)
	if err != nil {
		return 0, err
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(tendermintNodeRPC, "/websocket", httpClient)
	if err != nil {
		return 0, err
	}

	genesis, err := rpcClient.Genesis(ctx)
	if err != nil {
		return 0, err
	}

	genbz, err := genesis.Genesis.AppState.MarshalJSON()
	if err != nil {
		return 0, err
	}

	lastEthereumBlockSyncedUint := sidecarutils.MustGetLastEthereumBlockSyncedFromGenesis(genbz)

	return lastEthereumBlockSyncedUint, nil
}
