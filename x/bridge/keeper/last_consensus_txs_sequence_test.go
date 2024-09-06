package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
)

func TestGetLastConsensusTxsSequence(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := uint64(10)

	keeper.SetLastConsensusTxsSequence(ctx, value)
	rst, found := keeper.GetLastConsensusTxsSequence(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&value),
		nullify.Fill(&rst),
	)
}

func TestRemoveLastConsensusTxsSequence(t *testing.T) {
	keeper, ctx := keepertest.BridgeKeeper(t)
	value := uint64(10)

	keeper.SetLastConsensusTxsSequence(ctx, value)
	keeper.RemoveLastConsensusTxsSequence(ctx)
	_, found := keeper.GetLastConsensusTxsSequence(ctx)
	require.False(t, found)
}
