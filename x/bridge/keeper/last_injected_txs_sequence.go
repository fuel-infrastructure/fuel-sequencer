package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetLastConsensusTxsSequence sets lastConsensusTxsSequence in the store
func (k Keeper) SetLastConsensusTxsSequence(ctx context.Context, lastConsensusTxsSequence uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastConsensusTxsSequenceKey)
	b := sdk.Uint64ToBigEndian(lastConsensusTxsSequence)
	store.Set([]byte{0}, b)
}

// GetLastConsensusTxsSequence returns lastConsensusTxsSequence
func (k Keeper) GetLastConsensusTxsSequence(ctx context.Context) (val uint64, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastConsensusTxsSequenceKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	return sdk.BigEndianToUint64(b), true
}

// MustGetLastConsensusTxsSequence returns lastConsensusTxsSequence and panics if it doesn't find it
func (k Keeper) MustGetLastConsensusTxsSequence(ctx context.Context) uint64 {
	val, found := k.GetLastConsensusTxsSequence(ctx)
	if !found {
		panic("expected to find LastConsensusTxsSequence")
	}
	return val
}

// MustGetNextConsensusTxsSequence increments LastConsensusTxsSequence by one and returns the result.
// It panics if LastConsensusTxsSequence is not found
func (k Keeper) MustGetNextConsensusTxsSequence(ctx context.Context) uint64 {
	newVal := k.MustGetLastConsensusTxsSequence(ctx) + 1
	k.SetLastConsensusTxsSequence(ctx, newVal)
	return newVal
}

// RemoveLastConsensusTxsSequence removes lastConsensusTxsSequence from the store
func (k Keeper) RemoveLastConsensusTxsSequence(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.LastConsensusTxsSequenceKey)
	store.Delete([]byte{0})
}
