package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastEthereumNonce set lastEthereumNonce in the store
func (k Keeper) SetLastEthereumNonce(ctx context.Context, lastEthereumNonce math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumNonceKey)

	b, err := lastEthereumNonce.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetLastEthereumNonce returns lastEthereumNonce
func (k Keeper) GetLastEthereumNonce(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
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

// MustGetLastEthereumNonce returns lastEthereumNonce and panics if it does't find it
func (k Keeper) MustGetLastEthereumNonce(ctx context.Context) math.Int {
	val, found := k.GetLastEthereumNonce(ctx)
	if !found {
		panic("expected to find LastEthereumNonce")
	}
	return val
}

// RemoveLastEthereumNonce removes lastEthereumNonce from the store
func (k Keeper) RemoveLastEthereumNonce(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumNonceKey)
	store.Delete([]byte{0})
}
