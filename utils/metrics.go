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
	KeysStore       = append(KeysSequencer, "store")

	// Coin amounts are divided by this value to get the decimal representation
	scale = new(big.Float).SetFloat64(1e18)
)

// safeSetMetric helps us use the telemetry package in a safer and more effective way by protecting against panics and
// checking that we only set metrics at the Finalize mode, to reflect actual changes to state, as much as possible.
func safeSetMetric(goCtx context.Context, setMetric func(ctx sdk.Context)) {
	defer func() {
		if r := recover(); r != nil {
			// Recover from panics
		}
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

func ScaleCoinAmount(amount sdkmath.Int) float32 {
	amountFloat := new(big.Float).SetInt(amount.BigInt())
	amountScaled, _ := new(big.Float).Quo(amountFloat, scale).Float32()
	return amountScaled
}
