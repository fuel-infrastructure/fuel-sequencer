package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastEthereumBlockSynced set lastEthereumBlockSynced in the store
func (k Keeper) SetLastEthereumBlockSynced(ctx context.Context, lastEthereumBlockSynced math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)

	b, err := lastEthereumBlockSynced.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetLastEthereumBlockSynced returns lastEthereumBlockSynced
func (k Keeper) GetLastEthereumBlockSynced(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)

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

// MustGetLastEthereumBlockSynced returns lastEthereumBlockSynced and panics if it does't find it
func (k Keeper) MustGetLastEthereumBlockSynced(ctx context.Context) math.Int {
	val, found := k.GetLastEthereumBlockSynced(ctx)
	if !found {
		panic("expected to find last ethereum nonce")
	}
	return val
}

// RemoveLastEthereumBlockSynced removes lastEthereumBlockSynced from the store
func (k Keeper) RemoveLastEthereumBlockSynced(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)
	store.Delete([]byte{0})
}
