package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func createTestTopic(keeper keeper.Keeper, ctx context.Context, topicId math.Int) types.Topic {
	item := types.Topic{
		Id:    topicId,
		Owner: sample.AccAddress(),
		Order: math.NewInt(0),
	}
	keeper.SetTopic(ctx, item)
	return item
}

func TestGetTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	item := createTestTopic(keeper, ctx, math.ZeroInt())
	rst, found := keeper.GetTopic(ctx, math.ZeroInt())
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestRemoveSupplyDeltaInfo(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	createTestTopic(keeper, ctx, math.ZeroInt())
	keeper.RemoveTopic(ctx, math.ZeroInt())
	_, found := keeper.GetTopic(ctx, math.ZeroInt())
	require.False(t, found)
}

func TestGetAllTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	createTestTopic(keeper, ctx, math.ZeroInt())
	createTestTopic(keeper, ctx, math.OneInt())
	createTestTopic(keeper, ctx, math.NewInt(2))

	topics := keeper.GetAllTopic(ctx)
	require.Equal(t, topics[0].Id, math.ZeroInt())
	require.Equal(t, topics[1].Id, math.OneInt())
	require.Equal(t, topics[2].Id, math.NewInt(2))
	require.Len(t, topics, 3)
}

func TestHasTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	item := createTestTopic(keeper, ctx, math.ZeroInt())

	has := keeper.HasTopic(ctx, item.Id.String())
	require.True(t, has)

	has = keeper.HasTopic(ctx, "nonexistent")
	require.False(t, has)
}

func TestSetAndGetNextTopicId(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	initialId := math.NewInt(5)
	keeper.SetNextTopicId(ctx, initialId)

	retrievedId := keeper.MustGetNextTopicId(ctx)
	require.True(t, initialId.Equal(retrievedId))
}
