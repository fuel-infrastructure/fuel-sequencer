package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/math"
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
	"github.com/fuel-infrastructure/fuel-sequencer/testutil/sample"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

func CreateNSlashEntry(keeper keeper.Keeper, ctx context.Context, n int) []types.SlashEntry {
	slashEntries := make([]types.SlashEntry, n)
	for i := range slashEntries {
		slashEntries[i].ValidatorAddress = sample.AccAddress()
		slashEntries[i].DelegatorAddress = sample.AccAddress()
		slashEntries[i].DelegatorSlashAmount = math.OneInt().Add(math.NewInt(int64(i)))
		slashEntries[i].DelegatorUnbondingBalance = math.NewInt(10).Add(math.NewInt(int64(i)))
		slashEntries[i].DelegatorBondedBalance = math.NewInt(20).Add(math.NewInt(int64(i)))
	}

	return slashEntries
}

func CreateNSlashReport(keeper keeper.Keeper, ctx context.Context, n int) []types.SlashReport {
	slashReports := make([]types.SlashReport, n)
	for i := range slashReports {
		slashReports[i].Height = uint64(i)
		slashReports[i].Entries = CreateNSlashEntry(keeper, ctx, i)
		keeper.SetSlashReport(ctx, slashReports[i])
	}

	return slashReports
}

func ReportsKeeper(t testing.TB) (keeper.Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)

	k := keeper.NewKeeper(
		cdc,
		runtime.NewKVStoreService(storeKey),
		log.NewNopLogger(),
		authority.String(),
	)

	ctx := sdk.NewContext(stateStore, cmtproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	//nolint:errcheck
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
