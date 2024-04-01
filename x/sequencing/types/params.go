package types

import (
	errorsmod "cosmossdk.io/errors"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	DefaultMaxBlobSize = 10 * 1024 * 1024 // 10 MB
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maxBlobSizeBytes, gasPerBlobByte uint64) Params {
	return Params{
		MaxBlobSizeBytes: maxBlobSizeBytes,
		GasPerBlobByte:   gasPerBlobByte,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxBlobSize, 0)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {

	// Validate max blob size bytes is the maximum size of a blob the sequencer can accept.
	if err := ValidateMaxBlobSizeBytes(p.MaxBlobSizeBytes); err != nil {
		return err
	}

	return nil
}

// ValidateMaxBlobSizeBytes verifies that the max blob size bytes is of the correct type and is greater than 0.
func ValidateMaxBlobSizeBytes(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return errorsmod.Wrapf(
			ErrParamsInvalid,
			"invalid parameter type for maxBlobSizeBytes: %T",
			i,
		)
	}

	// Check that MaxBlobSizeBytes is greater than 0.
	if v == 0 {
		return errorsmod.Wrapf(
			ErrParamsInvalid,
			"maxBlobSizeBytes must be greater than 0",
		)
	}

	return nil
}
