package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	v2 "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/migrations/v2"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types/legacy"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig().Codec
	govKey := storetypes.NewKVStoreKey("bridge")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	defaultParams := types.DefaultParams()
	oldParams := legacy.Params{
		BridgeDenom:                  defaultParams.BridgeDenom,
		BridgeDenomTotalSupply:       defaultParams.BridgeDenomTotalSupply,
		EthereumProxyContractAddress: defaultParams.EthereumProxyContractAddress,
		AuthorizeMessagesAllowed:     []string{"DummyMsgTypeUrl1", "DummyMsgTypeUrl2", "DummyMsgTypeUrl3"},
		SupplyDeltaPeriod:            defaultParams.SupplyDeltaPeriod,
		VestingStartTime:             defaultParams.VestingStartTime,
		AdditionalBlockedAddresses:   defaultParams.AdditionalBlockedAddresses,
		MaxEthBlockUpdateDelay:       defaultParams.MaxEthBlockUpdateDelay,
		InjectedEventTxMaxBytes:      1024,
		SequencerTxsAllocation:       defaultParams.SequencerTxsAllocation,
		MaxAuthorizeMessages:         5,
	}

	expectNewParams := types.Params{
		BridgeDenom:                  defaultParams.BridgeDenom,
		BridgeDenomTotalSupply:       defaultParams.BridgeDenomTotalSupply,
		EthereumProxyContractAddress: defaultParams.EthereumProxyContractAddress,
		SupplyDeltaPeriod:            defaultParams.SupplyDeltaPeriod,
		VestingStartTime:             defaultParams.VestingStartTime,
		AdditionalBlockedAddresses:   defaultParams.AdditionalBlockedAddresses,
		MaxEthBlockUpdateDelay:       defaultParams.MaxEthBlockUpdateDelay,
		InjectedEventTxMaxBytes:      1024,
		SequencerTxsAllocation:       defaultParams.SequencerTxsAllocation,
	}

	// Set old params
	oldParamsBz := cdc.MustMarshal(&oldParams)
	store.Set(types.ParamsKey, oldParamsBz)

	// Check that getting and unmarshalling the params before the migration does not work
	var newParams types.Params
	newParamsBz := store.Get(types.ParamsKey)
	require.Error(t, cdc.Unmarshal(newParamsBz, &newParams))

	// Run migration
	storeService := runtime.NewKVStoreService(govKey)
	err := v2.MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	// Getting and unmarshalling the params after the migration works
	newParamsBz = store.Get(types.ParamsKey)
	require.NoError(t, cdc.Unmarshal(newParamsBz, &newParams))

	// Check that params are as expected
	require.Equal(t, expectNewParams, newParams)
}
