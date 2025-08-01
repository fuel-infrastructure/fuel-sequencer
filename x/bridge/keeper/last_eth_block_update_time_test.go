package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
)

func createTestLastEthBlockUpdateTime(keeper bridgekeeper.Keeper, ctx context.Context) time.Time {
	item := time.Now().Round(0)
	keeper.SetLastEthBlockUpdateTime(ctx, item)
	return item
}

func TestLastEthBlockUpdateTimeGet(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestLastEthBlockUpdateTime(keeper, ctx)
	rst, found := keeper.GetLastEthBlockUpdateTime(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestLastEthBlockUpdateTimeRemove(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	createTestLastEthBlockUpdateTime(keeper, ctx)
	keeper.RemoveLastEthBlockUpdateTime(ctx)
	_, found := keeper.GetLastEthBlockUpdateTime(ctx)
	require.False(t, found)
}
