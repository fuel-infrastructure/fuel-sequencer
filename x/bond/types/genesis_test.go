package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()
	require.NotNil(t, genesis)
	require.Equal(t, sdkmath.LegacyZeroDec(), genesis.Params.Inflation)
	require.Equal(t, "", genesis.Params.Authority)
}

func TestGenesisState_Validate(t *testing.T) {
	testCases := []struct {
		name    string
		genesis *types.GenesisState
		expErr  bool
	}{
		{
			name:    "default genesis",
			genesis: types.DefaultGenesis(),
			expErr:  false,
		},
		{
			name: "valid genesis with params",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
					"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				),
			},
			expErr: false,
		},
		{
			name: "invalid genesis - negative inflation",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					sdkmath.LegacyNewDec(-1),
					"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
				),
			},
			expErr: true,
		},
		{
			name: "invalid genesis - invalid authority",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
					"invalid",
				),
			},
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesis.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
