package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetLastEthereumBlockSynced(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := uint64(10)

	keeper.SetLastEthereumBlockSynced(ctx, value)
	rst, found := keeper.GetLastEthereumBlockSynced(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveLastEthereumBlockSynced(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := uint64(10)

	keeper.SetLastEthereumBlockSynced(ctx, value)
	keeper.RemoveLastEthereumBlockSynced(ctx)
	_, found := keeper.GetLastEthereumBlockSynced(ctx)
	require.False(t, found)
}
