package keeper_test

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"

	testkeeper "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
)

const (
	testAuthority = "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4"
)

func TestMsgBurnCoins(t *testing.T) {
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
		setup  func(*testing.T, *gomock.Controller) (*keeper.Keeper, context.Context)
	}{
		{
			name: "successful burn single denom",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  testCoins,
			},
			expErr: false,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, context.Context) {
				k, ctx, _, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)
				mockAccountKeeper.EXPECT().AddressCodec().Return(sdkAddressCodec.NewBech32Codec("fuelsequencer")).Times(1)
				mockBankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), bondtypes.ModuleName, testCoins).Return(nil)
				mockBankKeeper.EXPECT().BurnCoins(gomock.Any(), bondtypes.ModuleName, testCoins).Return(nil)
				return &k, ctx
			},
		},
		{
			name: "successful burn multiple denoms",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  multiDenomCoins,
			},
			expErr: false,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, context.Context) {
				k, ctx, _, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)
				mockAccountKeeper.EXPECT().AddressCodec().Return(sdkAddressCodec.NewBech32Codec("fuelsequencer")).Times(1)
				mockBankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), bondtypes.ModuleName, multiDenomCoins).Return(nil)
				mockBankKeeper.EXPECT().BurnCoins(gomock.Any(), bondtypes.ModuleName, multiDenomCoins).Return(nil)
				return &k, ctx
			},
		},
		{
			name: "unauthorized sender",
			msg: &bondtypes.MsgBurnCoins{
				Sender: "fuelsequencer1invalid",
				Coins:  testCoins,
			},
			expErr: true,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, context.Context) {
				k, ctx, _, mockAccountKeeper, _ := testkeeper.BondKeeperWithDependencies(t)
				mockAccountKeeper.EXPECT().AddressCodec().Return(sdkAddressCodec.NewBech32Codec("fuelsequencer")).Times(1)
				return &k, ctx
			},
		},
		{
			name: "zero coins",
			msg: &bondtypes.MsgBurnCoins{
				Sender: testAuthority,
				Coins:  sdk.NewCoins(),
			},
			expErr: true,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, context.Context) {
				k, ctx := testkeeper.BondKeeper(t)
				return &k, ctx
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			k, ctx := tc.setup(t, ctrl)
			msgServer := keeper.NewMsgServerImpl(*k)
			_, err := msgServer.BurnCoins(ctx, tc.msg)

			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
