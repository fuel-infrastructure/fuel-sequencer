package metrics

import (
	"context"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/hashicorp/go-metrics"
)

// ObserveBurnCoins tracks when coins are burned through the bond module
func ObserveBurnCoins(goCtx context.Context, coins sdk.Coins) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		for _, coin := range coins {
			telemetry.IncrCounterWithLabels(
				append(utils.KeysTxMsg, "burn", "coins"),
				1,
				[]metrics.Label{
					telemetry.NewLabel("denom", coin.Denom),
					telemetry.NewLabel("amount", coin.Amount.String()),
				},
			)
		}
	})
}

// SetInflation tracks changes to the bond module's inflation parameter
func SetInflation(goCtx context.Context, inflation sdkmath.LegacyDec) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		// Convert LegacyDec to float32 for telemetry
		inflationFloat, _ := inflation.Float64()
		telemetry.SetGauge(float32(inflationFloat), append(utils.KeysStore, "inflation")...)
	})
}
