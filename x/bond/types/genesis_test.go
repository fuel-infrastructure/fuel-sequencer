package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := types.DefaultGenesis()
	require.NotNil(t, genesis)
	require.Equal(t, "", genesis.Params.YieldRecipient)
	require.Equal(t, (*time.Time)(nil), genesis.Params.YieldTime)
	require.Equal(t, sdkmath.ZeroInt(), genesis.Params.YieldAmount)
	require.Equal(t, int64(0), genesis.State.YieldMintHeight)
}

func TestGenesisState_Validate(t *testing.T) {
	baseTime := time.Now()
	future := baseTime.Add(time.Hour)
	recipient := "cosmos124maqmcqv8tquy764ktz7cu0gxnzfw54k9cmz5"

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
					recipient,
					&future,
					sdkmath.NewInt(1000000),
				),
				State: types.DefaultState(),
			},
			expErr: false,
		},
		{
			name: "invalid genesis - invalid recipient",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					"invalid",
					&future,
					sdkmath.NewInt(1000000),
				),
				State: types.DefaultState(),
			},
			expErr: true,
		},
		{
			name: "invalid genesis - negative amount",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					recipient,
					&future,
					sdkmath.NewInt(-1),
				),
				State: types.DefaultState(),
			},
			expErr: true,
		},
		{
			name: "invalid genesis - past time",
			genesis: &types.GenesisState{
				Params: types.NewParams(
					recipient,
					&baseTime,
					sdkmath.NewInt(1000000),
				),
				State: types.DefaultState(),
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
