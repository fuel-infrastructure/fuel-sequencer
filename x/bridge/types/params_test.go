package types_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestVestingTimesFromVestingDuration(t *testing.T) {

	// Helper durations.
	years1 := time.Hour * 24 * 365
	years2 := years1 * 2
	years4 := years1 * 4

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t0Plus1Year := t0.Add(years1)  // accounts for vesting start time delay
	t0Plus2Years := t0.Add(years2) // used for 2-year vesting duration
	t0Plus4Years := t0.Add(years4) // used for 4-year vesting duration

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
			name:            "Zero vesting duration is less than vestingStartTimeDelay => err",
			vestingDuration: 0,
			expErrMsg:       "must be greater than vesting start time delay, got 0s <= 8760h0m0s",
		},
		{
			name:            "Almost 1 year vesting duration is less than vestingStartTimeDelay => err",
			vestingDuration: years1 - 1,
			expErrMsg:       "must be greater than vesting start time delay, got 8759h59m59.999999999s <= 8760h0m0s",
		},
		{
			name:            "1 year vesting duration is equal to vestingStartTimeDelay => err",
			vestingDuration: years1,
			expErrMsg:       "must be greater than vesting start time delay, got 8760h0m0s <= 8760h0m0s",
		},
		{
			name:            "Just over 1 year vesting duration is valid",
			vestingDuration: years1 + 1,
			expStartTime:    t0Plus1Year,        // t0 + 1 year
			expEndTime:      t0Plus1Year.Add(1), // t0 + vestingDuration
		},
		{
			name:            "2 years => 1 year lock followed by 1 year vesting",
			vestingDuration: years2,
			expStartTime:    t0Plus1Year,  // t0 + 1 year
			expEndTime:      t0Plus2Years, // t0 + vestingDuration
		},
		{
			name:            "4 years => 1 year lock followed by 3 year vesting",
			vestingDuration: years4,
			expStartTime:    t0Plus1Year,  // t0 + 1 year
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

func TestValidateEthereumProxyContractAddress(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid address", "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853", false},
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

func TestValidateAuthorizeMessagesAllowed(t *testing.T) {
	testCases := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{"Valid messages", []string{"message1", "message2"}, false},
		{"Empty slice", []string{}, true},
		{"Slice with empty message", []string{"message1", ""}, true},
		{"Non-slice type", "not a slice", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateAuthorizeMessagesAllowed(tc.input)
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

func TestParams_Validate(t *testing.T) {
	validBridgeDenom := "ufuel"
	validEthereumProxyContractAddress := "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
	validAuthorizeMessagesAllowed := []string{"authorizeMessage1", "authorizeMessage2"}
	validSupplyDeltaPeriod := uint64(10)
	validVestingStartTime := time.Now()
	validMaxEthBlockUpdateDelay := time.Hour

	// Creating an invalid ethereum proxy contract address for testing
	invalidEthereumProxyContractAddress := "0xInvalidAddress"

	testCases := []struct {
		name      string
		params    types.Params
		expectErr bool
	}{
		{
			name:      "Valid parameters - default params",
			params:    types.DefaultParams(),
			expectErr: false,
		},
		{
			name: "Valid parameters - non-default params",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
				MaxEthBlockUpdateDelay:       validMaxEthBlockUpdateDelay,
			},
			expectErr: false,
		},
		{
			name: "Invalid bridge denom (empty)",
			params: types.Params{
				BridgeDenom:                  "",
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
			},
			expectErr: true,
		},
		{
			name: "Invalid ethereum proxy contract address",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: invalidEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
			},
			expectErr: true,
		},
		{
			name: "Empty authorize messages allowed",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     []string{},
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
			},
			expectErr: true,
		},
		{
			name: "Zero supply delta period",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            0,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
			},
			expectErr: true,
		},
		{
			name: "Zero vesting start time",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             time.Time{},
				AdditionalBlockedAddresses:   []string{},
			},
			expectErr: true,
		},
		{
			name: "bad blocked address",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             time.Time{},
				AdditionalBlockedAddresses:   []string{"invalidBech32Address"},
			},
			expectErr: true,
		},
		{
			name: "negative tolerance for no Ethereum block syncing",
			params: types.Params{
				BridgeDenom:                  validBridgeDenom,
				EthereumProxyContractAddress: validEthereumProxyContractAddress,
				AuthorizeMessagesAllowed:     validAuthorizeMessagesAllowed,
				SupplyDeltaPeriod:            validSupplyDeltaPeriod,
				VestingStartTime:             validVestingStartTime,
				AdditionalBlockedAddresses:   []string{},
				MaxEthBlockUpdateDelay:       time.Duration(-1),
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

func TestIsAuthorizedMessage(t *testing.T) {
	testCases := []struct {
		name      string
		params    *types.Params
		msg       sdk.Msg
		expResult bool
	}{
		{
			name: "returns true if message is authorized (messages allowed is not *)",
			params: &types.Params{
				AuthorizeMessagesAllowed: []string{"msg1", "msg2", sdk.MsgTypeURL(&banktypes.MsgSend{})},
			},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: true,
		},
		{
			name: "returns true if message is authorized (messages allowed is *)",
			params: &types.Params{
				AuthorizeMessagesAllowed: []string{"*"},
			},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: true,
		},
		{
			name: "returns false if message is not authorized",
			params: &types.Params{
				AuthorizeMessagesAllowed: []string{"msg1", "msg2", "msg3"},
			},
			msg: &banktypes.MsgSend{
				FromAddress: "addr1",
				ToAddress:   "addr2",
				Amount:      nil,
			},
			expResult: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualResult := tc.params.IsAuthorizedMessage(tc.msg)
			require.Equal(t, tc.expResult, actualResult)
		})
	}
}
