package types_test

import (
	"testing"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
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
				Params:          types.NewParams(),
				SlashReportList: []types.SlashReport{testtypes.ValidSlashReport1, testtypes.ValidSlashReport2},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated SlashReport",
			genState: &types.GenesisState{
				Params:          types.NewParams(),
				SlashReportList: []types.SlashReport{testtypes.ValidSlashReport1, testtypes.ValidSlashReport1},
			},
			valid: false,
		},
		//{ // TODO: Enable test when params are no longer an empty struct
		//	desc: "params not set",
		//	genState: &types.GenesisState{
		//		SlashReportList: []types.SlashReport{
		//			{
		//				Height: 0,
		//				Entries: []string{"entry1", "entry2"},
		//			},
		//			{
		//				Height: 1,
		//				Entries: []string{"entry3", "entry4"},
		//			},
		//		},
		//	},
		//	valid: false,
		//},
		// this line is used by starport scaffolding # types/genesis/testcase
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
