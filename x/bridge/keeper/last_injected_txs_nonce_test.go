package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetLastInjectedTxsNonce(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastInjectedTxsNonce(ctx, value)
	rst, found := keeper.GetLastInjectedTxsNonce(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveLastInjectedTxsNonce(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastInjectedTxsNonce(ctx, value)
	keeper.RemoveLastInjectedTxsNonce(ctx)
	_, found := keeper.GetLastInjectedTxsNonce(ctx)
	require.False(t, found)
}
