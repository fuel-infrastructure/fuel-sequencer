package types_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestVestingTimesFromVestingDuration(t *testing.T) {

	// Helper durations.
	years1 := time.Hour * 24 * 365
	years2 := years1 * 2
	years4 := years1 * 4
	months6 := years1 / 2

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t0Plus1Year := t0.Add(years1)    // accounts for vesting start time delay
	t0Plus2Years := t0.Add(years2)   // used for 2-year vesting duration
	t0Plus4Years := t0.Add(years4)   // used for 4-year vesting duration
	t0Plus6Months := t0.Add(months6) // used for 6-month vesting duration

	// Fixed params.
	params := types.Params{
		VestingStartTime: t0,
	}

	testCases := []struct {
		name            string
		vestingDuration time.Duration
		expStartTime    time.Time
		expEndTime      time.Time
		expErrMsg       string
	}{
		{
			name:            "Zero vesting duration => err",
			vestingDuration: 0,
			expErrMsg:       "expected duration to be greater than 0",
		},
		{
			name:            "negative vesting duration => err",
			vestingDuration: -1,
			expErrMsg:       "expected duration to be greater than 0",
		},
		{
			name:            "6 months vesting duration => no cliff + 6 months vesting duration",
			vestingDuration: months6,
			expStartTime:    t0,
			expEndTime:      t0Plus6Months,
		},
		{
			name:            "Almost 1 year vesting duration => no cliff + almost 1 year vesting duration",
			vestingDuration: years1 - 1,
			expStartTime:    t0,
			expEndTime:      t0Plus1Year.Add(-1),
		},
		{
			name:            "1 year vesting duration => no cliff + 1 year vesting duration",
			vestingDuration: years1,
			expStartTime:    t0,
			expEndTime:      t0Plus1Year,
		},
		{
			name:            "Just over 1 year vesting duration => no cliff + just over 1 year of vesting duration",
			vestingDuration: years1 + 1,
			expStartTime:    t0,
			expEndTime:      t0Plus1Year.Add(1), // t0 + vestingDuration
		},
		{
			name:            "2 years vesting duration => no cliff + 2 years vesting duration",
			vestingDuration: years2,
			expStartTime:    t0,
			expEndTime:      t0Plus2Years, // t0 + vestingDuration
		},
		{
			name:            "4 years vesting duration => no cliff + 4 years vesting duration",
			vestingDuration: years4,
			expStartTime:    t0,
			expEndTime:      t0Plus4Years, // t0 + vestingDuration
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			startTime, endTime, err := params.VestingTimesFromVestingDuration(tc.vestingDuration)
			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expStartTime, startTime)
			require.Equal(t, tc.expEndTime, endTime)
		})
	}
}

