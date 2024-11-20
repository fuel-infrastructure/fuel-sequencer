package types

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/ethereum/go-ethereum/common"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (

	// DefaultAuthorizeMessagesAllowed is the Sequencer whitelisted messages we support for AuthorizeTx execution. By
	// default, we will support staking operations, bank transfers, voting on proposals and Ethereum withdrawals.
	DefaultAuthorizeMessagesAllowed = []string{
		"/fuelsequencer.bridge.v1.MsgWithdrawToEthereum",
		"/cosmos.bank.v1beta1.MsgSend",
		"/cosmos.staking.v1beta1.MsgDelegate",
		"/cosmos.staking.v1beta1.MsgBeginRedelegate",
		"/cosmos.staking.v1beta1.MsgUndelegate",
		"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward",
		"/cosmos.distribution.v1beta1.MsgSetWithdrawAddress",
		"/cosmos.gov.v1.MsgVote",
	}

	// DefaultMaxEthBlockUpdateDelay is the default value for tolerating validators not reaching consensus to sync
	// up with Ethereum. This is set to 1 hour by default.
	DefaultMaxEthBlockUpdateDelay = time.Hour

	// DefaultSequencerTxsAllocation is the default value for the percentage that controls the maximum amount of block
	// space allocated to Sequencer-native transactions during heavy bridge usage.
	DefaultSequencerTxsAllocation = sdkmath.LegacyMustNewDecFromStr("0.3")

	// MinimumSequencerTxsAllocation is the minimum value that SequencerTxsAllocation can be set to. This minimum is set
	// as a measure against mempool saturation and transaction censorship when the bridge is under heavy usage.
	MinimumSequencerTxsAllocation = sdkmath.LegacyMustNewDecFromStr("0.1")

	// MaximumSequencerTxsAllocation is the maximum value that SequencerTxsAllocation can be set to. This maximum is set
	// as a measure against mistakes. This is important because if set to a very high value, the block production
	// algorithm may not be able to allocate block space to critical transactions.
	MaximumSequencerTxsAllocation = sdkmath.LegacyMustNewDecFromStr("0.5")

	// DefaultVestingStartTime is the default vesting start time, intentionally invalid to enforce explicit setting of
	// this value.
	DefaultVestingStartTime = time.Time{}

	// DefaultBridgeDenomTotalSupply is the default bridge denom total supply, intentionally invalid to enforce explicit
	// setting of this value.
	DefaultBridgeDenomTotalSupply = sdkmath.ZeroInt()
)

const (
	// DefaultBridgeDenom is the default token that will be bridged from Ethereum to the sequencer.
	DefaultBridgeDenom = "ufuel"

	// DefaultEthereumProxyContractAddress is the default contract address we expect to
	// receive deposit and authorize messages from.
	DefaultEthereumProxyContractAddress = "0x0165878A594ca255338adfa4d48449f69242Eb8F"

	// DefaultSupplyDeltaPeriod is the default frequency in block at which we report supply
	// delta info to Ethereum.
	DefaultSupplyDeltaPeriod = uint64(10)

	// vestingStartTimeDelay is a constant period of time during which tokens are completely locked.
	vestingStartTimeDelay = time.Hour * 24 * 365

	// DefaultInjectedEventTxMaxBytes is the default max size in bytes for an injected event tx in a block. This is set
	// to 20000000 assuming a max block size of 22020096 bytes, index tx size 103 bytes and supply delta tx size of 105.
	// The remaining bytes serve as a buffer. This default needs to be revised if any of the values above change.
	DefaultInjectedEventTxMaxBytes = 20000000

	// MinimumInjectedEventTxMaxBytes is the minimum value that InjectedEventTxMaxBytes can be set to. This is done to
	// prevent setting InjectedEventTxMaxBytes to a very small value, and thus avoiding situations where event txs are
	// never injected due to a strict InjectedEventTxMaxBytes.
	MinimumInjectedEventTxMaxBytes = 1024

	// DefaultMaxAuthorizeMessages is the maximum amount of Cosmos SDK messages that an Authorize transaction can have
	// by default
	DefaultMaxAuthorizeMessages = 10

	// MinimumMaxAuthorizeMessages is the minimum value that MaxAuthorizeMessages can be set to. This is set to one to
	// prevent mistakes that could cause all Authorize transactions to get skipped. Governance should set
	// AuthorizeMessagesAllowed to [] if the execution of Authorize txs is to be disabled.
	MinimumMaxAuthorizeMessages = 1
)

