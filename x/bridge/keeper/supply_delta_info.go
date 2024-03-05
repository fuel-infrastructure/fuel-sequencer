package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// SetSupplyDeltaInfo set supplyDeltaInfo in the store
func (k Keeper) SetSupplyDeltaInfo(ctx context.Context, supplyDeltaInfo types.SupplyDeltaInfo) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)
	b := k.cdc.MustMarshal(&supplyDeltaInfo)
	store.Set([]byte{0}, b)
}

// GetSupplyDeltaInfo returns supplyDeltaInfo
func (k Keeper) GetSupplyDeltaInfo(ctx context.Context) (val types.SupplyDeltaInfo, found bool) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// MustGetSupplyDeltaInfo returns supplyDeltaInfo
func (k Keeper) MustGetSupplyDeltaInfo(ctx context.Context) types.SupplyDeltaInfo {
	val, found := k.GetSupplyDeltaInfo(ctx)
	if !found {
		panic("expected to find supply delta info")
	}
	return val
}

// MustResetSupplyDeltaInfo resets the offset and delta values. NOTE: LastSupply is not safe to reset as this needs to
// be continuously tracked by the blockchain. MustResetSupplyDeltaInfo panics if SupplyDeltaInfo is not found
func (k Keeper) MustResetSupplyDeltaInfo(ctx context.Context) {
	val, found := k.GetSupplyDeltaInfo(ctx)
	if !found {
		panic("expected to find supply delta info")
	}

	val.Delta = sdkmath.ZeroInt()
	val.Offset = sdkmath.ZeroInt()

	k.SetSupplyDeltaInfo(ctx, val)
}

// RemoveSupplyDeltaInfo removes supplyDeltaInfo from the store
func (k Keeper) RemoveSupplyDeltaInfo(ctx context.Context) {
	storeAdapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	store := prefix.NewStore(storeAdapter, types.SupplyDeltaInfoKey)
	store.Delete([]byte{0})
}

// UpdatedSupplyDeltaInfoWithNewDelta notes down any changes in supply of the bridge token
func (k Keeper) UpdatedSupplyDeltaInfoWithNewDelta(ctx sdk.Context, bankKeeper types.BankKeeper) {

	// Get latest recorded supply and actual supply.
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)
	latestSupply := bankKeeper.GetSupply(ctx, k.GetParams(ctx).BridgeDenom).Amount

	// Update if supply has changed.
	if !latestSupply.Equal(supplyDeltaInfo.LastSupply) {

		// If positive, new delta will add to the stored delta.
		// Otherwise, it will subtract from the stored delta.
		newDelta := latestSupply.Sub(supplyDeltaInfo.LastSupply)
		supplyDeltaInfo.Delta = supplyDeltaInfo.Delta.Add(newDelta)

		ctx.Logger().Debug("recorded change in bridge token supply", "delta", newDelta, "supply", latestSupply)

		supplyDeltaInfo.LastSupply = latestSupply
		k.SetSupplyDeltaInfo(ctx, supplyDeltaInfo)
	}
}
