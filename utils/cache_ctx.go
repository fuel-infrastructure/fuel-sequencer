package utils

// This file was imported from https://github.com/osmosis-labs/osmosis/blob/v14.0.0/osmoutils/cache_ctx.go

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
