package metrics

import (
	"context"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

func ObserveTotalBlobsPosted(goCtx context.Context, topic []byte) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "total", "blobs", "posted", hex.EncodeToString(topic))...)
	})
}

func ObserveTotalBlobsPostedSize(goCtx context.Context, topic []byte, size int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(
			float32(size),
			append(utils.KeysTxMsg, "total", "blobs", "posted", "size", hex.EncodeToString(topic))...,
		)
	})
}
