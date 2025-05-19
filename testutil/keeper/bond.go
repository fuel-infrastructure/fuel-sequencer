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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func BondKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	k, ctx, _, _, _, _ := BondKeeperWithDependencies(t)
	return k, ctx
}

func BondKeeperWithDependencies(t testing.TB) (
	keeper.Keeper,
	sdk.Context,
	codec.Codec,
	*testutil.MockAccountKeeper,
	*testutil.MockBankKeeper,
	*testutil.MockBridgeKeeper,
) {
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	return BondKeeperFromArgsWithDependencies(t, authority.String())
}

func BondKeeperFromArgsWithDependencies(t testing.TB, authority string) (
	keeper.Keeper,
	sdk.Context,
	codec.Codec,
	*testutil.MockAccountKeeper,
	*testutil.MockBankKeeper,
	*testutil.MockBridgeKeeper,
) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	logger := log.NewNopLogger()

	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)
	bridgeKeeper := testutil.NewMockBridgeKeeper(ctrl)
	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		logger,
		authority,
		accountKeeper,
		bankKeeper,
		bridgeKeeper,
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, logger)

	// Initialize params
	//nolint:errcheck
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	return k, ctx, cdc, accountKeeper, bankKeeper, bridgeKeeper
}
