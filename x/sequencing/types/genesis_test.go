package types_test

import (
	"testing"

	"cosmossdk.io/math"
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

				TopicList: []types.Topic{
					{
						Id: math.ZeroInt(),
					},
					{
						Id: math.OneInt(),
					},
				},
			},
			valid: true,
		},
		{
			desc: "duplicated topic",
			genState: &types.GenesisState{
				TopicList: []types.Topic{
					{
						Id: math.ZeroInt(),
					},
					{
						Id: math.ZeroInt(),
					},
				},
			},
			valid: false,
		},
		{
			desc: "non-sequential topic IDs",
			genState: &types.GenesisState{
				TopicList: []types.Topic{
					{
						Id: math.ZeroInt(),
					},
					{
						Id: math.NewInt(2),
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
