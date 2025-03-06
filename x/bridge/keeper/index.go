package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetIndex sets Index in the store.
func (k Keeper) SetIndex(ctx context.Context, index types.Index) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.IndexKey)
	b := k.cdc.MustMarshal(&index)
	store.Set(types.IndexKey, b)
}

// GetIndex returns Index.
func (k Keeper) GetIndex(ctx context.Context) (val types.Index, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.IndexKey)

	b := store.Get(types.IndexKey)
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveIndex removes Index from the store.
func (k Keeper) RemoveIndex(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.IndexKey)
	store.Delete(types.IndexKey)
}

// MustGetIndex returns Index and panics otherwise.
func (k Keeper) MustGetIndex(ctx context.Context) (val types.Index) {
	val, ok := k.GetIndex(ctx)
	if !ok {
		panic("expected to find Index")
	}
	return val
}
