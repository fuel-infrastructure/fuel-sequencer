package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetEthEventsTx set ethEventsTx in the store
func (k Keeper) SetEthEventsTx(ctx context.Context, ethEventsTx types.EthEventsTx) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxPrefixKey)
	b := k.cdc.MustMarshal(&ethEventsTx)
	store.Set(types.EthEventsTxKey(ethEventsTx.BlockNumber.Uint64()), b)
}

// GetEthEventsTx returns ethEventsTx
func (k Keeper) GetEthEventsTx(ctx context.Context, blockHeight uint64) (val types.EthEventsTx, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxPrefixKey)

	b := store.Get(types.EthEventsTxKey(blockHeight))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEthEventsTx removes EthEventsTx from the store
func (k Keeper) RemoveEthEventsTx(ctx context.Context, blockHeight uint64) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.EthEventsTxPrefixKey)
	store.Delete(types.EthEventsTxKey(blockHeight))
}
