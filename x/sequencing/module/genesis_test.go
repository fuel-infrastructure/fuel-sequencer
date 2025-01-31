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
				Owner: "fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm",
				Order: math.ZeroInt(),
			},
			{
				Id:    testutil.MockTopicIDHex(1),
				Owner: "fuelsequencer163rsv65t4893t2rz5rmda9sly7lgdlq2jgr36m",
				Order: math.ZeroInt(),
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SequencingKeeper(t)
	err := sequencing.InitGenesis(ctx, k, genesisState)
	require.NoError(t, err)
	got, err := sequencing.ExportGenesis(ctx, k)
	require.NoError(t, err)
	require.NotNil(t, got)

	require.Equal(t, genesisState.Params, got.Params)
	require.Equal(t, genesisState.TopicList, got.TopicList)
	// this line is used by starport scaffolding # genesis/test/assert
}