func TestValidateBridgeDenom(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid denom", "ufuel", false},
		{"Empty denom", "", true},
		{"Non-string denom", 123, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateBridgeDenom(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateBridgeDenomTotalSupply(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid supply", sdkmath.NewInt(10), false},
		{"Zero supply", sdkmath.NewInt(0), true},
		{"Non-Int type", "not an Int", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateBridgeDenomTotalSupply(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateEthereumProxyContractAddress(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid address", "0x0165878A594ca255338adfa4d48449f69242Eb8F", false},
		{"Invalid address", "0x123", true},
		{"Non-string address", 12345, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateEthereumProxyContractAddress(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSupplyDeltaPeriod(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid period", uint64(10), false},
		{"Zero period", uint64(0), true},
		{"Non-uint64 type", "not a uint64", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateSupplyDeltaPeriod(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateVestingStartTime(t *testing.T) {
	validTime := time.Now()
	zeroTime := time.Time{}

	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid time", validTime, false},
		{"Zero time", zeroTime, true},
		{"Non-time type", "not a time", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateVestingStartTime(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateMaxEthBlockUpdateDelay(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid delay - non-zero value", time.Hour, false},
		{"Valid delay - Zero value", time.Duration(0), false},
		{"Invalid delay - value is negative", time.Duration(-1), true},
		{"Non-time.Duration type", "not a time.Duration", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateMaxEthBlockUpdateDelay(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateSequencerTxsAllocation(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{
			"Valid SequencerTxsAllocation - MinimumSequencerTxsAllocation < value < MaximumSequencerTxsAllocation",
			types.MinimumSequencerTxsAllocation.Add(sdkmath.LegacyMustNewDecFromStr("0.1")),
			false,
		},
		{
			"Valid SequencerTxsAllocation - value == MinimumSequencerTxsAllocation",
			types.MinimumSequencerTxsAllocation,
			false,
		},
		{
			"Valid SequencerTxsAllocation - value == MaximumSequencerTxsAllocation",
			types.MaximumSequencerTxsAllocation,
			false,
		},
		{
			"Invalid SequencerTxsAllocation - value < MinimumSequencerTxsAllocation",
			types.MinimumSequencerTxsAllocation.Sub(sdkmath.LegacyMustNewDecFromStr("0.01")),
			true,
		},
		{
			"Invalid SequencerTxsAllocation - value > MaximumSequencerTxsAllocation",
			types.MaximumSequencerTxsAllocation.Add(sdkmath.LegacyMustNewDecFromStr("0.01")),
			true,
		},
		{"Non-LegacyDec type", "not a LegacyDec", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateSequencerTxsAllocation(tc.input)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParams_Validate(t *testing.T) {
	validBridgeDenom := "ufuel"
	validBridgeDenomTotalSupply := sdkmath.NewInt(10_000_000_000)
	validEthereumProxyContractAddress := "0x0165878A594ca255338adfa4d48449f69242Eb8F"
	validSupplyDeltaPeriod := uint64(10)
	validVestingStartTime := time.Now()
	validMaxEthBlockUpdateDelay := time.Hour
	validSequencerTxsAllocation := sdkmath.LegacyMustNewDecFromStr("0.3")
	validAdditionalBlockedAddresses := []string(nil)

	// Creating an invalid ethereum proxy contract address for testing
	invalidEthereumProxyContractAddress := "0xInvalidAddress"

	testCases := []struct {
		name      string
		params    types.Params
		expectErr bool
	}{
		{
			name:      "Invalid parameters - default params",
			params:    types.DefaultParams(),
			expectErr: true,
		},
		{
			name: "Valid parameters - non-default params",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: false,
		},
		{
			name: "Invalid bridge denom (empty)",
			params: types.Params{
				BridgeDenom:                  "",
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "Invalid ethereum proxy contract address",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: invalidEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "Zero supply delta period",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            0,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "Zero vesting start time",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             time.Time{},
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "bad blocked address",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{"invalidBech32Address"},
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "negative tolerance for no Ethereum block syncing",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       time.Duration(-1),
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
		{
			name: "SequencerTxsAllocation too small",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       validBridgeDenomTotalSupply,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation: types.MinimumSequencerTxsAllocation.Sub(
					sdkmath.LegacyMustNewDecFromStr("0.01"),
				),
			},
			expectErr: true,
		},
		{
			name: "BridgeDenomTotalSupply too small",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				BridgeDenomTotalSupply:       sdkmath.ZeroInt(), // 0 is too small
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   validAdditionalBlockedAddresses,
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
				SequencerTxsAllocation:       validSequencerTxsAllocation,
			},
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.params.Validate()
			if tc.expectErr {
				require.Error(t, err, "Expected an error for test case: %s", tc.name)
			} else {
				require.NoError(t, err, "Expected no error for test case: %s", tc.name)
			}
		})
	}
}

func TestValidateBlockedAddresses(t *testing.T) {
	testCases := []struct {
		name        string
		input       interface{}
		expectError bool
	}{
		{
			name:        "Valid account address",
			input:       []string{"fuelsequencer1zkaa9906nckwl4m0ysunscuv5edma04n3u3u8r"},
			expectError: false,
		},
		{
			name:        "Empty list is valid",
			input:       []string{},
			expectError: false,
		},
		{
			name:        "Invalid Bech32 address",
			input:       []string{"invalidBech32Address"},
			expectError: true,
		},
		{
			name:        "Empty address",
			input:       []string{""},
			expectError: true,
		},
		{
			name:        "Invalid parameter type",
			input:       "NotASliceOfString",
			expectError: true,
		},
		{
			name:        "Mixed valid and invalid addresses",
			input:       []string{"fuelsequencer1zkaa9906nckwl4m0ysunscuv5edma04n3u3u8r", "invalidBech32Address"},
			expectError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateBlockedAddresses(tc.input)
			if tc.expectError {
				require.Error(t, err, "Expected an error for test case: %s", tc.name)
			} else {
				require.NoError(t, err, "Expected no error for test case: %s", tc.name)
			}
		})
	}
}

func TestValidateBlockedAddresses_Nil(t *testing.T) {
	err := types.ValidateBlockedAddresses(nil)
	require.Error(t, err, "Expected an error")
}
