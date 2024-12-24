package keeper

import (
	"context"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// SetSlashReport sets a specific SlashReport in the store slash entry by slash entry using the slash report height,
// delegator address and validator address. We will be storing individual slash entries for optimized querying by the
// staking module hooks.
func (k Keeper) SetSlashReport(ctx context.Context, slashReport types.SlashReport) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	for _, slashEntry := range slashReport.Entries {
		b := k.cdc.MustMarshal(&slashEntry)
		slashEntryKeyPrefix := types.SlashEntryKeyPrefix(
			slashReport.Height, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress,
		)
		store.Set(slashEntryKeyPrefix, b)
	}
}

// GetSlashReport returns a SlashReport by its height. The SlashReport needs to be reconstructed from the individual
// SlashEntries
func (k Keeper) GetSlashReport(ctx context.Context, height uint64) (val types.SlashReport, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iteratorPrefix := types.SlashReportKeyPrefix(height)
	iterator := storetypes.KVStorePrefixIterator(store, iteratorPrefix)

	defer iterator.Close()

	var slashEntries []types.SlashEntry
	for ; iterator.Valid(); iterator.Next() {
		var slashEntry types.SlashEntry
		k.cdc.MustUnmarshal(iterator.Value(), &slashEntry)
		slashEntries = append(slashEntries, slashEntry)
	}

	if len(slashEntries) == 0 {
		return val, false
	} else {
		val.Height = height
		val.Entries = slashEntries
	}

	return val, true
}

// RemoveSlashReport removes a SlashReport from the store. It iterates through the individual slash entries for a
// particular height and deletes them one by one.
func (k Keeper) RemoveSlashReport(ctx context.Context, height uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iteratorPrefix := types.SlashReportKeyPrefix(height)
	iterator := storetypes.KVStorePrefixIterator(store, iteratorPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

// GetAllSlashReport returns all SlashReport
func (k Keeper) GetAllSlashReport(ctx context.Context) (list []types.SlashReport) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	// Iterate over entries in state and map them into individual slash reports by height. We have to be careful as to
	// not use maps, as this will make the ordering of the returned slash reports non-deterministic.
	var currentHeight uint64
	var currentEntries []types.SlashEntry
	for ; iterator.Valid(); iterator.Next() {
		var slashEntry types.SlashEntry
		k.cdc.MustUnmarshal(iterator.Value(), &slashEntry)
		height := types.ExtractHeightFromSlashEntryKey(iterator.Key())

		if currentEntries == nil {

			// First entry, we have to define both currentHeight and currentEntries
			currentHeight = height
			currentEntries = []types.SlashEntry{slashEntry}
		} else if height == currentHeight {

			// Same height as current, add to current entries
			currentEntries = append(currentEntries, slashEntry)
		} else {

			// New height, finalize current report and start new one
			list = append(list, types.SlashReport{
				Height:  currentHeight,
				Entries: currentEntries,
			})
			currentHeight = height
			currentEntries = []types.SlashEntry{slashEntry}
		}
	}

	// Add final report if we have entries
	if currentEntries != nil {
		list = append(list, types.SlashReport{
			Height:  currentHeight,
			Entries: currentEntries,
		})
	}

	return
}

// HasSlashReport checks if a SlashReport exists in the store.
func (k Keeper) HasSlashReport(ctx context.Context, height uint64) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iteratorPrefix := types.SlashReportKeyPrefix(height)
	iterator := storetypes.KVStorePrefixIterator(store, iteratorPrefix)
	defer iterator.Close()
	return iterator.Valid()
}

// UpdateSlashReportBalancesAtCurrentHeight calculates the bonded and unbonding balances for delegators mentioned in the
// slash report generated at this same height.
func (k Keeper) UpdateSlashReportBalancesAtCurrentHeight(ctx sdk.Context) error {

	report, found := k.GetSlashReport(ctx, uint64(ctx.BlockHeight()))
	if !found {
		return nil // no report
	}

	for i, entry := range report.Entries {

		// We can assume addresses are bech32 encoded since the reports are not populated from hex addresses.

		delAddr := sdk.MustAccAddressFromBech32(entry.DelegatorAddress)
		valAddr, err := sdk.ValAddressFromBech32(entry.ValidatorAddress)
		if err != nil {
			panic(fmt.Sprintf("unexpected invalid validator address in slash report: %s", entry.ValidatorAddress))
		}

		// Calculate delegator bonded balance

		del, err := k.stakingKeeper.GetDelegation(ctx, delAddr, valAddr)
		if err != nil {
			if !errors.Is(err, stakingtypes.ErrNoDelegation) {
				return err
			}
			report.Entries[i].DelegatorBondedBalance = sdkmath.ZeroInt()
		} else {

			// When calculating tokens from shares we always want to truncate, to not report tokens that the delegator
			// does not actually have. An example of this reasoning in practice is RemoveDelShares, which calculates
			// the tokens returned from removing a number of shared from a validator.
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/staking/types/validator.go#L414

			validator, err := k.stakingKeeper.GetValidator(ctx, valAddr)
			if err != nil {
				return err
			}
			report.Entries[i].DelegatorBondedBalance = validator.TokensFromShares(del.Shares).TruncateInt()
		}

		// Calculate delegator unbonding balance

		ubd, err := k.stakingKeeper.GetUnbondingDelegation(ctx, delAddr, valAddr)
		if err != nil {
			if !errors.Is(err, stakingtypes.ErrNoUnbondingDelegation) {
				return err
			}
			report.Entries[i].DelegatorUnbondingBalance = sdkmath.ZeroInt()
		} else {

			// Replicate logic from GetDelegatorUnbonding but for just one validator.
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/staking/keeper/delegation.go#L268

			unbonding := sdkmath.ZeroInt()
			for _, entry := range ubd.Entries {
				unbonding = unbonding.Add(entry.Balance)
			}
			report.Entries[i].DelegatorUnbondingBalance = unbonding
		}
	}

	err := ctx.EventManager().EmitTypedEvent(&types.EventSlashReportGenerated{
		Height:     uint64(ctx.BlockHeight()),
		NumEntries: uint64(len(report.Entries)),
	})
	if err != nil {
		return err
	}

	k.SetSlashReport(ctx, report)
	return nil
}
