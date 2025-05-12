package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// MintYield mints the yield amount to the yield recipient if the current time matches the yield time.
func (k Keeper) MintYield(ctx context.Context) error {
	params := k.GetParams(ctx)
	k.Logger().Debug("Starting MintYield",
		"yield_recipient", params.YieldRecipient,
		"yield_time", params.YieldTime,
		"yield_amount", params.YieldAmount.String())

	if params.YieldRecipient == "" || params.YieldTime == nil || params.YieldAmount.IsZero() {
		k.Logger().Debug("Skipping MintYield - missing required parameters",
			"has_recipient", params.YieldRecipient != "",
			"has_time", params.YieldTime != nil,
			"has_amount", !params.YieldAmount.IsZero())
		return nil
	}

	// Check if yield has already been minted
	currentMintHeight := k.GetYieldMintHeight(ctx)
	k.Logger().Debug("Checking yield mint height",
		"current_mint_height", currentMintHeight)

	if currentMintHeight > 0 {
		k.Logger().Debug("Skipping MintYield - already minted at height", "mint_height", currentMintHeight)
		return nil
	}

	// Check if it's time to mint yield
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentTime := sdkCtx.BlockTime()
	k.Logger().Debug("Checking yield time",
		"current_time", currentTime,
		"yield_time", params.YieldTime,
		"time_difference", currentTime.Sub(*params.YieldTime))

	if currentTime.Before(*params.YieldTime) {
		k.Logger().Debug("Skipping MintYield - yield time not reached")
		return nil
	}

	// Mint yield to recipient
	recipient, err := k.GetAccountAsBytes(params.YieldRecipient)
	if err != nil {
		k.Logger().Error("Failed to get recipient account",
			"recipient", params.YieldRecipient,
			"error", err)
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(bridgetypes.DefaultBridgeDenom, params.YieldAmount))
	k.Logger().Info("Attempting to mint coins",
		"amount", coins.String(),
		"module", minttypes.ModuleName)

	if err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins); err != nil {
		k.Logger().Error("Failed to mint coins",
			"amount", coins.String(),
			"error", err)
		return err
	}

	k.Logger().Info("Successfully minted coins, attempting transfer",
		"amount", coins.String(),
		"recipient", params.YieldRecipient)

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, recipient, coins); err != nil {
		k.Logger().Error("Failed to send coins to recipient",
			"amount", coins.String(),
			"recipient", params.YieldRecipient,
			"error", err)
		return err
	}

	// Set yield mint height
	blockHeight := sdkCtx.BlockHeight()
	k.Logger().Info("Setting yield mint height",
		"height", blockHeight)

	if err := k.SetYieldMintHeight(ctx, blockHeight); err != nil {
		k.Logger().Error("Failed to set yield mint height",
			"height", blockHeight,
			"error", err)
		return err
	}

	k.Logger().Info("Successfully completed MintYield",
		"mint_height", blockHeight,
		"amount", coins.String(),
		"recipient", params.YieldRecipient)
	return nil
}
