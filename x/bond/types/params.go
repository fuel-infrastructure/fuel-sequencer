package types

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	KeyInflation = []byte("Inflation")
	KeyAuthority = []byte("Authority")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	inflation sdkmath.LegacyDec,
	authority string,
) Params {
	return Params{
		Inflation: inflation,
		Authority: authority,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		sdkmath.LegacyZeroDec(),
		"", // Will be set to governance module account in InitGenesis
	)
}

// ParamSetPairs implements params.ParamSet
//
// Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := ValidateInflation(p.Inflation); err != nil {
		return err
	}
	if err := ValidateAuthority(p.Authority); err != nil {
		return err
	}
	return nil
}

// ValidateInflation validates the inflation parameter
func ValidateInflation(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation cannot be negative: %s", v)
	}
	if v.GT(sdkmath.LegacyOneDec()) {
		return fmt.Errorf("inflation cannot be greater than 1: %s", v)
	}

	return nil
}

// ValidateAuthority validates the authority parameter
func ValidateAuthority(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return nil // Empty string is allowed (will be set to governance module account)
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid authority address: %w", err)
	}

	return nil
}
