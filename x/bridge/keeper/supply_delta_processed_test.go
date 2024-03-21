package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func createTestSupplyDeltaProcessed(keeper keeper.Keeper, ctx context.Context) types.SupplyDeltaProcessed {
	item := types.SupplyDeltaProcessed{}
	keeper.SetSupplyDeltaProcessed(ctx, item)
	return item
}

func TestSupplyDeltaProcessedGet(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestSupplyDeltaProcessed(keeper, ctx)
	rst, found := keeper.GetSupplyDeltaProcessed(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestSupplyDeltaProcessedRemove(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	createTestSupplyDeltaProcessed(keeper, ctx)
	keeper.RemoveSupplyDeltaProcessed(ctx)
	_, found := keeper.GetSupplyDeltaProcessed(ctx)
	require.False(t, found)
}
