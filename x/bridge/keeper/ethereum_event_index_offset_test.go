package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetEthereumEventIndexOffset(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetEthereumEventIndexOffset(ctx, value)
	rst, found := keeper.GetEthereumEventIndexOffset(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveEthereumEventIndexOffset(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetEthereumEventIndexOffset(ctx, value)
	keeper.RemoveEthereumEventIndexOffset(ctx)
	_, found := keeper.GetEthereumEventIndexOffset(ctx)
	require.False(t, found)
}
