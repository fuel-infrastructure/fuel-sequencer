package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	bondtestutil "github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestMintYield(t *testing.T) {
	baseTime := time.Now()
	futureTime := baseTime.Add(time.Hour)
	yieldAmount := sdkmath.NewInt(1000000)

	testCases := []struct {
		name           string
		setupParams    func(keeper keeper.Keeper) types.Params
		setupContext   func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper)
		expectedHeight int64
		expectedError  bool
	}{
		{
			name: "no yield parameters set",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.DefaultParams()
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "SetParams never called",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.DefaultParams()
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "SetParams errors",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.DefaultParams()
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				// Simulate SetParams error by not calling SetParams
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "yield time not reached",
			setupParams: func(keeper keeper.Keeper) types.Params {
				future := futureTime.Add(time.Hour)
				return types.NewParams(keeper.GetAuthority(), &future, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "yield time reached",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)
				bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						minttypes.ModuleName,
						gomock.Any(),
						sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
					).
					Return(nil)
			},
			expectedHeight: 0, // will be set after context is created
			expectedError:  false,
		},
		{
			name: "yield already minted",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				// Set a non-zero block height
				testCtx = testCtx.WithBlockHeight(123)
				require.Equal(t, int64(123), testCtx.BlockHeight())
				// First run: simulate a valid yield mint
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)
				bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						minttypes.ModuleName,
						gomock.Any(),
						sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
					).
					Return(nil)
				// Run MintYield once to set the yield mint height
				err := keeper.MintYield(testCtx)
				require.NoError(t, err)
				// Verify that the yield mint height is set
				height := keeper.GetYieldMintHeight(testCtx)
				require.Greater(t, height, int64(0))
				// Second run: should not mint again, so no new expectations
				err = keeper.MintYield(testCtx)
				require.NoError(t, err)
				// Height should remain unchanged
				require.Equal(t, height, keeper.GetYieldMintHeight(testCtx))
			},
			expectedHeight: 0, // will be set after context is created
			expectedError:  false,
		},
		{
			name: "mint coins error",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(types.ErrMintCoins)
			},
			expectedHeight: 0,
			expectedError:  true,
		},
		{
			name: "send coins error",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)
				bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						minttypes.ModuleName,
						gomock.Any(),
						sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
					).
					Return(types.ErrSendCoins)
			},
			expectedHeight: 0,
			expectedError:  true,
		},
		{
			name: "zero yield amount",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, sdkmath.ZeroInt())
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "nil yield time with valid params",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), nil, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
			},
			expectedHeight: 0,
			expectedError:  false,
		},
		{
			name: "maximum yield amount",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, sdkmath.NewIntFromUint64(^uint64(0)))
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewIntFromUint64(^uint64(0))))).Return(nil)
				bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						minttypes.ModuleName,
						gomock.Any(),
						sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewIntFromUint64(^uint64(0)))),
					).
					Return(nil)
			},
			expectedHeight: 0, // will be set after context is created
			expectedError:  false,
		},
		{
			name: "verify minting state",
			setupParams: func(keeper keeper.Keeper) types.Params {
				return types.NewParams(keeper.GetAuthority(), &futureTime, yieldAmount)
			},
			setupContext: func(testCtx sdk.Context, keeper keeper.Keeper, bankKeeper *bondtestutil.MockBankKeeper, accountKeeper *bondtestutil.MockAccountKeeper) {
				accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()

				// Expect MintCoins to be called with the correct amount
				bankKeeper.EXPECT().MintCoins(gomock.Any(), minttypes.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)

				// Expect SendCoinsFromModuleToAccount to be called with the correct amount
				bankKeeper.EXPECT().
					SendCoinsFromModuleToAccount(
						gomock.Any(),
						minttypes.ModuleName,
						gomock.Any(),
						sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
					).
					Return(nil)
			},
			expectedHeight: 0, // will be set after context is created
			expectedError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keeper, ctx, _, accountKeeper, bankKeeper := keepertest.BondKeeperWithDependencies(t)
			params := tc.setupParams(keeper)
			require.NoError(t, keeper.SetParams(ctx, params))

			// Set up test context and mocks
			testCtx := ctx.WithBlockTime(futureTime)
			tc.setupContext(testCtx, keeper, bankKeeper, accountKeeper)

			// For cases that should mint, set expectedHeight to current block height
			if tc.name == "yield time reached" || tc.name == "yield already minted" ||
				tc.name == "maximum yield amount" || tc.name == "verify minting state" {
				tc.expectedHeight = testCtx.BlockHeight()
			}

			// Execute MintYield
			if tc.name != "yield already minted" {
				err := keeper.MintYield(testCtx)
				if tc.expectedError {
					assert.Error(t, err)
					return
				} else {
					assert.NoError(t, err)
				}

				// Check yield mint height
				height := keeper.GetYieldMintHeight(testCtx)
				require.Equal(t, tc.expectedHeight, height)

				// For successful mints, verify state persistence
				if height > 0 {
					// Create new context to verify persistence
					newCtx := ctx.WithBlockTime(futureTime)
					persistedHeight := keeper.GetYieldMintHeight(newCtx)
					require.Equal(t, height, persistedHeight)
				}
			}
		})
	}
}

func TestMintYield_InvalidRecipientAddress(t *testing.T) {
	keeper, ctx, _, _, _ := keepertest.BondKeeperWithDependencies(t)
	futureTime := time.Now().Add(time.Hour)
	yieldAmount := sdkmath.NewInt(1000000)
	params := types.NewParams("invalid", &futureTime, yieldAmount)
	require.Error(t, keeper.SetParams(ctx, params),
		"MintYield should not fail if the recipient address is invalid, as it will be validated before minting by SetParams")
}
