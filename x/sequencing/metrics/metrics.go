package metrics

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func ObserveTotalBlobsPosted(goCtx context.Context, topic []byte) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "total", "blobs", "posted", string(topic))...)
	})
}

func ObserveTotalBlobsPostedSize(goCtx context.Context, topic []byte, size int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(
			float32(size),
			append(utils.KeysTxMsg, "total", "blobs", "posted", "size", string(topic))...,
		)
	})
}
