package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func createTestIndex(keeper keeper.Keeper, ctx context.Context) types.Index {
	item := types.Index{
		NumInjectedTxsTotal: 2,
		NumInjectedTxsAnte:  1,
		NumSpecialTxsTotal:  4,
		NumSpecialTxsExec:   3,
	}
	keeper.SetIndex(ctx, item)
	return item
}

func TestSetAndGetIndex(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestIndex(keeper, ctx)

	// Test GetIndex
	retrievedItem, found := keeper.GetIndex(ctx)
	require.True(t, found, "Index should be found")
	require.Equal(t, item, retrievedItem, "retrieved Index should match the set one")
}

func TestRemoveIndex(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	createTestIndex(keeper, ctx)

	// Remove and then try to get the Index
	keeper.RemoveIndex(ctx)
	_, found := keeper.GetIndex(ctx)
	require.False(t, found, "Index should not be found after removal")
}
