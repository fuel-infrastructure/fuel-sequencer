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

		LastEthereumNonce:       math.NewInt(75),
		LastEthereumBlockSynced: math.NewInt(13),
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.BridgeKeeper(t)
	bridge.InitGenesis(ctx, k, genesisState)
	got := bridge.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.LastEthereumNonce, got.LastEthereumNonce)
	require.Equal(t, genesisState.LastEthereumBlockSynced, got.LastEthereumBlockSynced)
	// this line is used by starport scaffolding # genesis/test/assert
}
