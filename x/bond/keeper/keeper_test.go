package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	sdkstore "cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func setupKeeper(t testing.TB) (keeper.Keeper, bondtypes.QueryServer) {
	storeKey := storetypes.NewKVStoreKey(bondtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := runtime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()
	accountKeeper := mock.NewMockAccountKeeper(t)
	bankKeeper := mock.NewMockBankKeeper(t)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
		accountKeeper,
		bankKeeper,
	)
	return k, k
}

func TestNewKeeper(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(bondtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := runtime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()
	accountKeeper := mock.NewMockAccountKeeper(t)
	bankKeeper := mock.NewMockBankKeeper(t)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
		accountKeeper,
		bankKeeper,
	)

	require.NotNil(t, k)
	require.Equal(t, "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", k.GetAuthority())
}

func TestKeeper_GetAuthority(t *testing.T) {
	k, _ := setupKeeper(t)
	require.Equal(t, "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", k.GetAuthority())
}

func TestKeeper_Logger(t *testing.T) {
	k, _ := setupKeeper(t)
	require.NotNil(t, k.Logger())
}
