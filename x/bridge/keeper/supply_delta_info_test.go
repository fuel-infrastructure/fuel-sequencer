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

func createTestSupplyDeltaInfo(keeper keeper.Keeper, ctx context.Context) types.SupplyDeltaInfo {
	item := types.SupplyDeltaInfo{}
	keeper.SetSupplyDeltaInfo(ctx, item)
	return item
}

func TestSupplyDeltaInfoGet(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestSupplyDeltaInfo(keeper, ctx)
	rst, found := keeper.GetSupplyDeltaInfo(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestSupplyDeltaInfoRemove(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	createTestSupplyDeltaInfo(keeper, ctx)
	keeper.RemoveSupplyDeltaInfo(ctx)
	_, found := keeper.GetSupplyDeltaInfo(ctx)
	require.False(t, found)
}
