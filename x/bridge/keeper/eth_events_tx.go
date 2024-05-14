package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetEthEventsTxIndex set ethEventsTx in the store.
func (k Keeper) SetEthEventsTxIndex(ctx context.Context, ethEventsTx types.EthEventsTxIndex) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxIndexKey)
	b := k.cdc.MustMarshal(&ethEventsTx)
	store.Set(types.EthEventsTxIndexKey, b)
}

// GetEthEventsTxIndex returns ethEventsTx.
func (k Keeper) GetEthEventsTxIndex(ctx context.Context) (val types.EthEventsTxIndex, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxIndexKey)

	b := store.Get(types.EthEventsTxIndexKey)
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEthEventsTxIndex removes EthEventsTx from the store.
func (k Keeper) RemoveEthEventsTxIndex(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxIndexKey)
	store.Delete(types.EthEventsTxIndexKey)
}

// MustGetEthEventsTxIndex returns ethEventsTx and panics otherwise.
func (k Keeper) MustGetEthEventsTxIndex(ctx context.Context) (val types.EthEventsTxIndex) {
	val, ok := k.GetEthEventsTxIndex(ctx)
	if !ok {
		panic("expected to find EthEventsTxIndex")
	}
	return val
}
