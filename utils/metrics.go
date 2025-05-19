package utils

import (
	"context"
	"math/big"
	"strconv"

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
	KeysStore       = append(KeysSequencer, "store")

	// Coin amounts are divided by this value to get the decimal representation, assuming 9 decimal places
	scale = new(big.Float).SetFloat64(1e9)
)

// safeSetMetric helps us use the telemetry package in a safer and more effective way by protecting against panics and
// checking that we only set metrics at the Finalize mode, to reflect actual changes to state, as much as possible.
func safeSetMetric(goCtx context.Context, setMetric func(ctx sdk.Context)) {
	defer func() {
		_ = recover() // recover from panics without running any other logic
	}()
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !telemetry.IsTelemetryEnabled() || ctx.ExecMode() != sdk.ExecModeFinalize {
		return
	}
	setMetric(ctx)
}

func SafeSetMetric(goCtx context.Context, setMetric func(ctx sdk.Context)) {
	go safeSetMetric(goCtx, setMetric)
}

// ScaleCoinAmount converts a coin amount to a float32 by converting the
// amount to the decimal representation, assuming 9 decimal places.
func ScaleCoinAmount(amount sdkmath.Int) float32 {
	amountFloat := new(big.Float).SetInt(amount.BigInt())
	amountScaled, _ := new(big.Float).Quo(amountFloat, scale).Float32()
	return amountScaled
}

// Int64ToString converts an int64 to a string
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
