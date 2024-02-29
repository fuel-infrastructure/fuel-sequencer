package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	// AllowAllAuthorizeMessages can be used if we want to allow
	// all messages instead of specifying all of them one-by-one.
	AllowAllAuthorizeMessages = "*"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	depositContractAddress string,
	authorizeContractAddress string,
	authorizeMessagesAllowed []string,
	supplyDeltaPeriod uint64,
) Params {
	return Params{
		DepositContractAddress:   depositContractAddress,
		AuthorizeContractAddress: authorizeContractAddress,
		AuthorizeMessagesAllowed: authorizeMessagesAllowed,
		SupplyDeltaPeriod:        supplyDeltaPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams("", "", nil, 0)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params
func (p Params) Validate() error {

	// TODO: validate DepositContractAddress

	// TODO: validate AuthorizeContractAddress

	// TODO: validate AuthorizeMessagesAllowed

	// TODO: validate SupplyDeltaPeriod

	return nil
}
