package testutil

import (
	"testing"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// BlobKeeper creates a test blob keeper
func BlobKeeper(tb testing.TB) (*keeper.Keeper, sdk.Context, *codec.ProtoCodec) {
	tb.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	tStoreKey := storetypes.NewTransientStoreKey(types.TransientStoreKey)

	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())

	encCfg := moduletestutil.MakeTestEncodingConfig()
	cdc := codec.NewProtoCodec(encCfg.InterfaceRegistry)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		tStoreKey,
		"cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn", // dummy authority
	)

	return k, ctx, cdc
}

// SetupBlobKeeper creates a blob keeper with custom parameters
func SetupBlobKeeper(
	tb testing.TB,
	storeService store.KVStoreService,
	cdc codec.BinaryCodec,
	authority string,
) (*keeper.Keeper, sdk.Context) {
	tb.Helper()

	tStoreKey := storetypes.NewTransientStoreKey(types.TransientStoreKey)
	ctx := sdk.NewContext(nil, cmtproto.Header{}, false, log.NewNopLogger())

	k := keeper.NewKeeper(
		cdc,
		storeService,
		log.NewNopLogger(),
		tStoreKey,
		authority,
	)

	return k, ctx
}

// MakeTestEncodingConfig creates a test encoding config for blob module
func MakeTestEncodingConfig() moduletestutil.TestEncodingConfig {
	cfg := moduletestutil.MakeTestEncodingConfig()

	// Register interfaces
	registry := cfg.InterfaceRegistry
	types.RegisterInterfaces(registry)

	return cfg
}