// NewParams creates a new Params instance.
func NewParams(
	bridgeDenom string,
	bridgeDenomTotalSupply sdkmath.Int,
	ethereumProxyContractAddress string,
	authorizeMessagesAllowed []string,
	supplyDeltaPeriod uint64,
	vestingStartTime time.Time,
	additionalBlockedAddresses []string,
	maxEthBlockUpdateDelay time.Duration,
	injectedEventTxMaxBytes uint64,
	sequencerTxsAllocation sdkmath.LegacyDec,
	maxAuthorizeMessages uint64,
) Params {
	return Params{
		BridgeDenom:                  bridgeDenom,
		BridgeDenomTotalSupply:       bridgeDenomTotalSupply,
		EthereumProxyContractAddress: ethereumProxyContractAddress,
		AuthorizeMessagesAllowed:     authorizeMessagesAllowed,
		SupplyDeltaPeriod:            supplyDeltaPeriod,
		VestingStartTime:             vestingStartTime,
		AdditionalBlockedAddresses:   additionalBlockedAddresses,
		MaxEthBlockUpdateDelay:       maxEthBlockUpdateDelay,
		InjectedEventTxMaxBytes:      injectedEventTxMaxBytes,
		SequencerTxsAllocation:       sequencerTxsAllocation,
		MaxAuthorizeMessages:         maxAuthorizeMessages,
	}
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultBridgeDenom,
		DefaultBridgeDenomTotalSupply,
		DefaultEthereumProxyContractAddress,
		DefaultAuthorizeMessagesAllowed,
		DefaultSupplyDeltaPeriod,
		DefaultVestingStartTime,
		nil,
		DefaultMaxEthBlockUpdateDelay,
		DefaultInjectedEventTxMaxBytes,
		DefaultSequencerTxsAllocation,
		DefaultMaxAuthorizeMessages,
	)
}

// ParamSetPairs implements params.ParamSet
//
// Deprecated.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{}
}

