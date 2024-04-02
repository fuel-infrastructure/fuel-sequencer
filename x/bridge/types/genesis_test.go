package types_test

import (
	"testing"

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
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				SupplyDeltaInfo: &types.SupplyDeltaInfo{
					LastSupply: math.NewInt(99),
					Delta:      math.NewInt(34),
					Offset:     math.NewInt(123),
				},
				LastEthereumNonce:        math.NewInt(3),
				LastEthereumBlockSynced:  math.NewInt(1),
				EthereumEventIndexOffset: math.NewInt(2),
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
