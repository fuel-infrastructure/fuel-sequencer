package reports_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	reports "github.com/fuel-infrastructure/fuel-sequencer/x/reports/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
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

	require.Equal(
		t, keepertest.OrderSlashReportsLexicographically(genesisState.SlashReportList), got.SlashReportList,
	)
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

	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = r.(error).Error()
			}
		}()

		reports.InitGenesis(ctx, k, genesisState)
		require.Fail(t, "Expected panic")
	}()
	require.Contains(t, panicMsg, "duplicate slash entry found")
}
