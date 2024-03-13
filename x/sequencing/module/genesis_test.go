package sequencing_test

import (
	"testing"

	"cosmossdk.io/math"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		TopicList: []types.Topic{
			{
				Id:    math.ZeroInt(),
				Owner: "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q",
				Order: math.ZeroInt(),
			},
			{
				Id:    math.OneInt(),
				Owner: "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q",
				Order: math.ZeroInt(),
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SequencingKeeper(t)
	sequencing.InitGenesis(ctx, k, genesisState)
	got := sequencing.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.TopicList, got.TopicList)

	// Check that NextTopicId is set correctly
	nextTopicId := k.MustGetNextTopicId(ctx)
	require.True(t, nextTopicId.Equal(math.NewInt(2)))

	// Verify other genesis state elements as needed
	require.Equal(t, genesisState.Params, got.Params)
}
