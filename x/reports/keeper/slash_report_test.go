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

func TestGetSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntries := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 10)
	for _, expectedSlashEntry := range slashEntries {
		actualSlashEntry, found := testKeeper.GetSlashEntry(
			ctx, height, expectedSlashEntry.DelegatorAddress, expectedSlashEntry.ValidatorAddress,
		)
		require.True(t, found)
		require.Equal(t, expectedSlashEntry, actualSlashEntry)
	}
}

func TestRemoveSlashReport(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	slashReports := keepertest.CreateNSlashReport(testKeeper, ctx, 10)
	for _, slashReport := range slashReports {
		testKeeper.RemoveSlashReport(ctx, slashReport.Height)
		_, found := testKeeper.GetSlashReport(ctx, slashReport.Height)
		require.False(t, found)
	}
}

func TestRemoveSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntries := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 10)
	for _, slashEntry := range slashEntries {
		testKeeper.RemoveSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		_, found := testKeeper.GetSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
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

func TestHasSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntry := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 1)[0]

	has := testKeeper.HasSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
	require.True(t, has)

	// SlashEntry in state has a different height
	has = testKeeper.HasSlashEntry(
		ctx, height+1, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress,
	)
	require.False(t, has)

	// SlashEntry in state has a different delegator address
	has = testKeeper.HasSlashEntry(
		ctx, height, "bad_delegator_address", slashEntry.ValidatorAddress,
	)
	require.False(t, has)

	// SlashEntry in state has a different validator address
	has = testKeeper.HasSlashEntry(
		ctx, height, slashEntry.DelegatorAddress, "bad_validator_address",
	)
	require.False(t, has)
}
