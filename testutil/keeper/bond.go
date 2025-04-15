package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func BondKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	logger := log.NewNopLogger()

	var accountKeeper types.AccountKeeper = nil
	// for mock: mock.NewMockAccountKeeper(t)
	var bankKeeper types.BankKeeper = nil
	// for mock: mock.NewMockBankKeeper(t)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		logger,
		authority.String(),
		accountKeeper,
		bankKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	// Initialize params
	//nolint:errcheck
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	return k, ctx
}
