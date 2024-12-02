package reports_test

import (
	"testing"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/nullify"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	reports "github.com/fuel-infrastructure/fuel-sequencer/x/reports/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis_ValidState(t *testing.T) {
	genesisState := types.GenesisState{
		Params:          types.DefaultParams(),
		SlashReportList: []types.SlashReport{testtypes.ValidSlashReport1, testtypes.ValidSlashReport2},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ReportsKeeper(t)
	reports.InitGenesis(ctx, k, genesisState)
	got := reports.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.SlashReportList, got.SlashReportList)
	// this line is used by starport scaffolding # genesis/test/assert

	// Verify other genesis state elements as needed
	require.Equal(t, genesisState.Params, got.Params)
}

func TestGenesis_InvalidSlashReport(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		SlashReportList: []types.SlashReport{
			testtypes.ValidSlashReport1,
			testtypes.InvalidSlashReportNonUniqueSlashEntries,
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ReportsKeeper(t)

	require.Panics(t, func() {
		defer func() {
			if r := recover(); r != nil {
				require.Contains(t, r.(string), "duplicate slash entry found")
			}
		}()

		reports.InitGenesis(ctx, k, genesisState)
	})
}
