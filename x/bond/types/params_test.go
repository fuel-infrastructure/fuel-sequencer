package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestParamKeyTable(t *testing.T) {
	table := types.ParamKeyTable()
	require.NotNil(t, table)
}

func TestNewParams(t *testing.T) {
	yieldRecipient := "cosmos124maqmcqv8tquy764ktz7cu0gxnzfw54k9cmz5"
	future := time.Now().Add(time.Hour)
	testCases := []struct {
		name           string
		yieldRecipient string
		yieldTime      *time.Time
		yieldAmount    sdkmath.Int
		expErr         bool
	}{
		{
			name:           "valid params",
			yieldRecipient: yieldRecipient,
			yieldTime:      &future,
			yieldAmount:    sdkmath.NewInt(1000000),
			expErr:         false,
		},
		{
			name:           "empty recipient",
			yieldRecipient: "",
			yieldTime:      &future,
			yieldAmount:    sdkmath.NewInt(1000000),
			expErr:         false,
		},
		{
			name:           "invalid recipient",
			yieldRecipient: "invalid",
			yieldTime:      &future,
			yieldAmount:    sdkmath.NewInt(1000000),
			expErr:         true,
		},
		{
			name:           "nil time",
			yieldRecipient: yieldRecipient,
			yieldTime:      nil,
			yieldAmount:    sdkmath.NewInt(1000000),
			expErr:         false,
		},
		{
			name:           "negative amount",
			yieldRecipient: yieldRecipient,
			yieldTime:      &future,
			yieldAmount:    sdkmath.NewInt(-1),
			expErr:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(tc.yieldRecipient, tc.yieldTime, tc.yieldAmount)
			err := params.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.yieldRecipient, params.YieldRecipient)
				require.Equal(t, tc.yieldTime, params.YieldTime)
				require.Equal(t, tc.yieldAmount, params.YieldAmount)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.Equal(t, "", params.YieldRecipient)
	require.Equal(t, (*time.Time)(nil), params.YieldTime)
	require.Equal(t, sdkmath.ZeroInt(), params.YieldAmount)
}

func TestParamSetPairs(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()
	require.Empty(t, pairs)
}

func TestValidateYieldRecipient(t *testing.T) {
	yieldRecipient := "cosmos124maqmcqv8tquy764ktz7cu0gxnzfw54k9cmz5"
	testCases := []struct {
		name      string
		recipient string
		expErr    bool
	}{
		{
			name:      "valid recipient",
			recipient: yieldRecipient,
			expErr:    false,
		},
		{
			name:      "empty recipient",
			recipient: "",
			expErr:    false,
		},
		{
			name:      "invalid recipient",
			recipient: "invalid",
			expErr:    true,
		},
		{
			name:      "whitespace recipient",
			recipient: "  ",
			expErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldRecipient(tc.recipient)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateYieldTime(t *testing.T) {
	baseTime := time.Now()
	past := baseTime.Add(-time.Hour)
	future := baseTime.Add(time.Hour)

	testCases := []struct {
		name   string
		time   *time.Time
		expErr bool
	}{
		{
			name:   "valid future time",
			time:   &future,
			expErr: false,
		},
		{
			name:   "nil time",
			time:   nil,
			expErr: false,
		},
		{
			name:   "past time",
			time:   &past,
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldTime(tc.time)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateYieldAmount(t *testing.T) {
	testCases := []struct {
		name   string
		amount sdkmath.Int
		expErr bool
	}{
		{
			name:   "valid amount",
			amount: sdkmath.NewInt(1000000),
			expErr: false,
		},
		{
			name:   "zero amount",
			amount: sdkmath.ZeroInt(),
			expErr: false,
		},
		{
			name:   "negative amount",
			amount: sdkmath.NewInt(-1),
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldAmount(tc.amount)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
