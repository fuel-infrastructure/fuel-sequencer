package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetSupplyDeltaProcessed set supplyDeltaProcessed in the store
func (k Keeper) SetSupplyDeltaProcessed(ctx context.Context, supplyDeltaProcessed types.SupplyDeltaProcessed) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SupplyDeltaProcessedKey))
	b := k.cdc.MustMarshal(&supplyDeltaProcessed)
	store.Set([]byte{0}, b)
}

// GetSupplyDeltaProcessed returns supplyDeltaProcessed
func (k Keeper) GetSupplyDeltaProcessed(ctx context.Context) (val types.SupplyDeltaProcessed, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SupplyDeltaProcessedKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

func (k Keeper) MustGetSupplyDeltaProcessed(ctx context.Context) (val types.SupplyDeltaProcessed) {
	val, found := k.GetSupplyDeltaProcessed(ctx)
	if !found {
		panic("expected to find SupplyDeltaProcessed")
	}

	return val
}

// RemoveSupplyDeltaProcessed removes supplyDeltaProcessed from the store
func (k Keeper) RemoveSupplyDeltaProcessed(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.KeyPrefix(types.SupplyDeltaProcessedKey))
	store.Delete([]byte{0})
}
