package sequencing_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	sequencing "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		TopicList: []types.Topic{
			{
				Id:    testutil.MockTopicIDHex(0),
				Owner: "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q",
				Order: math.ZeroInt(),
			},
			{
				Id:    testutil.MockTopicIDHex(1),
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

	require.Equal(t, genesisState.Params, got.Params)
	require.Equal(t, genesisState.TopicList, got.TopicList)
	// this line is used by starport scaffolding # genesis/test/assert
}
