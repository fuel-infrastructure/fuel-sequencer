package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func GetMintAmountFromMintEvent(ctx sdk.Context, k Keeper, event sdk.Event) (mint, burn math.Int) {
	return getMintAmountFromMintEvent(ctx, k, event)
}

func GetBurnAmountFromSlashEvent(ctx sdk.Context, k Keeper, event sdk.Event) (mint, burn math.Int) {
	return getBurnAmountFromSlashEvent(ctx, k, event)
}

func GetBurnAmountFromCoinBurnEvent(ctx sdk.Context, k Keeper, event sdk.Event) (mint, burn math.Int) {
	return getBurnAmountFromCoinBurnEvent(ctx, k, event)
}