// Validate validates the set of params.
func (p Params) Validate() error {

	if err := ValidateBridgeDenom(p.BridgeDenom); err != nil {
		return err
	}

	if err := ValidateBridgeDenomTotalSupply(p.BridgeDenomTotalSupply); err != nil {
		return err
	}

	if err := ValidateEthereumProxyContractAddress(p.EthereumProxyContractAddress); err != nil {
		return err
	}

	if err := ValidateAuthorizeMessagesAllowed(p.AuthorizeMessagesAllowed); err != nil {
		return err
	}

	if err := ValidateSupplyDeltaPeriod(p.SupplyDeltaPeriod); err != nil {
		return err
	}

	if err := ValidateVestingStartTime(p.VestingStartTime); err != nil {
		return err
	}

	if err := ValidateBlockedAddresses(p.AdditionalBlockedAddresses); err != nil {
		return err
	}

	if err := ValidateMaxEthBlockUpdateDelay(p.MaxEthBlockUpdateDelay); err != nil {
		return err
	}

	if err := ValidateInjectedEventTxMaxBytes(p.InjectedEventTxMaxBytes); err != nil {
		return err
	}

	if err := ValidateSequencerTxsAllocation(p.SequencerTxsAllocation); err != nil {
		return err
	}

	if err := ValidateMaxAuthorizeMessages(p.MaxAuthorizeMessages); err != nil {
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

func ValidateBridgeDenomTotalSupply(i interface{}) error {
	v, ok := i.(sdkmath.Int)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type: %T", i)
	}
	if !v.IsPositive() {
		return ErrParamsInvalid.Wrapf("bridge denom total supply must be positive, got: %s", v.String())
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
	for _, msg := range messages {
		if msg == "" {
			return ErrParamsInvalid.Wrapf("authorizeMessagesAllowed cannot contain empty string literals")
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

func ValidateMaxEthBlockUpdateDelay(i interface{}) error {
	v, ok := i.(time.Duration)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for maxEthBlockUpdateDelay: %T", i)
	}
	if v < 0 {
		return ErrParamsInvalid.Wrapf("tolerance for no Ethereum block syncing cannot be negative")
	}

	return nil
}

func ValidateInjectedEventTxMaxBytes(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for injectedEventTxMaxBytes: %T", i)
	}

	// value cannot be less than MinimumInjectedEventTxMaxBytes, otherwise, we risk never injecting event txs in a block
	if v < MinimumInjectedEventTxMaxBytes {
		return ErrParamsInvalid.Wrapf(
			"injected event tx max bytes cannot be less than %d: given %d", MinimumInjectedEventTxMaxBytes, v,
		)
	}

	return nil
}

func ValidateMaxAuthorizeMessages(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for maxAuthorizeMessages: %T", i)
	}

	// value cannot be less than MinimumMaxAuthorizeMessages, otherwise, we risk skipping all Authorize txs.
	if v < MinimumMaxAuthorizeMessages {
		return ErrParamsInvalid.Wrapf(
			"injected event tx max bytes cannot be less than %d: given %d", MinimumInjectedEventTxMaxBytes, v,
		)
	}

	return nil
}

func ValidateSequencerTxsAllocation(i interface{}) error {
	v, ok := i.(sdkmath.LegacyDec)
	if !ok {
		return ErrParamsInvalid.Wrapf("invalid parameter type for sequencerTxsAllocation: %T", i)
	}

	// value must be within allowed range
	if v.LT(MinimumSequencerTxsAllocation) || v.GT(MaximumSequencerTxsAllocation) {
		return ErrParamsInvalid.Wrapf(
			"expected: %s <= value <= %s; actual %s",
			MinimumSequencerTxsAllocation.String(),
			MaximumSequencerTxsAllocation.String(),
			v.String(),
		)
	}

	return nil
}

// VestingTimesFromVestingDuration returns the vesting start and end time based on the VestingStartTime parameter, the
// vestingStartTimeDelay, and a specified vesting duration. If the specified vesting duration is greater than the
// vestingStartTimeDelay, then the delay is applied as a delay in the start of vesting. Otherwise, the delay is not
// applied, meaning that vesting will start from VestingStartTime.
//
// Example 1: for a VestingStartTime set to 2024-01 and a vesting duration of 2 years:
// - Actual vesting start time: 2024-01 + vestingStartTimeDelay = 2025-01
// - Actual vesting end time: 2024-01 + vesting duration = 2026-01
//
// Example 2: for a VestingStartTime set to 2024-01 and a vesting duration of 6 months:
// - Actual vesting start time: 2024-01
// - Actual vesting end time: 2024-01 + vesting duration = 2024-07
//
// Example 3: for a VestingStartTime set to 2024-01 and a vesting duration of 1 year:
// - Actual vesting start time: 2024-01
// - Actual vesting end time: 2024-01 + vesting duration = 2025-01
func (p Params) VestingTimesFromVestingDuration(duration time.Duration) (time.Time, time.Time, error) {

	if duration == 0 {
		return time.Time{}, time.Time{}, ErrInvalidVestingDuration.Wrapf("expected duration to be greater than 0")
	}

	var vestingStartTime time.Time
	if duration > vestingStartTimeDelay {
		vestingStartTime = p.VestingStartTime.Add(vestingStartTimeDelay)
	} else {
		vestingStartTime = p.VestingStartTime
	}
	vestingEndTime := p.VestingStartTime.Add(duration)
	return vestingStartTime, vestingEndTime, nil
}

// IsAuthorizedMessage returns true if the sdk.Msg TypeURL is present in Params.AuthorizeMessagesAllowed, otherwise,
// returns false
func (p *Params) IsAuthorizedMessage(msg sdk.Msg) bool {
	for _, messageAllowed := range p.AuthorizeMessagesAllowed {
		if messageAllowed == sdk.MsgTypeURL(msg) {
			return true
		}
	}

	return false
}

// IsMsgSupplyDeltaBlock returns true if it's the right height for a MsgSupplyDelta.
func (p *Params) IsMsgSupplyDeltaBlock(block uint64) bool {
	if p.SupplyDeltaPeriod == 0 {
		// We've already validated it at the Validate function.
		panic("supply delta period is zero")
	}
	return block%p.SupplyDeltaPeriod == 0
}
