package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"
)

func TestParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  Params
		wantErr bool
	}{
		{
			name: "valid params",
			params: Params{
				Inflation: sdkmath.LegacyMustNewDecFromStr("0.5"),
				Authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			},
			wantErr: false,
		},
		{
			name: "valid params with empty authority",
			params: Params{
				Inflation: sdkmath.LegacyMustNewDecFromStr("0.5"),
				Authority: "",
			},
			wantErr: false,
		},
		{
			name: "invalid inflation - negative",
			params: Params{
				Inflation: sdkmath.LegacyMustNewDecFromStr("-0.1"),
				Authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			},
			wantErr: true,
		},
		{
			name: "invalid inflation - greater than 1",
			params: Params{
				Inflation: sdkmath.LegacyMustNewDecFromStr("1.1"),
				Authority: "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu",
			},
			wantErr: true,
		},
		{
			name: "invalid authority - invalid bech32",
			params: Params{
				Inflation: sdkmath.LegacyMustNewDecFromStr("0.5"),
				Authority: "invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()
	require.Equal(t, sdkmath.LegacyZeroDec(), params.Inflation)
	require.Equal(t, "", params.Authority)
}

func TestNewParams(t *testing.T) {
	inflation := sdkmath.LegacyMustNewDecFromStr("0.5")
	authority := "cosmos1qypqxpq9qcrsszg2pvxq6rs0zqg3yyc5lzv7xu"

	params := NewParams(inflation, authority)
	require.Equal(t, inflation, params.Inflation)
	require.Equal(t, authority, params.Authority)
}
