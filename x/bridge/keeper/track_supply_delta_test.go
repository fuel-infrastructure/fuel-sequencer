package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestUpdateSupplyDeltaFromEventManager(t *testing.T) {

	bridgeDenom := "token"
	hundred := math.NewInt(100)
	twoHundred := math.NewInt(200)
	hundredTokens := sdk.NewCoin(bridgeDenom, hundred)
	twoHundredTokens := sdk.NewCoin(bridgeDenom, twoHundred)

	govAddress := authtypes.NewModuleAddress(govtypes.ModuleName)

	k, ctx := keepertest.BridgeKeeper(t)

	// Set of valid events.
	validMintEvent := sdk.NewEvent(
		types.EventTypeMint,
		sdk.NewAttribute(sdk.AttributeKeyAmount, hundred.String()),
	)
	validSlashEvent := sdk.NewEvent(
		slashingtypes.EventTypeSlash,
		sdk.NewAttribute(slashingtypes.AttributeKeyBurnedCoins, hundredTokens.String()),
	)
	validBurnEvent := banktypes.NewCoinBurnEvent(govAddress, sdk.NewCoins(twoHundredTokens))

	// Initialise bridge denom.
	err := k.SetParams(ctx, bridgetypes.Params{
		BridgeDenom: bridgeDenom,
	})
	require.NoError(t, err)

	// Initialise supply delta info.
	k.SetSupplyDeltaInfo(ctx, bridgetypes.SupplyDeltaInfo{
		Mint: math.NewInt(1000),
		Burn: math.NewInt(1000),
	})

	// Emit and process events.
	ctx.EventManager().EmitEvents(sdk.Events{
		validMintEvent,
		validSlashEvent,
		validBurnEvent,
	})
	k.UpdateSupplyDeltaFromEventManager(ctx, "source")

	// Confirm updated supply delta info.
	updatedSupplyDeltaInfo, found := k.GetSupplyDeltaInfo(ctx)
	require.True(t, found)
	require.True(t, updatedSupplyDeltaInfo.Mint.Equal(math.NewInt(1100))) // 1000 + 100
	require.True(t, updatedSupplyDeltaInfo.Burn.Equal(math.NewInt(1300))) // 1000 + 200 + 100
}

func TestGetMintAmountFromMintEvent(t *testing.T) {

	hundred := math.NewInt(100)

	testCases := []struct {
		name       string
		event      sdk.Event
		expectMint math.Int
	}{
		{
			name: "valid mint event",
			event: sdk.NewEvent(
				types.EventTypeMint,
				sdk.NewAttribute(sdk.AttributeKeyAmount, hundred.String()),
			),
			expectMint: hundred,
		},
		{
			name:       "mint event without amount",
			event:      sdk.NewEvent(types.EventTypeMint),
			expectMint: math.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := keepertest.BridgeKeeper(t)

			mint, burn := keeper.GetMintAmountFromMintEvent(ctx, k, tc.event)
			require.True(t, mint.Equal(tc.expectMint))
			require.True(t, burn.IsZero())
		})
	}
}

func TestGetBurnAmountFromSlashEvent(t *testing.T) {

	bridgeDenom := "token"
	AnotherDenom := "token2"

	params := bridgetypes.Params{
		BridgeDenom: bridgeDenom,
	}

	hundred := math.NewInt(100)
	hundredTokens := sdk.NewCoin(bridgeDenom, hundred)
	hundredOtherTokens := sdk.NewCoin(AnotherDenom, hundred)

	testCases := []struct {
		name       string
		event      sdk.Event
		expectBurn math.Int
	}{
		{
			name: "valid slash event",
			event: sdk.NewEvent(
				slashingtypes.EventTypeSlash,
				sdk.NewAttribute(slashingtypes.AttributeKeyBurnedCoins, hundredTokens.String()),
			),
			expectBurn: hundred,
		},
		{
			name: "non-bridge denoms",
			event: sdk.NewEvent(
				slashingtypes.EventTypeSlash,
				sdk.NewAttribute(slashingtypes.AttributeKeyBurnedCoins, hundredOtherTokens.String()),
			),
			expectBurn: math.ZeroInt(),
		},
		{
			name:       "slash event without amount",
			event:      sdk.NewEvent(slashingtypes.EventTypeSlash),
			expectBurn: math.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := keepertest.BridgeKeeper(t)

			require.NoError(t, k.SetParams(ctx, params))

			mint, burn := keeper.GetBurnAmountFromSlashEvent(ctx, k, tc.event)
			require.True(t, mint.IsZero())
			require.True(t, burn.Equal(tc.expectBurn))
		})
	}
}

func TestGetBurnAmountFromCoinBurnEvent(t *testing.T) {

	bridgeDenom := "token"
	AnotherDenom := "token2"

	params := bridgetypes.Params{
		BridgeDenom: bridgeDenom,
	}

	hundred := math.NewInt(100)
	hundredTokens := sdk.NewCoins(sdk.NewCoin(bridgeDenom, hundred))
	hundredOtherTokens := sdk.NewCoins(sdk.NewCoin(AnotherDenom, hundred))

	govAddress := authtypes.NewModuleAddress(govtypes.ModuleName)
	anotherAddress := authtypes.NewModuleAddress("anotherAddress")

	testCases := []struct {
		name       string
		event      sdk.Event
		expectBurn math.Int
	}{
		{
			name:       "valid burn event",
			event:      banktypes.NewCoinBurnEvent(govAddress, hundredTokens),
			expectBurn: hundred,
		},
		{
			name:       "non-governance address",
			event:      banktypes.NewCoinBurnEvent(anotherAddress, hundredTokens),
			expectBurn: math.ZeroInt(),
		},
		{
			name:       "non-bridge denoms",
			event:      banktypes.NewCoinBurnEvent(anotherAddress, hundredOtherTokens),
			expectBurn: math.ZeroInt(),
		},
		{
			name:       "burn event without amount",
			event:      banktypes.NewCoinBurnEvent(govAddress, sdk.NewCoins()),
			expectBurn: math.ZeroInt(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			k, ctx := keepertest.BridgeKeeper(t)

			require.NoError(t, k.SetParams(ctx, params))

			mint, burn := keeper.GetBurnAmountFromCoinBurnEvent(ctx, k, tc.event)
			require.True(t, mint.IsZero())
			require.True(t, burn.Equal(tc.expectBurn))
		})
	}
}
