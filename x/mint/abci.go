package mint

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
)

// BeginBlocker was copied from https://github.com/cosmos/cosmos-sdk/blob/v0.50.6/x/mint/abci.go.
// It is almost identical to the original, but uses BridgeDenomTotalSupply from the bridge module
// instead of getting the StakingTokenSupply from the Staking module.
// tokens being minted by
func BeginBlocker(
	ctx context.Context,
	k mintkeeper.Keeper,
	bk types.BridgeKeeper,
	ic minttypes.InflationCalculationFn,
) error {
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

	bridgeParams := bk.GetParams(ctx)

	// Ensure that the denom that we will be minting matches the BridgeDenom.
	if bridgeParams.BridgeDenom != params.MintDenom {
		return fmt.Errorf("mismatching bridge and mint denoms: %s != %s", bridgeParams.BridgeDenom, params.MintDenom)
	}

	// Use BridgeDenomTotalSupply as the supply for the NextAnnualProvisions calculation.
	totalSupply := bridgeParams.BridgeDenomTotalSupply

	bondedRatio, err := k.BondedRatio(ctx)
	if err != nil {
		return err
	}

	minter.Inflation = ic(ctx, minter, params, bondedRatio)
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

	// send the minted coins to the fee collector account
	err = k.AddCollectedFees(ctx, mintedCoins)
	if err != nil {
		return err
	}

	if mintedCoin.Amount.IsInt64() {
		defer telemetry.ModuleSetGauge(minttypes.ModuleName, float32(mintedCoin.Amount.Int64()), "minted_tokens")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent(
			minttypes.EventTypeMint,
			sdk.NewAttribute(minttypes.AttributeKeyBondedRatio, bondedRatio.String()),
			sdk.NewAttribute(minttypes.AttributeKeyInflation, minter.Inflation.String()),
			sdk.NewAttribute(minttypes.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		),
	)

	return nil
}
