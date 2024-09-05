package keeper

import (
	"context"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastInjectedTxsSequence set lastInjectedTxsSequence in the store
func (k Keeper) SetLastInjectedTxsSequence(ctx context.Context, lastInjectedTxsSequence math.Int) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)

	b, err := lastInjectedTxsSequence.Marshal()
	if err != nil {
		panic(err)
	}
	store.Set([]byte{0}, b)
}

// GetLastInjectedTxsSequence returns lastInjectedTxsSequence
func (k Keeper) GetLastInjectedTxsSequence(ctx context.Context) (val math.Int, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)

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

// MustGetLastInjectedTxsSequence returns lastInjectedTxsSequence and panics if it doesn't find it
func (k Keeper) MustGetLastInjectedTxsSequence(ctx context.Context) math.Int {
	val, found := k.GetLastInjectedTxsSequence(ctx)
	if !found {
		panic("expected to find LastInjectedTxsSequence")
	}
	return val
}

// MustGetNextInjectedTxsSequence increments LastInjectedTxsSequence by one and returns the result.
// It panics if LastInjectedTxsSequence is not found
func (k Keeper) MustGetNextInjectedTxsSequence(ctx context.Context) math.Int {
	val, found := k.GetLastInjectedTxsSequence(ctx)
	if !found {
		panic("expected to find LastInjectedTxsSequence")
	}

	newVal := val.AddRaw(1)
	k.SetLastInjectedTxsSequence(ctx, newVal)

	return newVal
}

// RemoveLastInjectedTxsSequence removes lastInjectedTxsSequence from the store
func (k Keeper) RemoveLastInjectedTxsSequence(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastInjectedTxsSequenceKey)
	store.Delete([]byte{0})
}
