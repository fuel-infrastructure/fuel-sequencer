package types_test

import (
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	"github.com/stretchr/testify/require"
)

func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		// Default is not valid due to vestingStartTime
		{
			desc:     "default is not valid",
			genState: types.DefaultGenesis(),
			valid:    false,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.NewParams(
					types.DefaultBridgeDenom,
					types.DefaultEthereumProxyContractAddress,
					types.DefaultAuthorizeMessagesAllowed,
					types.DefaultSupplyDeltaPeriod,
					testutiltypes.TestVestingStartingTime,
					nil,
					types.DefaultMaxEthBlockUpdateDelay,
					types.DefaultInjectedEventTxMaxBytes,
					types.DefaultSequencerTxsAllocation,
					types.DefaultMaxAuthorizeMessages,
				),
				SupplyDeltaInfo: &types.SupplyDeltaInfo{
					LastSupply: math.NewInt(99),
					Offset:     math.NewInt(123),
					ToReport:   math.NewInt(34),
				},
				LastEthereumNonce:        math.NewInt(3),
				LastEthereumBlockSynced:  1,
				EthereumEventIndexOffset: 2,
				LastEthBlockUpdateTime:   time.Now(),
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
