package bridge_test

import (
	"testing"

	"cosmossdk.io/math"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	bridge "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SupplyDeltaInfo: &types.SupplyDeltaInfo{
			LastSupply: math.NewInt(99),
			Delta:      math.NewInt(87),
			Offset:     math.NewInt(123),
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.BridgeKeeper(t)
	bridge.InitGenesis(ctx, k, genesisState)
	got := bridge.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.SupplyDeltaInfo, got.SupplyDeltaInfo)
	// this line is used by starport scaffolding # genesis/test/assert
}
