package metrics

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	"github.com/stretchr/testify/require"
)

func TestObserveYieldMinting(t *testing.T) {
	// Create a context with a mock SDK context
	ctx := context.Background()
	sdkCtx := sdk.Context{}
	ctx = sdkCtx.WithContext(ctx)

	// Test cases
	testCases := []struct {
		name     string
		amount   sdk.Coins
		height   int64
		expected int
	}{
		{
			name:     "single coin",
			amount:   sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100))),
			height:   100,
			expected: 1,
		},
		{
			name:     "multiple coins",
			amount:   sdk.NewCoins(sdk.NewCoin("ufuel", sdkmath.NewInt(100)), sdk.NewCoin("stfuel", sdkmath.NewInt(50))),
			height:   200,
			expected: 2,
		},
		{
			name:     "zero amount",
			amount:   sdk.NewCoins(),
			height:   300,
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			ObserveYieldMinting(ctx, tc.amount, tc.height)
			// Note: Since telemetry is a side effect, we can't directly test the counter value
			// We're mainly testing that the function doesn't panic
		})
	}
}

func TestSetParamsUpdate(t *testing.T) {
	// Create a context with a mock SDK context
	ctx := context.Background()
	sdkCtx := sdk.Context{}
	ctx = sdkCtx.WithContext(ctx)

	now := time.Now()
	zeroTime := time.Time{}

	// Test cases
	testCases := []struct {
		name     string
		params   types.Params
		expected int
	}{
		{
			name: "valid params",
			params: types.Params{
				YieldRecipient: "fuel1...",
				YieldTime:      &now,
				YieldAmount:    sdkmath.NewInt(1000),
			},
			expected: 1,
		},
		{
			name: "empty params",
			params: types.Params{
				YieldRecipient: "",
				YieldTime:      &zeroTime,
				YieldAmount:    sdkmath.ZeroInt(),
			},
			expected: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the function
			require.NotPanics(t, func() {
				SetParamsUpdate(ctx, tc.params)
			})
			// Note: Since telemetry is a side effect, we can't directly test the counter value
			// We're mainly testing that the function doesn't panic
		})
	}
}
