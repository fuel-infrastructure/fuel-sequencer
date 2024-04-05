package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastEthereumBlockSynced sets lastEthereumBlockSynced in the store
func (k Keeper) SetLastEthereumBlockSynced(ctx context.Context, lastEthereumBlockSynced uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)

	b := sdk.Uint64ToBigEndian(lastEthereumBlockSynced)
	store.Set([]byte{0}, b)
}

// GetLastEthereumBlockSynced returns lastEthereumBlockSynced
func (k Keeper) GetLastEthereumBlockSynced(ctx context.Context) (val uint64, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	return sdk.BigEndianToUint64(b), true
}

// MustGetLastEthereumBlockSynced returns lastEthereumBlockSynced and panics if it does't find it
func (k Keeper) MustGetLastEthereumBlockSynced(ctx context.Context) uint64 {
	val, found := k.GetLastEthereumBlockSynced(ctx)
	if !found {
		panic("expected to find LastEthereumBlockSynced")
	}
	return val
}

// RemoveLastEthereumBlockSynced removes lastEthereumBlockSynced from the store
func (k Keeper) RemoveLastEthereumBlockSynced(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthereumBlockSyncedKey)
	store.Delete([]byte{0})
}
