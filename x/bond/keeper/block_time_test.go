package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	bondtestutil "github.com/fuel-infrastructure/fuel-sequencer/x/bond/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestBlockTimeAfterYieldTime(t *testing.T) {
	yieldAmount := sdkmath.NewInt(1000000)

	keeper, ctx, _, accountKeeper, bankKeeper, bridgeKeeper := keepertest.BondKeeperWithDependencies(t)

	// Set yield time to a minute in the future relative to wall clock
	yieldTime := time.Now().Add(time.Minute)
	params := types.NewParams(keeper.GetAuthority(), &yieldTime, yieldAmount)
	require.NoError(t, keeper.SetParams(ctx, params))

	// Set up test context with block time after yield time
	testCtx := ctx.WithBlockTime(yieldTime.Add(time.Second))
	accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
	bridgeKeeper.EXPECT().GetParams(gomock.Any()).Return(bridgetypes.Params{BridgeDenom: "ufuel"}).AnyTimes()
	bankKeeper.EXPECT().MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)
	bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			gomock.Any(),
			sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
		).
		Return(nil)

	// Execute MintYield
	err := keeper.MintYield(testCtx)
	assert.NoError(t, err)

	// Check yield mint height
	height := keeper.GetYieldMintHeight(testCtx)
	require.Equal(t, testCtx.BlockHeight(), height)

	// Verify state persistence
	newCtx := ctx.WithBlockTime(yieldTime)
	persistedHeight := keeper.GetYieldMintHeight(newCtx)
	require.Equal(t, height, persistedHeight)
}

func TestBlockTimeExactlyAtYieldTime(t *testing.T) {
	yieldAmount := sdkmath.NewInt(1000000)

	keeper, ctx, _, accountKeeper, bankKeeper, bridgeKeeper := keepertest.BondKeeperWithDependencies(t)

	// Set yield time to a minute in the future relative to wall clock
	yieldTime := time.Now().Add(time.Minute)
	params := types.NewParams(keeper.GetAuthority(), &yieldTime, yieldAmount)
	require.NoError(t, keeper.SetParams(ctx, params))

	// Set up test context with block time exactly at yield time
	testCtx := ctx.WithBlockTime(yieldTime)
	accountKeeper.EXPECT().AddressCodec().Return(bondtestutil.MockAddressCodec{}).AnyTimes()
	bridgeKeeper.EXPECT().GetParams(gomock.Any()).Return(bridgetypes.Params{BridgeDenom: "ufuel"}).AnyTimes()
	bankKeeper.EXPECT().MintCoins(gomock.Any(), types.ModuleName, sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount))).Return(nil)
	bankKeeper.EXPECT().
		SendCoinsFromModuleToAccount(
			gomock.Any(),
			types.ModuleName,
			gomock.Any(),
			sdk.NewCoins(sdk.NewCoin("ufuel", yieldAmount)),
		).
		Return(nil)

	// Execute MintYield
	err := keeper.MintYield(testCtx)
	assert.NoError(t, err)

	// Check yield mint height
	height := keeper.GetYieldMintHeight(testCtx)
	require.Equal(t, testCtx.BlockHeight(), height)

	// Verify state persistence
	newCtx := ctx.WithBlockTime(yieldTime)
	persistedHeight := keeper.GetYieldMintHeight(newCtx)
	require.Equal(t, height, persistedHeight)
}
