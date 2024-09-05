package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastInjectedTxsNonce set lastInjectedTxsNonce in the store
func (k Keeper) SetLastInjectedTxsNonce(ctx context.Context, lastInjectedTxsNonce math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsNonceKey)

	b, err := lastInjectedTxsNonce.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetLastInjectedTxsNonce returns lastInjectedTxsNonce
func (k Keeper) GetLastInjectedTxsNonce(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsNonceKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	err := val.Unmarshal(b)
	if err != nil {
		panic(err)
	}
	return val, true
}

// MustGetLastInjectedTxsNonce returns lastInjectedTxsNonce and panics if it doesn't find it
func (k Keeper) MustGetLastInjectedTxsNonce(ctx context.Context) math.Int {
	val, found := k.GetLastInjectedTxsNonce(ctx)
	if !found {
		panic("expected to find LastInjectedTxsNonce")
	}
	return val
}

// MustGetNextInjectedTxsNonce increments LastInjectedTxsNonce by one and returns the result.
// It panics if LastInjectedTxsNonce is not found
func (k Keeper) MustGetNextInjectedTxsNonce(ctx context.Context) math.Int {
	val, found := k.GetLastInjectedTxsNonce(ctx)
	if !found {
		panic("expected to find LastInjectedTxsNonce")
	}

	newVal := val.AddRaw(1)
	k.SetLastInjectedTxsNonce(ctx, newVal)

	return newVal
}

// RemoveLastInjectedTxsNonce removes lastInjectedTxsNonce from the store
func (k Keeper) RemoveLastInjectedTxsNonce(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsNonceKey)
	store.Delete([]byte{0})
}
