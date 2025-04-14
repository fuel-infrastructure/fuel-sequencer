package bond_test

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

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bond "github.com/fuel-infrastructure/fuel-sequencer/x/bond/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestGenesis(t *testing.T) {
	// Create test dependencies
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	logger := log.NewNopLogger()
	accountKeeper := mock.NewMockAccountKeeper(t)
	bankKeeper := mock.NewMockBankKeeper(t)

	// Create test module
	appModule := bond.NewAppModule(
		cdc,
		keeper.NewKeeper(
			cdc,
			runtime.NewKVStoreService(storeKey),
			logger,
			authtypes.NewModuleAddress(govtypes.ModuleName).String(), // Will be set to governance module account
			accountKeeper,
			bankKeeper,
		),
		nil, // Account keeper will be set in actual app
		nil, // Bank keeper will be set in actual app
	)

	// Test module name
	require.Equal(t, "bond", appModule.Name())

	// Test module version
	require.Equal(t, uint64(1), appModule.ConsensusVersion())

	// Test module genesis
	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())
	require.NotPanics(t, func() {
		appModule.InitGenesis(ctx, cdc, cdc.MustMarshalJSON(types.DefaultGenesis()))
	})

	// Test module export
	exported := appModule.ExportGenesis(ctx, cdc)
	var exportedState types.GenesisState
	cdc.MustUnmarshalJSON(exported, &exportedState)
	require.Equal(t, types.DefaultGenesis().Params, exportedState.Params)
}
