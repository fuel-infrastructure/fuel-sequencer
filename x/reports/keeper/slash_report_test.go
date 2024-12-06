package keeper_test

import (
	"testing"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/stretchr/testify/require"
)

func TestGetSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	for _, expectedSlashReport := range slashReports {
		actualSlashReport, found := testKeeper.GetSlashReport(ctx, expectedSlashReport.Height)
		require.True(t, found)
		require.Equal(t, keepertest.OrderSlashReportLexicographically(expectedSlashReport), actualSlashReport)
	}
}

func TestRemoveSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	for _, slashReport := range slashReports {
		// Report must be is state before deleted
		_, found := testKeeper.GetSlashReport(ctx, slashReport.Height)
		require.True(t, found)

		testKeeper.RemoveSlashReport(ctx, slashReport.Height)
		_, found = testKeeper.GetSlashReport(ctx, slashReport.Height)
		require.False(t, found)
	}
}

func TestGetAllSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	require.Equal(t,
		keepertest.OrderSlashReportsLexicographically(slashReports),
		testKeeper.GetAllSlashReport(ctx),
	)
}

func TestHasSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReport := keepertest.CreateNSlashReport(testKeeper, ctx, 1)[0]

	has := testKeeper.HasSlashReport(ctx, slashReport.Height)
	require.True(t, has)

	has = testKeeper.HasSlashReport(ctx, 2) // SlashReport in state has height 1
	require.False(t, has)
}
