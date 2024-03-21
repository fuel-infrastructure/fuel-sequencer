package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func createTestEthEventsTx(keeper keeper.Keeper, ctx context.Context, blockHeight uint64) types.EthEventsTx {
	item := types.EthEventsTx{
		Events: []*sidecartypes.Event{
			{
				EventType: "authorize",
				Data:      []byte("auth"),
			},
		},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
		BlockNumber:      math.NewInt(int64(blockHeight)),
	}
	keeper.SetEthEventsTx(ctx, item)
	return item
}

func TestSetAndGetEthEventsTx(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	blockHeight := uint64(1)
	item := createTestEthEventsTx(keeper, ctx, blockHeight)

	// Test GetEthEventsTx
	retrievedItem, found := keeper.GetEthEventsTx(ctx, blockHeight)
	require.True(t, found, "EthEventsTx should be found")
	require.Equal(t, item, retrievedItem, "retrieved EthEventsTx should match the set one")
}

func TestRemoveEthEventsTx(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	blockHeight := uint64(1)
	createTestEthEventsTx(keeper, ctx, blockHeight)

	// Remove and then try to get the EthEventsTx
	keeper.RemoveEthEventsTx(ctx, blockHeight)
	_, found := keeper.GetEthEventsTx(ctx, blockHeight)
	require.False(t, found, "EthEventsTx should not be found after removal")
}
