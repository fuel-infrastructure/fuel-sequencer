package types_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	"github.com/stretchr/testify/require"
)

func TestValidateGenesisState(t *testing.T) {
	tests := []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TopicList: []types.Topic{
					{
						Id: testutil.MockTopicIDHex(0),
					},
					{
						Id: testutil.MockTopicIDHex(1),
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicated topic",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				TopicList: []types.Topic{
					{
						Id: testutil.MockTopicIDHex(0),
					},
					{
						Id: testutil.MockTopicIDHex(0),
					},
				},
			},
			valid: false,
		},
		{
			desc: "params not set",
			genState: &types.GenesisState{
				TopicList: []types.Topic{
					{
						Id: testutil.MockTopicIDHex(0),
					},
					{
						Id: testutil.MockTopicIDHex(1),
					},
				},
			},
			valid: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
