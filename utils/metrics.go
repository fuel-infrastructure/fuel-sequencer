package utils

import (
	"context"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
