package mint

import (
	"context"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	minttypes "cosmossdk.io/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
)

func ProvideMintFn(bankKeeper types.MintBankKeeper, bridgeKeeper types.BridgeKeeper) minttypes.MintFn {
	return func(ctx context.Context, env appmodule.Environment, minter *minttypes.Minter, epochID string, epochNumber int64) error {

		resp, err := env.QueryRouterService.Invoke(ctx, &minttypes.QueryParamsRequest{})
		if err != nil {
			return err
		}
		mintParams, ok := resp.(*minttypes.QueryParamsResponse)
		if !ok {
			return fmt.Errorf("unexpected response type: %T", resp)
		}
		params := mintParams.Params

		// Use BridgeDenomTotalSupply as the supply for the NextAnnualProvisions calculation.
		// This replaces the supply from the StakingTokenSupply call to the staking module.
		totalSupply := bridgeKeeper.GetParams(ctx).BridgeDenomTotalSupply

		// We want inflation to be configurable via the InflationMin and InflationMax params. To remove any doubts
		// as to which inflation value will be picked, we require that the InflationMin and InflationMax are equal.
		if params.InflationMin.Equal(params.InflationMax) {
			minter.Inflation = params.InflationMin
		}

		minter.AnnualProvisions = minter.NextAnnualProvisions(params, totalSupply)

		// mint coins, update supply
		mintedCoin := minter.BlockProvision(params)
		mintedCoins := sdk.NewCoins(mintedCoin)

		err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, mintedCoins)
		if err != nil {
			return err
		}
		metrics.MintCoins(ctx, mintedCoin)

		// send the minted coins to the fee collector account
		if err = bankKeeper.SendCoinsFromModuleToModule(ctx, minttypes.ModuleName, authtypes.FeeCollectorName, mintedCoins); err != nil {
			return err
		}

		err = env.EventService.EventManager(ctx).EmitKV(
			minttypes.EventTypeMint,
			event.NewAttribute(minttypes.AttributeKeyInflation, minter.Inflation.String()),
			event.NewAttribute(minttypes.AttributeKeyAnnualProvisions, minter.AnnualProvisions.String()),
			event.NewAttribute(sdk.AttributeKeyAmount, mintedCoin.Amount.String()),
		)
		if err != nil {
			return err
		}

		return nil
	}
}
