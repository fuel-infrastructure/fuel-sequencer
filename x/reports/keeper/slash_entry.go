package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// SetSlashEntry sets a SlashEntry in the store by its height, delegator address and validator address.
func (k Keeper) SetSlashEntry(ctx context.Context, slashEntryHeight uint64, slashEntry types.SlashEntry) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	b := k.cdc.MustMarshal(&slashEntry)
	slashEntryKeyPrefix := types.SlashEntryKeyPrefix(
		slashEntryHeight, slashEntry.DelegatorAddress, slashEntry.ValidatorAddress,
	)
	store.Set(slashEntryKeyPrefix, b)
}

// GetSlashEntry returns a SlashEntry by its height, delegator address and validator address
func (k Keeper) GetSlashEntry(
	ctx context.Context, height uint64, delegatorAddress, validatorAddress string,
) (val types.SlashEntry, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))

	b := store.Get(types.SlashEntryKeyPrefix(height, delegatorAddress, validatorAddress))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSlashEntry removes a SlashEntry from the store
func (k Keeper) RemoveSlashEntry(ctx context.Context, height uint64, delegatorAddress, validatorAddress string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	store.Delete(types.SlashEntryKeyPrefix(height, delegatorAddress, validatorAddress))
}

// HasSlashEntry checks if a HasSlashEntry exists in the store.
func (k Keeper) HasSlashEntry(ctx context.Context, height uint64, delegatorAddress, validatorAddress string) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	slashEntryKey := types.SlashEntryKeyPrefix(height, delegatorAddress, validatorAddress)
	return store.Has(slashEntryKey)
}

// IterateSlashEntries iterates through all slash entries for a particular height and calls a callback on every entry.
func (k Keeper) IterateSlashEntries(
	ctx context.Context, height uint64, cb func(slashEntry types.SlashEntry) (stop bool),
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iteratorPrefix := types.SlashReportKeyPrefix(height)
	iterator := storetypes.KVStorePrefixIterator(store, iteratorPrefix)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var slashEntry types.SlashEntry
		k.cdc.MustUnmarshal(iterator.Value(), &slashEntry)

		if cb(slashEntry) {
			break
		}
	}
}

// InsertSlashEntry is a helper function that adds slash entries with pre-checks. For instance, it ensures that an
// existing slash entry's slashed amount is not overwritten, but incremented.
func (k Keeper) InsertSlashEntry(
	ctx context.Context, valAddr sdk.ValAddress, delAddr sdk.AccAddress, slashAmount sdkmath.LegacyDec,
) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Make sure that the height is positive, as slashing is not expected at heights less or equal to zero.
	if sdkCtx.BlockHeight() <= 0 {
		return fmt.Errorf("slashing height must be positive, received: %v", sdkCtx.BlockHeight())
	}

	// Compute height, validator address and delegator address.
	// Note: uint64 is bigger than int64, therefore, the height can be converted safely to uint64.
	height := uint64(sdkCtx.BlockHeight())
	validatorAddress := valAddr.String()
	delegatorAddress := delAddr.String()

	// Make sure that slashed amount is positive. We do not want to store any entries in state if they were not slashed.
	if !slashAmount.IsPositive() {
		return fmt.Errorf("slash amount must be positive, received %s", slashAmount.String())
	}

	// Check if a slash entry already exists in state for the current height.
	slashEntry, found := k.GetSlashEntry(ctx, height, delegatorAddress, validatorAddress)
	if found {
		// If a slash entry already exists, increment the slashed amount as it must be that the delegator has already
		// been slashed at this height. Set the unbonding and bonded balances to zero as that should be computed by
		// other functionality that is called from the reports module's BeginBlocker.
		slashEntry.DelegatorSlashAmount = slashEntry.DelegatorSlashAmount.Add(slashAmount)
		slashEntry.DelegatorBondedBalance = sdkmath.LegacyZeroDec()
		slashEntry.DelegatorUnbondingBalance = sdkmath.LegacyZeroDec()
	} else {
		// Otherwise create a new slash entry. For the same reason as the found=true case, delegator bonded and
		// unbonding balances should be set to zero.
		slashEntry = types.SlashEntry{
			ValidatorAddress:          validatorAddress,
			DelegatorAddress:          delegatorAddress,
			DelegatorSlashAmount:      slashAmount,
			DelegatorBondedBalance:    sdkmath.LegacyZeroDec(),
			DelegatorUnbondingBalance: sdkmath.LegacyZeroDec(),
		}
	}

	// Make sure that the slash entry satisfies the basic validation checks.
	if err := slashEntry.ValidateBasic(); err != nil {
		return errors.Wrap(err, "constructed invalid slash entry")
	}

	k.SetSlashEntry(ctx, height, slashEntry)

	return nil
}
