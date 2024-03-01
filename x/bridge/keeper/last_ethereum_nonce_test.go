package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetLastEthereumNonce(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastEthereumNonce(ctx, value)
	rst, found := keeper.GetLastEthereumNonce(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveLastEthereumNonce(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastEthereumNonce(ctx, value)
	keeper.RemoveLastEthereumNonce(ctx)
	_, found := keeper.GetLastEthereumNonce(ctx)
	require.False(t, found)
}
