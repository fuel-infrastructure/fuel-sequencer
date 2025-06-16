package types

import (
	"fmt"
	"strings"
	time "time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	yieldRecipient string,
	yieldTime *time.Time,
	yieldAmount sdkmath.Int,
) Params {
	return Params{
		YieldRecipient: yieldRecipient,
		YieldTime:      yieldTime,
		YieldAmount:    yieldAmount,
	}
}

// DefaultParams returns default bond module parameters.
func DefaultParams() Params {
	return Params{
		YieldRecipient: "",
		YieldTime:      nil,
		YieldAmount:    sdkmath.ZeroInt(),
	}
}

// ParamSetPairs implements params.ParamSet
//
// Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := ValidateYieldRecipient(p.YieldRecipient); err != nil {
		return err
	}
	if err := ValidateYieldTime(p.YieldTime); err != nil {
		return err
	}
	if err := ValidateYieldAmount(p.YieldAmount); err != nil {
		return err
	}
	return nil
}

// ValidateYieldRecipient validates the yield recipient parameter
func ValidateYieldRecipient(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return nil // Empty string is allowed (to be updated later)
	}

	_, err := sdk.AccAddressFromBech32(v)
	if err != nil {
		return fmt.Errorf("invalid yield recipient address: %w", err)
	}

	return nil
}

// ValidateYieldTime validates the yield time parameter
func ValidateYieldTime(i interface{}) error {
	if i == nil {
		return nil // nil is allowed (to be updated later)
	}

	v, ok := i.(*time.Time)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == nil {
		return nil // nil pointer is allowed (to be updated later)
	}

	if v.Before(time.Now()) {
		return fmt.Errorf("yield time cannot be in the past")
	}

	return nil
}

// ValidateYieldAmount validates the yield amount parameter
func ValidateYieldAmount(i interface{}) error {
	if i == nil {
		return nil // nil is allowed (to be updated later)
	}

	v, ok := i.(sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNil() {
		return nil // nil is allowed (to be updated later)
	}

	if v.IsNegative() {
		return fmt.Errorf("yield amount cannot be negative: %s", v)
	}

	return nil
}
