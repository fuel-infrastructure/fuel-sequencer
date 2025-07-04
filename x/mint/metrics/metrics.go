package metrics

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func MintCoins(ctx context.Context, mint sdk.Coin) {
	utils.SafeSetMetric(ctx, func(ctx sdk.Context) {
		telemetry.SetGauge(utils.ScaleCoinAmount(mint.Amount), append(utils.KeysBeginBlock, "minted", "tokens")...)
	})
}
