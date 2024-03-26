package types

import (
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

const (
	// AllowAllAuthorizeMessages can be used if we want to allow
	// all messages instead of specifying all of them one-by-one.
	AllowAllAuthorizeMessages = "*"

	// vestingStartTimeDelay is a constant period of time during which tokens are completely locked.
	vestingStartTimeDelay = time.Hour * 24 * 365
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	ethereumProxyContractAddress string,
	authorizeMessagesAllowed []string,
	supplyDeltaPeriod uint64,
) Params {
	return Params{
		EthereumProxyContractAddress: ethereumProxyContractAddress,
		AuthorizeMessagesAllowed:     authorizeMessagesAllowed,
		SupplyDeltaPeriod:            supplyDeltaPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	// TODO: consider setting more meaningful default params
	return NewParams("", nil, 0)
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

// VestingTimesFromVestingDuration returns the vesting start and end time based on the VestingStartTime parameter, the
// vestingStartTimeDelay, and a specified vesting duration, which must be greater than vestingStartTimeDelay since the
// vesting duration is included in the vestingStartTimeDelay.
//
// Example: for a VestingStartTime set to 2024 and a vesting duration of 2 years:
// - Actual vesting start time: 2024 + vestingStartTimeDelay = 2025
// - Actual vesting end time: 2024 + vesting duration = 2026
func (p Params) VestingTimesFromVestingDuration(duration time.Duration) (time.Time, time.Time, error) {

	if duration <= vestingStartTimeDelay {
		return time.Time{}, time.Time{}, ErrInvalidVestingDuration.Wrapf(
			"must be greater than vesting start time delay, got %s <= %s",
			duration, vestingStartTimeDelay,
		)
	}

	vestingStartTime := p.VestingStartTime.Add(vestingStartTimeDelay)
	vestingEndTime := p.VestingStartTime.Add(duration)
	return vestingStartTime, vestingEndTime, nil
}
