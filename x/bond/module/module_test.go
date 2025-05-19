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
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"google.golang.org/grpc"
)

func setupModule(t testing.TB) (*bond.AppModule, bondtypes.AccountKeeper, bondtypes.BankKeeper, keeper.Keeper) {
	storeKey := storetypes.NewKVStoreKey(bondtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	_ = codec.NewProtoCodec(registry) // cdc is not used

	// Create keeper with dependencies
	k, _, depCdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	// Use depCdc (codec.Codec) for appModule
	appModule := bond.NewAppModule(depCdc, k, mockAccountKeeper, mockBankKeeper)

	return &appModule, mockAccountKeeper, mockBankKeeper, k
}

func TestAppModuleBasic(t *testing.T) {
	appModule, _, _, _ := setupModule(t)

	require.Equal(t, bondtypes.ModuleName, appModule.Name())

	// Test ConsensusVersion
	require.Equal(t, uint64(1), appModule.ConsensusVersion())

	// Test DefaultGenesis
	genState := bondtypes.DefaultGenesis()
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
	// Create keeper with dependencies
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	// Create app module
	appModule := bond.NewAppModule(cdc, k, mockAccountKeeper, mockBankKeeper)

	// Expect module account to be retrieved during genesis
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), bondtypes.ModuleName).
		Return(nil).
		Times(1)

	// Test InitGenesis
	genesisState := bondtypes.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// This should panic since we returned nil for the module account
	require.Panics(t, func() {
		appModule.InitGenesis(ctx, cdc, genJSON)
	}, "should panic when module account is not found")
}

func TestAppModule_InitExportGenesis(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Create a mock module account
	mockModuleAccount := authtypes.NewEmptyModuleAccount(bondtypes.ModuleName)
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), bondtypes.ModuleName).
		Return(mockModuleAccount).
		Times(1)

	// Test InitGenesis
	genesisState := bondtypes.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// This should not panic
	require.NotPanics(t, func() { appModule.InitGenesis(ctx, cdc, genJSON) })

	// Test ExportGenesis
	exported := appModule.ExportGenesis(ctx, cdc)
	require.NotNil(t, exported)

	var exportedGenesis bondtypes.GenesisState
	err = cdc.UnmarshalJSON(exported, &exportedGenesis)
	require.NoError(t, err)
	require.Equal(t, genesisState.Params, exportedGenesis.Params)
	require.Equal(t, genesisState.State, exportedGenesis.State)
}

func TestAppModule_BeginBlock(t *testing.T) {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

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
	storeService := sdkruntime.NewKVStoreService(storetypes.NewKVStoreKey(bondtypes.StoreKey))
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

type testSetup struct {
	k                 keeper.Keeper
	ctx               sdk.Context
	cdc               codec.Codec
	mockAccountKeeper bondtypes.AccountKeeper
	mockBankKeeper    bondtypes.BankKeeper
	mockBridgeKeeper  *testutil.MockBridgeKeeper
	appModule         bond.AppModule
	baseTime          time.Time
	recipient         string
	yieldAmount       sdkmath.Int
}

func setupYieldMintingTest(t *testing.T) testSetup {
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, mockBridgeKeeper := testkeeper.BondKeeperWithDependencies(t)

	// Create app module
	appModule := bond.NewAppModule(cdc, k, mockAccountKeeper, mockBankKeeper)

	// Set up test context
	baseTime := time.Now()
	recipient := k.GetAuthority()
	yieldAmount := sdkmath.NewInt(1000000)

	// Set up bridge keeper expectations
	mockBridgeKeeper.EXPECT().
		GetParams(gomock.Any()).
		Return(bridgetypes.Params{
			BridgeDenom: "utest",
		}).
		AnyTimes()

	// Set up bank keeper expectations
	mockBankKeeper.EXPECT().
		MintCoins(gomock.Any(), bondtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("utest", yieldAmount))).
		Return(nil).
		AnyTimes()

	mockBankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			bondtypes.ModuleName,
			gomock.Any(),
			sdk.NewCoins(sdk.NewCoin("utest", yieldAmount)),
		).
		Return(nil).
		AnyTimes()

	// Set up account keeper expectations
	mockAccountKeeper.EXPECT().
		AddressCodec().
		Return(testutil.MockAddressCodec{}).
		AnyTimes()

	return testSetup{
		k:                 k,
		ctx:               ctx,
		cdc:               cdc,
		mockAccountKeeper: mockAccountKeeper,
		mockBankKeeper:    mockBankKeeper,
		mockBridgeKeeper:  mockBridgeKeeper,
		appModule:         appModule,
		baseTime:          baseTime,
		recipient:         recipient,
		yieldAmount:       yieldAmount,
	}
}

func TestAppModule_BeginBlock_NoYieldParameters(t *testing.T) {
	setup := setupYieldMintingTest(t)
	require.NoError(t, setup.appModule.BeginBlock(setup.ctx))
	require.Equal(t, int64(0), setup.k.GetYieldMintHeight(setup.ctx))
}

func TestAppModule_BeginBlock_YieldTimeNotReached(t *testing.T) {
	setup := setupYieldMintingTest(t)
	future := setup.baseTime.Add(time.Hour)

	params := bondtypes.NewParams(setup.recipient, &future, setup.yieldAmount)
	require.NoError(t, setup.k.SetParams(setup.ctx, params))
	setup.ctx = setup.ctx.WithBlockTime(setup.baseTime)

	require.NoError(t, setup.appModule.BeginBlock(setup.ctx))
	require.Equal(t, int64(0), setup.k.GetYieldMintHeight(setup.ctx))
}

