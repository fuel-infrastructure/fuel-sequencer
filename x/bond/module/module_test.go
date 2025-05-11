package bond_test

import (
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/mock/gomock"
	grpcgateway "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types/module"
	modulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/bond/module"
	testkeeper "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bond "github.com/fuel-infrastructure/fuel-sequencer/x/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	types "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	"google.golang.org/grpc"
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

func TestAppModule_InitGenesis_ModuleAccountMissing(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Expect module account to be retrieved during genesis
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(nil).
		Times(1)

	// Test InitGenesis
	genesisState := types.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// This should panic since we returned nil for the module account
	require.Panics(t, func() {
		appModule.InitGenesis(ctx, cdc, genJSON)
	}, "should panic when module account is not found")
}

func TestAppModule_InitExportGenesis(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Create a mock module account
	mockModuleAccount := authtypes.NewEmptyModuleAccount(types.ModuleName)
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), types.ModuleName).
		Return(mockModuleAccount).
		Times(1)

	// Test InitGenesis
	genesisState := types.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// This should not panic
	require.NotPanics(t, func() { appModule.InitGenesis(ctx, cdc, genJSON) })

	// Test ExportGenesis
	exported := appModule.ExportGenesis(ctx, cdc)
	require.NotNil(t, exported)

	var exportedGenesis types.GenesisState
	err = cdc.UnmarshalJSON(exported, &exportedGenesis)
	require.NoError(t, err)
	require.Equal(t, genesisState.Params, exportedGenesis.Params)
	require.Equal(t, genesisState.State, exportedGenesis.State)
}

func TestAppModule_BeginBlock(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

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

func TestAppModule_BeginBlock_YieldMinting(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	baseTime := time.Now()
	recipient := k.GetAuthority()
	yieldAmount := sdkmath.NewInt(1000000)

	// Test case: no yield parameters set
	require.NoError(t, appModule.BeginBlock(ctx))
	require.Equal(t, int64(0), k.GetYieldMintHeight(ctx))

	// Test case: yield time not reached
	future := baseTime.Add(time.Hour)
	params := types.NewParams(recipient, &future, yieldAmount)
	require.NoError(t, k.SetParams(ctx, params))
	ctx = ctx.WithBlockTime(baseTime)
	require.NoError(t, appModule.BeginBlock(ctx))
	require.Equal(t, int64(0), k.GetYieldMintHeight(ctx))

	// Test case: yield time reached
	params = types.NewParams(recipient, &baseTime, yieldAmount)
	require.Error(t, k.SetParams(ctx, params))

	// Test case: yield already minted
	ctx = ctx.WithBlockTime(baseTime.Add(time.Second))
	require.NoError(t, appModule.BeginBlock(ctx))
	require.Equal(t, ctx.BlockHeight(), k.GetYieldMintHeight(ctx))

	// Test case: invalid recipient
	params = types.NewParams("invalid", &baseTime, yieldAmount)
	require.Error(t, k.SetParams(ctx, params))

	// Test case: zero yield amount
	params = types.NewParams(recipient, &future, sdkmath.ZeroInt())
	require.NoError(t, k.SetParams(ctx, params))
	require.NoError(t, appModule.BeginBlock(ctx))
	require.Equal(t, ctx.BlockHeight(), k.GetYieldMintHeight(ctx))
}

func TestRegisterLegacyAminoCodec(t *testing.T) {
	c := codec.NewLegacyAmino()
	appModuleBasic := bond.NewAppModuleBasic(nil)
	appModuleBasic.RegisterLegacyAminoCodec(c)
	// No panic or error expected
}

func TestRegisterInterfaces(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	appModuleBasic := bond.NewAppModuleBasic(nil)
	appModuleBasic.RegisterInterfaces(registry)
	// No panic or error expected
}

func TestDefaultGenesis(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	appModuleBasic := bond.NewAppModuleBasic(cdc)
	defaultGenesis := appModuleBasic.DefaultGenesis(cdc)
	require.NotNil(t, defaultGenesis)
}

func TestValidateGenesis(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	appModuleBasic := bond.NewAppModuleBasic(cdc)
	defaultGenesis := types.DefaultGenesis()
	bz, err := cdc.MarshalJSON(defaultGenesis)
	require.NoError(t, err)
	err = appModuleBasic.ValidateGenesis(cdc, nil, json.RawMessage(bz))
	require.NoError(t, err)
}

func TestAppModule_RegisterServices(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Create a proper configurator with the necessary dependencies
	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)

	// Create a gRPC server
	server := grpc.NewServer()

	// Create a configurator with the gRPC server
	config := module.NewConfigurator(cdc, server, server)

	// Register services
	require.NotPanics(t, func() {
		appModule.RegisterServices(config)
	})
}

func TestAppModule_RegisterInvariants(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	require.NotPanics(t, func() {
		appModule.RegisterInvariants(nil)
	})
}

func TestAppModule_IsOnePerModuleType(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	require.NotPanics(t, func() {
		appModule.IsOnePerModuleType()
	})
}

func TestAppModule_IsAppModule(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	require.NotPanics(t, func() {
		appModule.IsAppModule()
	})
}

func TestAppModule_RegisterGRPCGatewayRoutes(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	clientCtx := client.Context{}.WithCodec(cdc)
	mux := grpcgateway.NewServeMux()
	require.NotPanics(t, func() {
		appModule.RegisterGRPCGatewayRoutes(clientCtx, mux)
	})
	require.Panics(t, func() {
		appModule.RegisterGRPCGatewayRoutes(clientCtx, nil)
	})
}
