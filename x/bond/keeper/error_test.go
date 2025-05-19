package keeper_test

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestKeeperErrors(t *testing.T) {
	// Create test dependencies
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := runtime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()

	ctrl := gomock.NewController(t)
	accountKeeper := testutil.NewMockAccountKeeper(ctrl)
	bankKeeper := testutil.NewMockBankKeeper(ctrl)

	// Test cases
	tests := []struct {
		name      string
		authority string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "valid authority",
			authority: "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4",
			expErr:    false,
		},
		{
			name:      "empty authority",
			authority: "",
			expErr:    true,
			expErrMsg: "invalid authority address: ",
		},
		{
			name:      "invalid authority",
			authority: "invalid",
			expErr:    true,
			expErrMsg: "invalid authority address: invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expErr {
				require.PanicsWithValue(t, tc.expErrMsg, func() {
					keeper.NewKeeper(cdc, storeService, logger, tc.authority, accountKeeper, bankKeeper)
				})
			} else {
				require.NotPanics(t, func() {
					keeper.NewKeeper(cdc, storeService, logger, tc.authority, accountKeeper, bankKeeper)
				})
			}
		})
	}
}
