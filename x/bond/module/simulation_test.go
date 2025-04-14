package bond_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bond "github.com/fuel-infrastructure/fuel-sequencer/x/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestSimulation(t *testing.T) {
	// Create test dependencies
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	logger := log.NewNopLogger()
	bankKeeper := mock.NewMockBankKeeper(t)

	// Create test module
	appModule := bond.NewAppModule(
		cdc,
		keeper.NewKeeper(
			cdc,
			runtime.NewKVStoreService(storeKey),
			logger,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(), // Will be set to governance module account
			bankKeeper,
		),
		nil, // Account keeper will be set in actual app
		nil, // Bank keeper will be set in actual app
	)

	// Create test account
	addr, err := sdk.AccAddressFromBech32("cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu")
	require.NoError(t, err)

	// Create simulation state
	simState := module.SimulationState{
		AppParams: make(simtypes.AppParams),
		Cdc:       cdc,
		Rand:      nil,
		GenState:  make(map[string]json.RawMessage),
		Accounts:  []simtypes.Account{{Address: addr}},
	}

	// Test genesis simulation
	require.NotPanics(t, func() {
		appModule.GenerateGenesisState(&simState)
		require.NotNil(t, simState.GenState[types.ModuleName])
	})

	// Test random weighted operations
	require.NotPanics(t, func() {
		ops := appModule.WeightedOperations(simState)
		require.Empty(t, ops) // Currently no weighted operations
	})

	// Test random weighted proposal contents
	require.NotPanics(t, func() {
		msgs := appModule.ProposalMsgs(simState)
		require.Empty(t, msgs) // Currently no proposal messages
	})
}
