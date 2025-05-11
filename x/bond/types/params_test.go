package types_test

import (
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	// Create a test address
	pk := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pk.Address())

	testCases := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "valid address",
			input:    addr.String(),
			expected: nil,
		},
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "invalid address",
			input:    "invalid",
			expected: fmt.Errorf("invalid yield recipient address: decoding bech32 failed: invalid bech32 string length 7"),
		},
		{
			name:     "invalid type",
			input:    123,
			expected: fmt.Errorf("invalid parameter type: int"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldRecipient(tc.input)
			if tc.expected == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expected.Error(), err.Error())
			}
		})
	}
}

func TestValidateYieldTime(t *testing.T) {
	now := time.Now()
	pastTime := now.Add(-time.Hour)
	futureTime := now.Add(time.Hour)

	testCases := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "nil time pointer",
			input:    (*time.Time)(nil),
			expected: nil,
		},
		{
			name:     "future time",
			input:    &futureTime,
			expected: nil,
		},
		{
			name:     "past time",
			input:    &pastTime,
			expected: fmt.Errorf("yield time cannot be in the past"),
		},
		{
			name:     "invalid type",
			input:    "invalid",
			expected: fmt.Errorf("invalid parameter type: string"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldTime(tc.input)
			if tc.expected == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expected.Error(), err.Error())
			}
		})
	}
}

func TestValidateYieldAmount(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected error
	}{
		{
			name:     "nil input",
			input:    nil,
			expected: nil,
		},
		{
			name:     "zero amount",
			input:    sdkmath.ZeroInt(),
			expected: nil,
		},
		{
			name:     "positive amount",
			input:    sdkmath.NewInt(1000),
			expected: nil,
		},
		{
			name:     "negative amount",
			input:    sdkmath.NewInt(-1000),
			expected: fmt.Errorf("yield amount cannot be negative: -1000"),
		},
		{
			name:     "invalid type",
			input:    "invalid",
			expected: fmt.Errorf("invalid parameter type: string"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateYieldAmount(tc.input)
			if tc.expected == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expected.Error(), err.Error())
			}
		})
	}
}

func TestParams_Validate(t *testing.T) {
	// Create a test address
	pk := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pk.Address())
	futureTime := time.Now().Add(time.Hour)

	testCases := []struct {
		name     string
		params   types.Params
		expected error
	}{
		{
			name: "valid params",
			params: types.Params{
				YieldRecipient: addr.String(),
				YieldTime:      &futureTime,
				YieldAmount:    sdkmath.NewInt(1000),
			},
			expected: nil,
		},
		{
			name: "default params",
			params: types.Params{
				YieldRecipient: "",
				YieldTime:      nil,
				YieldAmount:    sdkmath.ZeroInt(),
			},
			expected: nil,
		},
		{
			name: "invalid recipient",
			params: types.Params{
				YieldRecipient: "invalid",
				YieldTime:      &futureTime,
				YieldAmount:    sdkmath.NewInt(1000),
			},
			expected: fmt.Errorf("invalid yield recipient address: decoding bech32 failed: invalid bech32 string length 7"),
		},
		{
			name: "invalid time",
			params: types.Params{
				YieldRecipient: addr.String(),
				YieldTime:      &time.Time{},
				YieldAmount:    sdkmath.NewInt(1000),
			},
			expected: fmt.Errorf("yield time cannot be in the past"),
		},
		{
			name: "invalid amount",
			params: types.Params{
				YieldRecipient: addr.String(),
				YieldTime:      &futureTime,
				YieldAmount:    sdkmath.NewInt(-1000),
			},
			expected: fmt.Errorf("yield amount cannot be negative: -1000"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expected == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Equal(t, tc.expected.Error(), err.Error())
			}
		})
	}
}
