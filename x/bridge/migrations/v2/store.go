package v2

import (
	corestoretypes "cosmossdk.io/core/store"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types/legacy"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// migrateParams migrates params to the new structure.
func migrateParams(store corestoretypes.KVStore, cdc codec.BinaryCodec) error {

	// Get and unmarshal
	var params legacy.Params
	adapter := runtime.KVStoreAdapter(store)
	bz := adapter.Get(types.ParamsKey)
	if err := cdc.Unmarshal(bz, &params); err != nil {
		return err
	}

	// Move params to new structure, disregarding the removed params.
	newParams := types.NewParams(
		params.BridgeDenom,
		params.BridgeDenomTotalSupply,
		params.EthereumProxyContractAddress,
		params.SupplyDeltaPeriod,
		params.VestingStartTime,
		params.AdditionalBlockedAddresses,
		params.MaxEthBlockUpdateDelay,
		params.SequencerTxsAllocation,
	)

	// Marshal and set
	bz, err := cdc.Marshal(&newParams)
	if err != nil {
		return err
	}
	adapter.Set(types.ParamsKey, bz)

	return nil
}

// MigrateStore performs in-place store migrations from v1 to v2. The migration includes:
//
// - Switch to new parameters.
func MigrateStore(ctx sdk.Context, storeService corestoretypes.KVStoreService, cdc codec.BinaryCodec) error {
	store := storeService.OpenKVStore(ctx)
	return migrateParams(store, cdc)
}
