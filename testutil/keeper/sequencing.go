package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	govtypes "cosmossdk.io/x/gov/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func SequencingKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	logger := log.NewNopLogger()

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	cdc := testutiltypes.TestCdc
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	authorityAddr, err := testutiltypes.TestAddressCdc.BytesToString(authority)
	if err != nil {
		panic(err)
	}

	kvs := runtime.NewKVStoreService(storeKey)
	k := keeper.NewKeeper(
		cdc,
		runtime.NewEnvironment(kvs, logger.With(log.ModuleKey, "x/sequencing")),
		log.NewNopLogger(),
		nil,
		authorityAddr,
	)

	ctx := sdk.NewContext(stateStore, false, log.NewNopLogger())

	// Initialize params
	//nolint:errcheck
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
