package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func TestParamKeyTable(t *testing.T) {
	table := types.ParamKeyTable()
	require.NotNil(t, table)
}

func TestNewParams(t *testing.T) {
	testCases := []struct {
		name      string
		inflation sdkmath.LegacyDec
		authority string
		expErr    bool
	}{
		{
			name:      "valid params",
			inflation: sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
			authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expErr:    false,
		},
		{
			name:      "empty authority",
			inflation: sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
			authority: "",
			expErr:    false,
		},
		{
			name:      "invalid authority",
			inflation: sdkmath.LegacyNewDec(5),
			authority: "invalid",
			expErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := types.NewParams(tc.inflation, tc.authority)
			err := params.Validate()
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.inflation, params.Inflation)
				require.Equal(t, tc.authority, params.Authority)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := types.DefaultParams()
	require.Equal(t, sdkmath.LegacyZeroDec(), params.Inflation)
	require.Equal(t, "", params.Authority)
}

func TestParamSetPairs(t *testing.T) {
	params := types.DefaultParams()
	pairs := params.ParamSetPairs()
	require.Len(t, pairs, 2)
	require.Equal(t, types.KeyInflation, pairs[0].Key)
	require.Equal(t, types.KeyAuthority, pairs[1].Key)
}

func TestValidateInflation(t *testing.T) {
	testCases := []struct {
		name      string
		inflation sdkmath.LegacyDec
		expErr    bool
	}{
		{
			name:      "zero inflation",
			inflation: sdkmath.LegacyZeroDec(),
			expErr:    false,
		},
		{
			name:      "valid inflation",
			inflation: sdkmath.LegacyNewDecWithPrec(5, 1), // 0.5
			expErr:    false,
		},
		{
			name:      "negative inflation",
			inflation: sdkmath.LegacyNewDec(-1),
			expErr:    true,
		},
		{
			name:      "inflation greater than 1",
			inflation: sdkmath.LegacyNewDec(2),
			expErr:    true,
		},
		{
			name:      "inflation equal to 1",
			inflation: sdkmath.LegacyOneDec(),
			expErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateInflation(tc.inflation)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateAuthority(t *testing.T) {
	testCases := []struct {
		name      string
		authority string
		expErr    bool
	}{
		{
			name:      "valid authority",
			authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			expErr:    false,
		},
		{
			name:      "empty authority",
			authority: "",
			expErr:    false,
		},
		{
			name:      "invalid authority",
			authority: "invalid",
			expErr:    true,
		},
		{
			name:      "whitespace authority",
			authority: "  ",
			expErr:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := types.ValidateAuthority(tc.authority)
			if tc.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
