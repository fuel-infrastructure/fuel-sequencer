package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetLastInjectedTxsSequence(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastInjectedTxsSequence(ctx, value)
	rst, found := keeper.GetLastInjectedTxsSequence(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveLastInjectedTxsSequence(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := math.NewInt(10)

	keeper.SetLastInjectedTxsSequence(ctx, value)
	keeper.RemoveLastInjectedTxsSequence(ctx)
	_, found := keeper.GetLastInjectedTxsSequence(ctx)
	require.False(t, found)
}