func TestAppModule_BeginBlock_YieldTimeReached(t *testing.T) {
	setup := setupYieldMintingTest(t)
	params := bondtypes.NewParams(setup.recipient, &setup.baseTime, setup.yieldAmount)
	require.Error(t, setup.k.SetParams(setup.ctx, params))
}

func TestAppModule_BeginBlock_YieldAlreadyMinted(t *testing.T) {
	setup := setupYieldMintingTest(t)
	setup.ctx = setup.ctx.WithBlockTime(setup.baseTime.Add(time.Second))
	require.NoError(t, setup.appModule.BeginBlock(setup.ctx))
	require.Equal(t, setup.ctx.BlockHeight(), setup.k.GetYieldMintHeight(setup.ctx))
}

func TestAppModule_BeginBlock_InvalidRecipient(t *testing.T) {
	setup := setupYieldMintingTest(t)
	params := bondtypes.NewParams("invalid", &setup.baseTime, setup.yieldAmount)
	require.Error(t, setup.k.SetParams(setup.ctx, params))
}

func TestAppModule_BeginBlock_ZeroYieldAmount(t *testing.T) {
	setup := setupYieldMintingTest(t)
	future := setup.baseTime.Add(time.Hour)

	params := bondtypes.NewParams(setup.recipient, &future, sdkmath.ZeroInt())
	require.NoError(t, setup.k.SetParams(setup.ctx, params))
	require.NoError(t, setup.appModule.BeginBlock(setup.ctx))
	require.Equal(t, setup.ctx.BlockHeight(), setup.k.GetYieldMintHeight(setup.ctx))
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
	defaultGenesis := bondtypes.DefaultGenesis()
	bz, err := cdc.MarshalJSON(defaultGenesis)
	require.NoError(t, err)
	err = appModuleBasic.ValidateGenesis(cdc, nil, json.RawMessage(bz))
	require.NoError(t, err)
}

func TestAppModule_RegisterServices(t *testing.T) {
	k, _, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	appModule := bond.NewAppModule(
		cdc,
		k,
		mockAccountKeeper,
		mockBankKeeper,
	)

	// Create a proper configurator with the necessary dependencies
	registry := codectypes.NewInterfaceRegistry()
	bondtypes.RegisterInterfaces(registry)

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
	k, _, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

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
	k, _, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

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
	k, _, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

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
	k, _, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

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

func TestAppModule_InitGenesis(t *testing.T) {
	// Create keeper with dependencies
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	// Create app module
	appModule := bond.NewAppModule(cdc, k, mockAccountKeeper, mockBankKeeper)

	// Create a mock module account
	mockModuleAccount := authtypes.NewEmptyModuleAccount(bondtypes.ModuleName)
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), bondtypes.ModuleName).
		Return(mockModuleAccount).
		Times(1)

	// Test InitGenesis
	genesisState := bondtypes.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// Test InitGenesis
	appModule.InitGenesis(ctx, cdc, genJSON)
}

func TestAppModule_ExportGenesis(t *testing.T) {
	// Create keeper with dependencies
	k, ctx, cdc, mockAccountKeeper, mockBankKeeper, _ := testkeeper.BondKeeperWithDependencies(t)

	// Set up mock expectations
	mockAccountKeeper.EXPECT().
		GetModuleAccount(gomock.Any(), bondtypes.ModuleName).
		Return(authtypes.NewEmptyModuleAccount(bondtypes.ModuleName))

	// Create app module
	appModule := bond.NewAppModule(cdc, k, mockAccountKeeper, mockBankKeeper)

	// Initialize the store with default genesis state
	genesisState := bondtypes.DefaultGenesis()
	genJSON, err := json.Marshal(genesisState)
	require.NoError(t, err)

	// Initialize the store with the genesis state
	appModule.InitGenesis(ctx, cdc, genJSON)

	// Export genesis
	exported := appModule.ExportGenesis(ctx, cdc)
	require.NotNil(t, exported)

	var exportedGenesis bondtypes.GenesisState
	err = cdc.UnmarshalJSON(exported, &exportedGenesis)
	require.NoError(t, err)
	require.Equal(t, bondtypes.DefaultGenesis().Params, exportedGenesis.Params)
	require.Equal(t, bondtypes.DefaultGenesis().State, exportedGenesis.State)
}

func TestAppModuleBasic_DefaultGenesis(t *testing.T) {
	cdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
	appModuleBasic := bond.NewAppModuleBasic(cdc)
	defaultGenesis := bondtypes.DefaultGenesis()
	bz, err := cdc.MarshalJSON(defaultGenesis)
	require.NoError(t, err)
	require.Equal(t, json.RawMessage(bz), appModuleBasic.DefaultGenesis(cdc))
}

func TestAppModuleBasic_RegisterGRPCGatewayRoutes(t *testing.T) {
	// Create a proper configurator with the necessary dependencies
	registry := codectypes.NewInterfaceRegistry()
	bondtypes.RegisterInterfaces(registry)

	// Create a client context
	clientCtx := client.Context{}.WithCodec(codec.NewProtoCodec(registry))

	// Create a mux
	mux := grpcgateway.NewServeMux()

	// Create app module basic
	appModuleBasic := bond.NewAppModuleBasic(codec.NewProtoCodec(registry))

	// Register gRPC gateway routes
	appModuleBasic.RegisterGRPCGatewayRoutes(clientCtx, mux)

	// No assertions needed as this is just a registration test
}
