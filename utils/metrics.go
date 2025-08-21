package utils

import (
	"context"
	"math/big"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	KeysSequencer   = []string{"sequencer"}
	KeysAnteHandler = append(KeysSequencer, "ante", "handler")
	KeysBeginBlock  = append(KeysSequencer, "begin", "block")
	KeysEndBlock    = append(KeysSequencer, "end", "block")
	KeysTxMsg       = append(KeysSequencer, "tx", "msg")
	KeysBlobhub     = append(KeysSequencer, "blobhub")
	KeysBlobpool    = append(KeysSequencer, "blobpool")
	KeysStore       = append(KeysSequencer, "store")

	// Coin amounts are divided by this value to get the decimal representation, assuming 9 decimal places
	scale = new(big.Float).SetFloat64(1e9)
)

// SafeSetMetric helps us use the telemetry package in a safer and more effective way by protecting against panics.
func SafeSetMetric(setMetric func()) {
	go func() {
		defer func() {
			_ = recover() // recover from panics without running any other logic
		}()
		if !telemetry.IsTelemetryEnabled() {
			return
		}
		setMetric()
	}()
}

// SafeSetFinalizedMetric helps us use the telemetry package in a safer and more effective way by protecting against panics and
// checking that we only set metrics at the Finalize mode, to reflect actual changes to state, as much as possible.
func SafeSetFinalizedMetric(goCtx context.Context, setMetric func(ctx sdk.Context)) {
	go func() {
		defer func() {
			_ = recover() // recover from panics without running any other logic
		}()
		ctx := sdk.UnwrapSDKContext(goCtx)
		if !telemetry.IsTelemetryEnabled() || ctx.ExecMode() != sdk.ExecModeFinalize {
			return
		}
		setMetric(ctx)
	}()
}

// ScaleCoinAmount converts a coin amount to a float32 by converting the
// amount to the decimal representation, assuming 9 decimal places.
func ScaleCoinAmount(amount sdkmath.Int) float32 {
	amountFloat := new(big.Float).SetInt(amount.BigInt())
	amountScaled, _ := new(big.Float).Quo(amountFloat, scale).Float32()
	return amountScaled
}
