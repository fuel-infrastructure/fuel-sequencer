package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	sdkstore "cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	testmock "github.com/fuel-infrastructure/fuel-sequencer/testutil/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

const (
	testAuthority = "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"
)

func setupMsgBurnCoins(t *testing.T) (keeper.Keeper, context.Context, *testmock.MockAccountKeeper, *testmock.MockBankKeeper) {
	storeKey := storetypes.NewKVStoreKey(bondtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := sdkstore.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	storeService := runtime.NewKVStoreService(storeKey)
	logger := log.NewNopLogger()

	mockAccountKeeper := testmock.NewMockAccountKeeper(t)
	mockBankKeeper := testmock.NewMockBankKeeper(t)

	// Setup AddressCodec mock
	mockAccountKeeper.On("AddressCodec").Return(sdkAddressCodec.NewBech32Codec("cosmos"))

	k := keeper.NewKeeper(
		cdc,
		storeService,
		logger,
		testAuthority,
		mockAccountKeeper,
		mockBankKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, logger)
	return k, ctx, mockAccountKeeper, mockBankKeeper
}

func TestMsgBurnCoins(t *testing.T) {
	k, ctx, _, mockBankKeeper := setupMsgBurnCoins(t)

	// Test coins
	testCoins := sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100)))
	multiDenomCoins := sdk.NewCoins(
		sdk.NewCoin("ufuel", sdkmath.NewInt(100)),
		sdk.NewCoin("uatom", sdkmath.NewInt(50)),
	)

	// Test cases
	testCases := []struct {
		name   string
		msg    *bondtypes.MsgBurnCoins
		expErr bool
		setup  func()
	}{
		{
			name: "successful burn single denom",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  testCoins,
			},
			expErr: false,
			setup: func() {
				mockBankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, mock.Anything, bondtypes.ModuleName, testCoins).Return(nil)
				mockBankKeeper.On("BurnCoins", mock.Anything, bondtypes.ModuleName, testCoins).Return(nil)
			},
		},
		{
			name: "successful burn multiple denoms",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  multiDenomCoins,
			},
			expErr: false,
			setup: func() {
				mockBankKeeper.On("SendCoinsFromAccountToModule", mock.Anything, mock.Anything, bondtypes.ModuleName, multiDenomCoins).Return(nil)
				mockBankKeeper.On("BurnCoins", mock.Anything, bondtypes.ModuleName, multiDenomCoins).Return(nil)
			},
		},
		{
			name: "unauthorized sender",
			msg: &bondtypes.MsgBurnCoins{
				Sender: "cosmos1invalid",
				Coins:  testCoins,
			},
			expErr: true,
			setup:  func() {},
		},
		{
			name: "zero coins",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  sdk.NewCoins(),
			},
			expErr: true,
			setup:  func() {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockBankKeeper.ExpectedCalls = nil // Reset mock expectations for each test case
			tc.setup()
			msgServer := keeper.NewMsgServerImpl(k)
			_, err := msgServer.BurnCoins(ctx, tc.msg)

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				mockBankKeeper.AssertExpectations(t)
			}
		})
	}
}
