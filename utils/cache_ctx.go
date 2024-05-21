package utils

// This file was imported from https://github.com/osmosis-labs/osmosis/blob/v14.0.0/osmoutils/cache_ctx.go

import (
	"errors"
	"runtime"
	"runtime/debug"

	"cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ApplyFuncIfNoErrorAndNoPanic lets you run the function f, but if there's an error or panic
// drop the state machine change and log the error.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
func ApplyFuncIfNoErrorAndNoPanic(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	// Add a panic safeguard
	defer func() {
		if recoveryError := recover(); recoveryError != nil {
			PrintPanicRecoveryError(ctx, recoveryError)
			err = errors.New("panic occurred during execution")
		}
	}()
	return ApplyFuncIfNoError(ctx, f)
}

// ApplyFuncIfNoError lets you run the function f, but if there's an error
// drop the state machine change and log the error. Panics are not handled.
// If there is no error, proceeds as normal (but with some slowdown due to SDK store weirdness)
// Try to avoid usage of iterators in f.
func ApplyFuncIfNoError(ctx sdk.Context, f func(ctx sdk.Context) error) (err error) {
	// makes a new cache context, which all state changes get wrapped inside of.
	cacheCtx, write := ctx.CacheContext()
	err = f(cacheCtx)
	if err != nil {
		ctx.Logger().Error(err.Error())
	} else {
		// no error, write the output of f
		write()
		// Commented out ctx.EventManager as it was outputting duplicate events.
		// ctx.EventManager().EmitEvents(cacheCtx.EventManager().Events())
	}
	return err
}

// PrintPanicRecoveryError error logs the recoveryError, along with the stacktrace, if it can be parsed.
// If not emits them to stdout.
func PrintPanicRecoveryError(ctx sdk.Context, recoveryError interface{}) {
	errStackTrace := string(debug.Stack())
	switch e := recoveryError.(type) {
	case types.ErrorOutOfGas:
		ctx.Logger().Debug("out of gas error inside panic recovery block: " + e.Descriptor)
		return
	case string:
		ctx.Logger().Error("Recovering from (string) panic: " + e)
	case runtime.Error:
		ctx.Logger().Error("recovered (runtime.Error) panic: " + e.Error())
	case error:
		ctx.Logger().Error("recovered (error) panic: " + e.Error())
	default:
		ctx.Logger().Error("recovered (default) panic. Could not capture logs in ctx, see stdout")
		debug.PrintStack()
		return
	}
	ctx.Logger().Error("stack trace: " + errStackTrace)
}
