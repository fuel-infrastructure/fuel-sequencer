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
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
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
	accountKeeper := mock.NewMockAccountKeeper(t)
	bankKeeper := mock.NewMockBankKeeper(t)

	// Test cases
	tests := []struct {
		name      string
		authority string
		expErr    bool
		expErrMsg string
	}{
		{
			name:      "valid authority",
			authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expErr:    false,
		},
		{
			name:      "empty authority",
			authority: "",
			expErr:    false, // Empty authority is allowed (will be set to governance module account)
		},
		{
			name:      "invalid authority",
			authority: "invalid",
			expErr:    true,
			expErrMsg: "decoding bech32 failed: invalid bech32 string length 7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expErr {
				require.PanicsWithError(t, tc.expErrMsg, func() {
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
