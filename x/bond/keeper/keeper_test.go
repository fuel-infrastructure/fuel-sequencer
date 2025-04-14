package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdkstore "cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	testifymock "github.com/stretchr/testify/mock"
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
	bankKeeper := mock.NewMockBankKeeper(t)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
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
	bankKeeper := mock.NewMockBankKeeper(t)

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		"cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu", // Test authority
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

func TestKeeper_BurnCoins(t *testing.T) {
	k, _ := setupKeeper(t)
	storeKey := storetypes.NewKVStoreKey(bondtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Test cases
	testCases := []struct {
		name          string
		sender        sdk.AccAddress
		coins         sdk.Coins
		bankKeeperErr error
		expectErr     bool
	}{
		{
			name:          "successful burn",
			sender:        sdk.AccAddress("test1"),
			coins:         sdk.NewCoins(sdk.NewCoin("test", sdkmath.NewInt(100))),
			bankKeeperErr: nil,
			expectErr:     false,
		},
		{
			name:          "invalid coins",
			sender:        sdk.AccAddress("test1"),
			coins:         sdk.Coins{sdk.Coin{Denom: "test", Amount: sdkmath.NewInt(-100)}},
			bankKeeperErr: nil,
			expectErr:     true,
		},
		{
			name:          "zero coins",
			sender:        sdk.AccAddress("test1"),
			coins:         sdk.NewCoins(),
			bankKeeperErr: nil,
			expectErr:     true,
		},
		{
			name:          "bank keeper error",
			sender:        sdk.AccAddress("test1"),
			coins:         sdk.NewCoins(sdk.NewCoin("test", sdkmath.NewInt(100))),
			bankKeeperErr: fmt.Errorf("bank error"),
			expectErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset mock for each test case
			mockBankKeeper := mock.NewMockBankKeeper(t)
			k := keeper.NewKeeper(
				k.GetCodec(),
				k.GetStoreService(),
				k.Logger(),
				k.GetAuthority(),
				mockBankKeeper,
			)

			// Setup mock expectations
			mockBankKeeper.On("SendCoinsFromAccountToModule", testifymock.Anything, tc.sender, bondtypes.ModuleName, tc.coins).Return(tc.bankKeeperErr)
			if tc.bankKeeperErr == nil {
				mockBankKeeper.On("BurnCoins", testifymock.Anything, bondtypes.ModuleName, tc.coins).Return(nil)
			}

			// Execute burn coins
			err := k.BurnCoins(ctx, tc.sender, tc.coins)

			// Verify results
			if tc.expectErr {
				require.Error(t, err)
				if tc.bankKeeperErr != nil {
					require.Contains(t, err.Error(), tc.bankKeeperErr.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
