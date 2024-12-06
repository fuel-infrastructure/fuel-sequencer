package keeper_test

import (
	"testing"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/stretchr/testify/require"
)

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

func TestRemoveSlashEntry(t *testing.T) {
	testKeeper, ctx := keepertest.ReportsKeeper(t)
	height := uint64(100)
	slashEntries := keepertest.CreateNSlashEntry(testKeeper, ctx, height, 10)
	for _, slashEntry := range slashEntries {
		// Entry must be is state before deleted
		_, found := testKeeper.GetSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		require.True(t, found)

		testKeeper.RemoveSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		_, found = testKeeper.GetSlashEntry(ctx, height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress)
		require.False(t, found)
	}
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
