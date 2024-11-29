package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// SetSlashReport set a specific SlashReport in the store by its height
func (k Keeper) SetSlashReport(ctx context.Context, slashReport types.SlashReport) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	b := k.cdc.MustMarshal(&slashReport)
	store.Set(types.SlashReportKeyPrefix(slashReport.Height), b)
}

// GetSlashReport returns a SlashReport by its height
func (k Keeper) GetSlashReport(
	ctx context.Context,
	height uint64,
) (val types.SlashReport, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))

	b := store.Get(types.SlashReportKeyPrefix(height))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSlashReport removes a SlashReport from the store
func (k Keeper) RemoveSlashReport(
	ctx context.Context,
	height uint64,
) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	store.Delete(types.SlashReportKeyPrefix(height))
}

// GetAllSlashReport returns all SlashReport
func (k Keeper) GetAllSlashReport(ctx context.Context) (list []types.SlashReport) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.SlashReport
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// HasSlashReport checks if a SlashReport exists in the store.
func (k Keeper) HasSlashReport(ctx context.Context, height uint64) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	slashReportKey := types.SlashReportKeyPrefix(height)
	return store.Has(slashReportKey)
}
