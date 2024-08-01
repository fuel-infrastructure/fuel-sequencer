package comet_utils_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cmtcoretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
	cmtutils "github.com/fuel-infrastructure/fuel-sequencer/sidecar/cometutils"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil/servers"
	"github.com/stretchr/testify/require"
)

func TestQuerySequencerGenesisForLastEthereumBlockSynced(t *testing.T) {
	testCases := []struct {
		name                       string
		mockGenesis                *cmtcoretypes.ResultGenesis
		expLastEthereumBlockSynced uint64
		setBadServerUrl            bool
		expErrMsg                  string
	}{
		{
			name: "Returns LastEthereumBlockSynced if query successful",
			mockGenesis: &cmtcoretypes.ResultGenesis{
				Genesis: &cmttypes.GenesisDoc{
					AppState:      json.RawMessage(`{"bridge": {"last_ethereum_block_synced": "12345"}}`),
					InitialHeight: int64(1),
				},
			},
			expLastEthereumBlockSynced: 12345,
		},
		{
			name: "Returns error and LastEthereumBlockSynced 0 if could not connect with server",
			mockGenesis: &cmtcoretypes.ResultGenesis{
				Genesis: &cmttypes.GenesisDoc{
					AppState:      json.RawMessage(`{"bridge": {"last_ethereum_block_synced": "12345"}}`),
					InitialHeight: int64(1),
				},
			},
			setBadServerUrl: true,
			expErrMsg:       "connection refused",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a MockTendermintServer for the CometBFT client to connect with.
			mockTendermintServer := servers.NewMockTendermintServer()
			serverURL := mockTendermintServer.Start()
			defer mockTendermintServer.Stop()

			// Set bad server URL if dictated by test
			if tc.setBadServerUrl {
				serverURL = "http://localhost:0" // The OS will never assign port 0 to a server
			}

			// Set mock value
			mockTendermintServer.SetMockGenesis(tc.mockGenesis)

			// Create context
			timeout := 10 * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Perform query.
			actualLastEthereumBlockSynced, err := cmtutils.QuerySequencerGenesisForLastEthereumBlockSynced(
				ctx, serverURL,
			)

			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				require.Equal(t, uint64(0), actualLastEthereumBlockSynced)
				return
			}
			require.NoError(t, err)

			// Confirm that the result is as expected.
			require.Equal(t, tc.expLastEthereumBlockSynced, actualLastEthereumBlockSynced)
		})
	}
}
