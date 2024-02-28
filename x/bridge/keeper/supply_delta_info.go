package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetSupplyDeltaInfo set supplyDeltaInfo in the store
func (k Keeper) SetSupplyDeltaInfo(ctx context.Context, supplyDeltaInfo types.SupplyDeltaInfo) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)
	b := k.cdc.MustMarshal(&supplyDeltaInfo)
	store.Set([]byte{0}, b)
}

// GetSupplyDeltaInfo returns supplyDeltaInfo
func (k Keeper) GetSupplyDeltaInfo(ctx context.Context) (val types.SupplyDeltaInfo, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveSupplyDeltaInfo removes supplyDeltaInfo from the store
func (k Keeper) RemoveSupplyDeltaInfo(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)
	store.Delete([]byte{0})
}
