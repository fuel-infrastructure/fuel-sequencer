package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastInjectedTxsSequence sets lastInjectedTxsSequence in the store
func (k Keeper) SetLastInjectedTxsSequence(ctx context.Context, lastInjectedTxsSequence uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)
	b := sdk.Uint64ToBigEndian(lastInjectedTxsSequence)
	store.Set([]byte{0}, b)
}

// GetLastInjectedTxsSequence returns lastInjectedTxsSequence
func (k Keeper) GetLastInjectedTxsSequence(ctx context.Context) (val uint64, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	return sdk.BigEndianToUint64(b), true
}

// MustGetLastInjectedTxsSequence returns lastInjectedTxsSequence and panics if it doesn't find it
func (k Keeper) MustGetLastInjectedTxsSequence(ctx context.Context) uint64 {
	val, found := k.GetLastInjectedTxsSequence(ctx)
	if !found {
		panic("expected to find LastInjectedTxsSequence")
	}
	return val
}

// MustGetNextInjectedTxsSequence increments LastInjectedTxsSequence by one and returns the result.
// It panics if LastInjectedTxsSequence is not found
func (k Keeper) MustGetNextInjectedTxsSequence(ctx context.Context) uint64 {
	val, found := k.GetLastInjectedTxsSequence(ctx)
	if !found {
		panic("expected to find LastInjectedTxsSequence")
	}

	newVal := val + 1
	k.SetLastInjectedTxsSequence(ctx, newVal)

	return newVal
}

// RemoveLastInjectedTxsSequence removes lastInjectedTxsSequence from the store
func (k Keeper) RemoveLastInjectedTxsSequence(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)
	store.Delete([]byte{0})
}
