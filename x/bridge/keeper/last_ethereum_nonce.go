package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastEthereumNonce set lastEthereumNonce in the store
func (k Keeper) SetLastEthereumNonce(ctx context.Context, lastEthereumNonce math.Int) {
	defer metrics.SetLastEthereumNonce(ctx, lastEthereumNonce)
	storeAdapter := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumNonceKey)

	b, err := lastEthereumNonce.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetLastEthereumNonce returns lastEthereumNonce
func (k Keeper) GetLastEthereumNonce(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumNonceKey)

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

// MustGetLastEthereumNonce returns lastEthereumNonce and panics if it doesn't find it
func (k Keeper) MustGetLastEthereumNonce(ctx context.Context) math.Int {
	val, found := k.GetLastEthereumNonce(ctx)
	if !found {
		panic("expected to find LastEthereumNonce")
	}
	return val
}

// MustGetNextEthereumNonce returns LastEthereumNonce + 1.
// It panics if LastEthereumNonce is not found.
func (k Keeper) MustGetNextEthereumNonce(ctx context.Context) math.Int {
	return k.MustGetLastEthereumNonce(ctx).AddRaw(1)
}

// MustGetNextEthereumNonceAndIncrement calculates LastEthereumNonce + 1, saves it, and returns the result.
// It panics if LastEthereumNonce is not found.
func (k Keeper) MustGetNextEthereumNonceAndIncrement(ctx context.Context) math.Int {
	newVal := k.MustGetNextEthereumNonce(ctx)
	k.SetLastEthereumNonce(ctx, newVal)
	return newVal
}

// RemoveLastEthereumNonce removes lastEthereumNonce from the store
func (k Keeper) RemoveLastEthereumNonce(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.Environment.KVStoreService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumNonceKey)
	store.Delete([]byte{0})
}
