package keeper

import (
	"context"
	"time"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastEthBlockUpdateTime sets lastEthBlockUpdateTime in the store
func (k Keeper) SetLastEthBlockUpdateTime(ctx context.Context, lastEthBlockUpdateTime time.Time) {
	defer metrics.SetLastEthBlockUpdate(ctx, lastEthBlockUpdateTime)
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthBlockUpdateTimeKey)
	b := sdk.Uint64ToBigEndian(uint64(lastEthBlockUpdateTime.UnixNano()))
	store.Set([]byte{0}, b)
}

// GetLastEthBlockUpdateTime returns lastEthBlockUpdateTime
func (k Keeper) GetLastEthBlockUpdateTime(ctx context.Context) (val time.Time, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthBlockUpdateTimeKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}
	nanosecondCount := sdk.BigEndianToUint64(b)

	return time.Unix(0, int64(nanosecondCount)), true
}

// MustGetLastEthBlockUpdateTime returns lastEthBlockUpdateTime and panics if it doesn't find it
func (k Keeper) MustGetLastEthBlockUpdateTime(ctx context.Context) time.Time {
	val, found := k.GetLastEthBlockUpdateTime(ctx)
	if !found {
		panic("expected to find LastEthBlockUpdateTime")
	}
	return val
}

// RemoveLastEthBlockUpdateTime removes lastEthBlockUpdateTime from the store
func (k Keeper) RemoveLastEthBlockUpdateTime(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastEthBlockUpdateTimeKey)
	store.Delete([]byte{0})
}
