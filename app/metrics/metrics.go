package metrics

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/hashicorp/go-metrics"
)

func ObserveInjectedTransactionAtAnteHandler(goCtx context.Context, tx sdk.Tx) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		for _, msg := range tx.GetMsgs() {
			telemetry.IncrCounterWithLabels(
				append(utils.KeysAnteHandler, "injected", "msg"),
				1,
				[]metrics.Label{
					telemetry.NewLabel("type", sdk.MsgTypeURL(msg)),
				})
		}
	})
}
