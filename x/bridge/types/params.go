package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// DefaultAllowAllAuthorizeMessages is the default messages we allow.
	DefaultAllowAllAuthorizeMessages = []string{AllowAllAuthorizeMessages}
)

const (
	// DefaultBridgeDenom is the default token that will be bridged from Ethereum to the sequencer.
	DefaultBridgeDenom = "ufuel"

	// DefaultEthereumProxyContractAddress is the default contract address we expect to
	// receive deposit and authorize messages from.
	DefaultEthereumProxyContractAddress = "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"

	// AllowAllAuthorizeMessages can be used if we want to allow
	// all messages instead of specifying all of them one-by-one.
	AllowAllAuthorizeMessages = "*"

	// DefaultSupplyDeltaPeriod is the default frequency in block at which we report supply
	// delta info to Ethereum.
	DefaultSupplyDeltaPeriod = uint64(10)

	// vestingStartTimeDelay is a constant period of time during which tokens are completely locked.
	vestingStartTimeDelay = time.Hour * 24 * 365
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance.
func NewParams(
	bridgeDenom string,
	ethereumProxyContractAddress string,
	authorizeMessagesAllowed []string,
	supplyDeltaPeriod uint64,
	additionalBlockedAddresses []string,
) Params {
	// Setting a default start time.
	t0, err := time.Parse(time.DateOnly, "2024-01-01")
	if err != nil {

		// Panic if we error here, because shouldn't.
		panic(err)
	}

	return Params{
		BridgeDenom:                  bridgeDenom,
		EthereumProxyContractAddress: ethereumProxyContractAddress,
		AuthorizeMessagesAllowed:     authorizeMessagesAllowed,
		SupplyDeltaPeriod:            supplyDeltaPeriod,
		VestingStartTime:             t0,
		AdditionalBlockedAddresses:   additionalBlockedAddresses,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultBridgeDenom,
		DefaultEthereumProxyContractAddress,
		DefaultAllowAllAuthorizeMessages,
		DefaultSupplyDeltaPeriod,
		nil,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params.
func (p Params) Validate() error {

	// Validate bridge_denom.
	if err := ValidateBridgeDenom(p.BridgeDenom); err != nil {
		return err
	}

	// Validate the ethereum proxy contract address.
	if err := ValidateEthereumProxyContractAddress(p.EthereumProxyContractAddress); err != nil {
		return err
	}

	// Validate the authorize messages allowed.
	if err := ValidateAuthorizeMessagesAllowed(p.AuthorizeMessagesAllowed); err != nil {
		return err
	}

	// Validate supply delta period.
	if err := ValidateSupplyDeltaPeriod(p.SupplyDeltaPeriod); err != nil {
		return err
	}

	// Validate the vesting start time.
	if err := ValidateVestingStartTime(p.VestingStartTime); err != nil {
		return err
	}

	// AdditionalBlockedAddresses blocked addresses
	if err := ValidateBlockedAddresses(p.AdditionalBlockedAddresses); err != nil {
		return err
	}

	return nil
}

func ValidateBridgeDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type: %T", i)
	}
	if v == "" {
		return ErrParamsInvalid.Wrapf("bridge denom cannot be empty")
	}
	return nil
}

func ValidateEthereumProxyContractAddress(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type: %T", i)
	}
	if !common.IsHexAddress(v) {
		return ErrParamsInvalid.Wrapf("ethereum proxy contract address is invalid: %s", v)
	}
	return nil
}

func ValidateAuthorizeMessagesAllowed(i interface{}) error {
	messages, ok := i.([]string)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for authorizeMessagesAllowed: %T", i)
	}
	if len(messages) == 0 {
		return ErrParamsInvalid.Wrapf("authorize messages cannot be empty")
	}
	for _, msg := range messages {
		if msg == "" {
			return ErrParamsInvalid.Wrapf("authorize message cannot be empty")
		}
	}
	return nil
}

func ValidateSupplyDeltaPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for supplyDeltaPeriod: %T", i)
	}
	if v == 0 {
		return ErrParamsInvalid.Wrapf("supply delta period cannot be 0")
	}
	return nil
}

func ValidateVestingStartTime(i interface{}) error {
	v, ok := i.(time.Time)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for vestingStartTime: %T", i)
	}

	// Ensure the time is not zero, which is the zero value for time.Time and represents an unset value.
	if v.IsZero() {
		return ErrParamsInvalid.Wrapf("vesting start time must be set and cannot be the zero value")
	}

	return nil
}

func ValidateBlockedAddresses(i interface{}) error {
	additionalBlockedAddresses, ok := i.([]string)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for authorizeMessagesAllowed: %T", i)
	}

	for _, addr := range additionalBlockedAddresses {
		if addr == "" {
			return ErrParamsInvalid.Wrapf("blocked address cannot be empty")
		}

		// Attempt to decode the Bech32 address into an normal address
		_, errAcc := sdk.AccAddressFromBech32(addr)

		// If both decodings fail, return an error for this address
		if errAcc != nil {
			return ErrParamsInvalid.Wrapf("address %s is not a valid Bech32 encoded address", addr)
		}
	}

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
			"duration must be greater than vesting start time delay, got %s <= %s",
			duration, vestingStartTimeDelay,
		)
	}

	vestingStartTime := p.VestingStartTime.Add(vestingStartTimeDelay)
	vestingEndTime := p.VestingStartTime.Add(duration)
	return vestingStartTime, vestingEndTime, nil
}

// IsAuthorizedMessage returns true if the sdk.Msg TypeURL is present in Params.AuthorizeMessagesAllowed, otherwise,
// returns false
func (p *Params) IsAuthorizedMessage(msg sdk.Msg) bool {
	// Check that wildcard * option for allowing all message types is the only string in the array, if so, return true
	if len(p.AuthorizeMessagesAllowed) == 1 && p.AuthorizeMessagesAllowed[0] == AllowAllAuthorizeMessages {
		return true
	}

	for _, messageAllowed := range p.AuthorizeMessagesAllowed {
		if messageAllowed == sdk.MsgTypeURL(msg) {
			return true
		}
	}

	return false
}
