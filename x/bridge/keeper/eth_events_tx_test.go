package keeper_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func createTestEthEventsTxIndex(keeper keeper.Keeper, ctx context.Context) types.EthEventsTxIndex {
	item := types.EthEventsTxIndex{
		NumUnhandledEventTxs: 2,
	}
	keeper.SetEthEventsTxIndex(ctx, item)
	return item
}

func TestSetAndGetEthEventsTx(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	item := createTestEthEventsTxIndex(keeper, ctx)

	// Test GetEthEventsTx
	retrievedItem, found := keeper.GetEthEventsTxIndex(ctx)
	require.True(t, found, "EthEventsTx should be found")
	require.Equal(t, item, retrievedItem, "retrieved EthEventsTx should match the set one")
}

func TestRemoveEthEventsTx(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	createTestEthEventsTxIndex(keeper, ctx)

	// Remove and then try to get the EthEventsTx
	keeper.RemoveEthEventsTxIndex(ctx)
	_, found := keeper.GetEthEventsTxIndex(ctx)
	require.False(t, found, "EthEventsTx should not be found after removal")
}
