package metrics

import (
	"context"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/hashicorp/go-metrics"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func ObserveTotalBlobsPosted(goCtx context.Context, topic []byte) {
	utils.SafeSetFinalizedMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "total", "blobs", "posted"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("topic", hex.EncodeToString(topic)),
			},
		)
	})
}

func ObserveTotalBlobsPostedSize(goCtx context.Context, topic []byte, size int) {
	utils.SafeSetFinalizedMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "total", "blobs", "posted", "size"),
			float32(size),
			[]metrics.Label{
				telemetry.NewLabel("topic", hex.EncodeToString(topic)),
			},
		)
	})
}
