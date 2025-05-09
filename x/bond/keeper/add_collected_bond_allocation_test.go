package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	testkeeper "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
)

func TestAddCollectedBondAllocation(t *testing.T) {
	// Test coins
	testCoins := sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100)))
	multiDenomCoins := sdk.NewCoins(
		sdk.NewCoin("ufuel", sdkmath.NewInt(100)),
		sdk.NewCoin("uatom", sdkmath.NewInt(50)),
	)

	// Test cases
	testCases := []struct {
		name      string
		coins     sdk.Coins
		expErr    bool
		expPanic  bool
		expErrMsg string
		setup     func(*testing.T, *gomock.Controller) (*keeper.Keeper, sdk.Context)
	}{
		{
			name:   "successful transfer single denom",
			coins:  testCoins,
			expErr: false,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, sdk.Context) {
				k, ctx, _, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)
				mockAccountKeeper.EXPECT().AddressCodec().Return(sdkAddressCodec.NewBech32Codec("fuelsequencer")).Times(1)
				mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), minttypes.ModuleName, gomock.Any(), testCoins).Return(nil)
				return &k, ctx
			},
		},
		{
			name:   "successful transfer multiple denoms",
			coins:  multiDenomCoins,
			expErr: false,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, sdk.Context) {
				k, ctx, _, mockAccountKeeper, mockBankKeeper := testkeeper.BondKeeperWithDependencies(t)
				mockAccountKeeper.EXPECT().AddressCodec().Return(sdkAddressCodec.NewBech32Codec("fuelsequencer")).Times(1)
				mockBankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), minttypes.ModuleName, gomock.Any(), multiDenomCoins).Return(nil)
				return &k, ctx
			},
		},
		{
			name:   "successful transfer zero coins",
			coins:  sdk.NewCoins(),
			expErr: false,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, sdk.Context) {
				k, ctx := testkeeper.BondKeeper(t)
				return &k, ctx
			},
		},
		{
			name:     "invalid authority address",
			coins:    testCoins,
			expPanic: true,
			setup: func(t *testing.T, ctrl *gomock.Controller) (*keeper.Keeper, sdk.Context) {
				// Create a keeper with an invalid authority address
				k, ctx, _, _, _ := testkeeper.BondKeeperFromArgsWithDependencies(t, "invalid")
				return &k, ctx
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			if tc.expPanic {
				require.Panics(t, func() {
					k, ctx := tc.setup(t, ctrl)
					_ = k.AddCollectedBondAllocation(ctx, tc.coins)
				})
				return
			}

			k, ctx := tc.setup(t, ctrl)
			err := k.AddCollectedBondAllocation(ctx, tc.coins)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
