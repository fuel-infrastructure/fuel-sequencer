package bond_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	sdkstore "cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkruntime "github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	grpcgateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"

	modulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bond "github.com/fuel-infrastructure/fuel-sequencer/x/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func setupModule(t testing.TB) (*bond.AppModule, types.AccountKeeper, types.BankKeeper, keeper.Keeper) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := sdkruntime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAccountKeeper := testutil.NewMockAccountKeeper(ctrl)
	mockBankKeeper := testutil.NewMockBankKeeper(ctrl)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
		mockAccountKeeper,
		mockBankKeeper,
	)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	return &appModule, mockAccountKeeper, mockBankKeeper, k
}

func TestAppModuleBasic(t *testing.T) {
	appModule, _, _, _ := setupModule(t)

	require.Equal(t, types.ModuleName, appModule.Name())

	// Test ConsensusVersion
	require.Equal(t, uint64(1), appModule.ConsensusVersion())

	// Test DefaultGenesis
	genState := types.DefaultGenesis()
	require.NotNil(t, genState)

	// Test ValidateGenesis
	err := genState.Validate()
	require.NoError(t, err)

	// Test RegisterGRPCGatewayRoutes
	clientCtx := client.Context{}
	mux := grpcgateway.NewServeMux()
	appModule.RegisterGRPCGatewayRoutes(clientCtx, mux)
}

func TestAppModule_InitExportGenesis(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := sdkruntime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAccountKeeper := testutil.NewMockAccountKeeper(ctrl)
	mockBankKeeper := testutil.NewMockBankKeeper(ctrl)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
		mockAccountKeeper,
		mockBankKeeper,
	)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)

	// Test InitGenesis
	genesisState := types.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	appModule.InitGenesis(ctx, codec.NewProtoCodec(codectypes.NewInterfaceRegistry()), genJSON)

	// Test ExportGenesis
	exported := appModule.ExportGenesis(ctx, codec.NewProtoCodec(codectypes.NewInterfaceRegistry()))
	require.NotNil(t, exported)

	var exportedGenesis types.GenesisState
	err = json.Unmarshal(exported, &exportedGenesis)
	require.NoError(t, err)
	require.Equal(t, genesisState.Params, exportedGenesis.Params)
}

func TestAppModule_BeginBlock(t *testing.T) {
	appModule, _, _, _ := setupModule(t)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	// BeginBlock should not panic
	require.NoError(t, appModule.BeginBlock(ctx))
}

func TestAppModule_EndBlock(t *testing.T) {
	appModule, _, _, _ := setupModule(t)
	ctx := sdk.NewContext(nil, tmproto.Header{}, false, log.NewNopLogger())

	// EndBlock should not panic
	require.NoError(t, appModule.EndBlock(ctx))
}

func TestProvideModule(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := sdkruntime.NewKVStoreService(storetypes.NewKVStoreKey(types.StoreKey))
	logger := log.NewNopLogger()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	inputs := bond.ModuleInputs{
		StoreService:  storeService,
		Cdc:           cdc,
		Config:        &modulev1.Module{},
		Logger:        logger,
		AccountKeeper: accountKeeper,
		BankKeeper:    bankKeeper,
	}

	outputs := bond.ProvideModule(inputs)
	require.NotNil(t, outputs.BondKeeper)
	require.NotNil(t, outputs.Module)
}
