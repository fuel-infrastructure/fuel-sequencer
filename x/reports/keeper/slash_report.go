package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
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

// RemoveSlashEntry removes a SlashEntry from the store
func (k Keeper) RemoveSlashEntry(ctx context.Context, height uint64, delegatorAddress, validatorAddress string) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	store.Delete(types.SlashEntryKeyPrefix(height, delegatorAddress, validatorAddress))
}

// GetAllSlashReport returns all SlashReport
func (k Keeper) GetAllSlashReport(ctx context.Context) (list []types.SlashReport) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	// map slash entries by height
	reportMap := make(map[uint64][]types.SlashEntry)
	for ; iterator.Valid(); iterator.Next() {
		var slashEntry types.SlashEntry
		k.cdc.MustUnmarshal(iterator.Value(), &slashEntry)
		height := types.ExtractHeightFromSlashEntryKey(iterator.Key())
		reportMap[height] = append(reportMap[height], slashEntry)
	}

	// Convert map to slice of SlashReports
	for height, slashEntries := range reportMap {
		list = append(list, types.SlashReport{
			Height:  height,
			Entries: slashEntries,
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

// HasSlashEntry checks if a HasSlashEntry exists in the store.
func (k Keeper) HasSlashEntry(ctx context.Context, height uint64, delegatorAddress, validatorAddress string) bool {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SlashReportKey))
	slashEntryKey := types.SlashEntryKeyPrefix(height, delegatorAddress, validatorAddress)
	return store.Has(slashEntryKey)
}
