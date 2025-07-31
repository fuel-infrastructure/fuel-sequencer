package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// ResetEthereumEventIndexOffset resets ethereumEventIndexOffset in the store
func (k Keeper) ResetEthereumEventIndexOffset(ctx context.Context) {
	k.SetEthereumEventIndexOffset(ctx, 0)
}

// SetEthereumEventIndexOffset sets ethereumEventIndexOffset in the store and metrics server
func (k Keeper) SetEthereumEventIndexOffset(ctx context.Context, ethereumEventIndexOffset uint64) {
	defer metrics.SetEthereumEventIndexOffset(ctx, ethereumEventIndexOffset)
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthereumEventIndexOffsetKey)
	b := sdk.Uint64ToBigEndian(ethereumEventIndexOffset)
	store.Set([]byte{0}, b)
}

// GetEthereumEventIndexOffset returns ethereumEventIndexOffset
func (k Keeper) GetEthereumEventIndexOffset(ctx context.Context) (val uint64, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthereumEventIndexOffsetKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	return sdk.BigEndianToUint64(b), true
}

// MustGetEthereumEventIndexOffset returns ethereumEventIndexOffset and panics if it does't find it
func (k Keeper) MustGetEthereumEventIndexOffset(ctx context.Context) uint64 {
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
