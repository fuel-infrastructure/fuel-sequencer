package metrics

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	"github.com/hashicorp/go-metrics"
)

func ObserveMintCoins(ctx context.Context, mint sdk.Coins) {
	utils.SafeSetMetric(ctx, func(ctx sdk.Context) {
		for _, coin := range mint {
			telemetry.SetGauge(utils.ScaleCoinAmount(coin.Amount), append(utils.KeysBeginBlock, "minted", "tokens")...)
		}
	})
}

// ObserveYieldMinting tracks when yield is minted through the bond module
func ObserveYieldMinting(goCtx context.Context, amount sdk.Coins, height int64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		for _, coin := range amount {
			telemetry.IncrCounterWithLabels(
				append(utils.KeysTxMsg, "yield", "mint"),
				1,
				[]metrics.Label{
					telemetry.NewLabel("denom", coin.Denom),
					telemetry.NewLabel("amount", coin.Amount.String()),
					telemetry.NewLabel("height", utils.Int64ToString(height)),
				},
			)
		}
	})
}

// SetParamsUpdate tracks when bond module parameters are updated
func SetParamsUpdate(goCtx context.Context, params types.Params) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "params", "update"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("yield_recipient", params.YieldRecipient),
				telemetry.NewLabel("yield_time", params.YieldTime.String()),
				telemetry.NewLabel("yield_amount", params.YieldAmount.String()),
			},
		)
	})
}
