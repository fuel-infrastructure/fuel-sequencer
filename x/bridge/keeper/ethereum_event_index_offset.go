package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// ResetEthereumEventIndexOffset resets ethereumEventIndexOffset in the store
func (k Keeper) ResetEthereumEventIndexOffset(ctx context.Context) {
	k.SetEthereumEventIndexOffset(ctx, math.ZeroInt())
}

// SetEthereumEventIndexOffset sets ethereumEventIndexOffset in the store
func (k Keeper) SetEthereumEventIndexOffset(ctx context.Context, ethereumEventIndexOffset math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthereumEventIndexOffsetKey)

	b, err := ethereumEventIndexOffset.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetEthereumEventIndexOffset returns ethereumEventIndexOffset
func (k Keeper) GetEthereumEventIndexOffset(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthereumEventIndexOffsetKey)

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

// MustGetEthereumEventIndexOffset returns ethereumEventIndexOffset and panics if it does't find it
func (k Keeper) MustGetEthereumEventIndexOffset(ctx context.Context) math.Int {
	val, found := k.GetEthereumEventIndexOffset(ctx)
	if !found {
		panic("expected to find EthereumEventIndexOffset")
	}
	return val
}

// RemoveEthereumEventIndexOffset removes ethereumEventIndexOffset from the store
func (k Keeper) RemoveEthereumEventIndexOffset(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthereumEventIndexOffsetKey)
	store.Delete([]byte{0})
}
