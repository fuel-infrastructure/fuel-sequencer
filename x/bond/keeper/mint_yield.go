package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// MintYield mints the yield amount to the yield recipient if the current time matches the yield time.
func (k Keeper) MintYield(ctx context.Context) error {
	params := k.GetParams(ctx)
	if params.YieldRecipient == "" || params.YieldTime == nil || params.YieldAmount.IsZero() {
		return nil
	}

	// Check if yield has already been minted
	if k.GetYieldMintHeight(ctx) > 0 {
		return nil
	}

	// Check if it's time to mint yield
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	if !sdkCtx.BlockTime().Equal(*params.YieldTime) {
		return nil
	}

	// Validate recipient address
	if _, err := sdk.AccAddressFromBech32(params.YieldRecipient); err != nil {
		return err
	}

	// Mint yield to recipient
	recipient, err := sdk.AccAddressFromBech32(params.YieldRecipient)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, params.YieldAmount))
	if err := k.bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, recipient, coins); err != nil {
		return err
	}

	// Set yield mint height
	return k.SetYieldMintHeight(ctx, sdkCtx.BlockHeight())
}
