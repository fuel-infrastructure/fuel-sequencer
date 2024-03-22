package keeper_test

import (
	"context"
	"encoding/hex"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	utilstest "github.com/fuel-infrastructure/fuel-sequencer/testutil/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func createTestTopic(keeper keeper.Keeper, ctx context.Context, num int) types.Topic {
	topicId := utilstest.MockTopicIDHex(num)
	item := types.Topic{
		Id:    topicId,
		Owner: sample.AccAddress(),
		Order: math.ZeroInt(),
	}
	keeper.SetTopic(ctx, item)
	return item
}

func TestGetTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	item := createTestTopic(keeper, ctx, 1)
	rst, found := keeper.GetTopic(ctx, item.Id)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestRemoveTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	item := createTestTopic(keeper, ctx, 1)
	keeper.RemoveTopic(ctx, item.Id)
	_, found := keeper.GetTopic(ctx, item.Id)
	require.False(t, found)
}

func TestGetAllTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	// Create multiple topics with unique IDs
	expectedTopics := []types.Topic{
		createTestTopic(keeper, ctx, 1),
		createTestTopic(keeper, ctx, 2),
		createTestTopic(keeper, ctx, 3),
	}

	topics := keeper.GetAllTopic(ctx)
	require.Len(t, topics, len(expectedTopics))

	// Create a map of the expected IDs for easy lookup
	expectedIDs := make(map[string]bool)
	for _, topic := range expectedTopics {
		expectedIDs[hex.EncodeToString(topic.Id)] = true
	}

	// Check each retrieved topic is expected
	for _, topic := range topics {
		_, found := expectedIDs[hex.EncodeToString(topic.Id)]
		require.True(t, found, "Unexpected topic ID found")
	}
}

func TestHasTopic(t *testing.T) {
	keeper, ctx := keepertest.SequencingKeeper(t)
	item := createTestTopic(keeper, ctx, 1)

	has := keeper.HasTopic(ctx, item.Id)
	require.True(t, has)

	has = keeper.HasTopic(ctx, []byte("nonexistent"))
	require.False(t, has)
}
