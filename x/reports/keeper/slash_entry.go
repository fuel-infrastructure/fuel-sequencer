package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
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
