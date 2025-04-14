package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"gopkg.in/yaml.v2"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ sdk.AccountI             = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ authtypes.GenesisAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ banktypes.VestingAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ exported.VestingAccount  = (*EthOwnedMultiContinuousVestingAccount)(nil)
)

// NewEthOwnedMultiContinuousVestingAccount creates and returns a new EthOwnedMultiContinuousVestingAccount type
func NewEthOwnedMultiContinuousVestingAccount(
	baseAccount *authtypes.BaseAccount,
	infos []*VestingInfo,
	owner string,
) *EthOwnedMultiContinuousVestingAccount {
	return &EthOwnedMultiContinuousVestingAccount{
		BaseAccount:      baseAccount,
		Infos:            infos,
		DelegatedFree:    nil,
		DelegatedVesting: nil,
		AccountOwner:     owner,
	}
}

func NewEthOwnedMultiContinuousVestingAccountWithDelegation(
	baseAccount *authtypes.BaseAccount,
	infos []*VestingInfo,
	delegatedFree sdk.Coins,
	delegatedVesting sdk.Coins,
	owner string,
) *EthOwnedMultiContinuousVestingAccount {
	return &EthOwnedMultiContinuousVestingAccount{
		BaseAccount:      baseAccount,
		Infos:            infos,
		DelegatedFree:    delegatedFree,
		DelegatedVesting: delegatedVesting,
		AccountOwner:     owner,
	}
}

func NewVestingInfo(originalVesting sdk.Coins, startTime, endTime int64) *VestingInfo {
	return &VestingInfo{
		OriginalVesting: originalVesting,
		StartTime:       startTime,
		EndTime:         endTime,
	}
}

// ------------------------------------ EthOwnedAccountI implementations

// AddVestingCoins TODO
func (a *EthOwnedMultiContinuousVestingAccount) AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error) {

	startTimeUnix := startTime.Unix()
	endTimeUnix := endTime.Unix()

	for _, info := range a.Infos {
		if info.StartTime == startTimeUnix && info.EndTime == endTimeUnix {
			info.OriginalVesting = info.OriginalVesting.Add(coins...)
			return a, nil
		}
	}

	a.Infos = append(a.Infos, NewVestingInfo(coins, startTimeUnix, endTimeUnix))
	return a, nil
}

// ------------------------------------ VestingAccount implementations

func (a *EthOwnedMultiContinuousVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(a.DelegatedVesting)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

func (a *EthOwnedMultiContinuousVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return a.LockedCoinsFromVesting(a.GetVestingCoins(blockTime))
}

// TrackDelegation TODO
func (a *EthOwnedMultiContinuousVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackDelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("delegation attempt with zero amount in coins %s", amount.String()))
	}

	// Calculate spendable coins, where balance will only ever be an amount in FUEL.
	// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/keeper/view.go#L212
	locked := a.LockedCoins(blockTime)
	spendable, hasNeg := balance.SafeSub(locked...)
	if hasNeg {
		spendable = sdk.NewCoins()
	}

	// Delegation amount must be spendable
	if !spendable.IsAllGTE(amount) {
		panic(fmt.Sprintf("cannot delegate locked coins; max spendable is %s", spendable.String()))
	}

	a.DelegatedFree = a.DelegatedFree.Add(amount...)
}

// TrackUndelegation TODO
func (a *EthOwnedMultiContinuousVestingAccount) TrackUndelegation(amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackUndelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("undelegation attempt with zero amount in coins %s", amount.String()))
	}

	a.DelegatedFree = a.DelegatedFree.Sub(amount...)
}

// GetVestedCoins TODO
func (a *EthOwnedMultiContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins

	for _, info := range a.Infos {
		// We must handle the case where the start time for a vesting account has
		// been set into the future or when the start of the chain is not exactly
		// known.
		if blockTime.Unix() <= info.StartTime {
			return vestedCoins
		} else if blockTime.Unix() >= info.EndTime {
			return info.OriginalVesting
		}

		// calculate the vesting scalar
		x := blockTime.Unix() - info.StartTime
		y := info.EndTime - info.StartTime
		s := math.LegacyNewDec(x).Quo(math.LegacyNewDec(y))

		for _, ovc := range info.OriginalVesting {
			vestedAmt := math.LegacyNewDecFromInt(ovc.Amount).Mul(s).RoundInt()
			vestedCoins = append(vestedCoins, sdk.NewCoin(ovc.Denom, vestedAmt))
		}
	}

	return vestedCoins
}

func (a *EthOwnedMultiContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return a.GetOriginalVesting().Sub(a.GetVestedCoins(blockTime)...)
}

func (a *EthOwnedMultiContinuousVestingAccount) GetStartTime() int64 {
	panic("cannot get start time for eth owned multi continuous vesting account")
}

func (a *EthOwnedMultiContinuousVestingAccount) GetEndTime() int64 {
	panic("cannot get end time for eth owned multi continuous vesting account")
}

func (a *EthOwnedMultiContinuousVestingAccount) GetOriginalVesting() sdk.Coins {
	var originalVesting sdk.Coins

	for _, info := range a.Infos {
		originalVesting = originalVesting.Add(info.OriginalVesting...)
	}

	return originalVesting
}

func (a *EthOwnedMultiContinuousVestingAccount) GetDelegatedFree() sdk.Coins {
	return a.DelegatedFree
}

func (a *EthOwnedMultiContinuousVestingAccount) GetDelegatedVesting() sdk.Coins {
	return a.DelegatedVesting
}

// ------------------------------------ AccountI implementations

func (EthOwnedMultiContinuousVestingAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned multi continuous vesting account")
}

func (EthOwnedMultiContinuousVestingAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned multi continuous vesting account")
}

// ------------------------------------ GenesisAccount implementations

// Validate implements basic validation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	return a.BaseAccount.Validate()
}

// ------------------------------------ Miscellaneous implementations

// String returns a string representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) String() string {
	out, _ := a.MarshalYAML()
	return string(out)
}

// MarshalYAML returns the YAML representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) MarshalYAML() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(a.Address)
	if err != nil {
		return nil, err
	}

	bz, err := yaml.Marshal(ethOwnedAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: a.AccountNumber,
		Sequence:      a.Sequence,
		AccountOwner:  a.AccountOwner,
	})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// MarshalJSON returns the JSON representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) MarshalJSON() ([]byte, error) {
	accAddr, err := sdk.AccAddressFromBech32(a.Address)
	if err != nil {
		return nil, err
	}

	bz, err := json.Marshal(ethOwnedAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: a.AccountNumber,
		Sequence:      a.Sequence,
		AccountOwner:  a.AccountOwner,
	})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// UnmarshalJSON unmarshals raw JSON bytes into the EthOwnedMultiContinuousVestingAccount
func (a *EthOwnedMultiContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
	var alias ethOwnedAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	a.BaseAccount = authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	a.AccountOwner = alias.AccountOwner

	return nil
}

// ToEthOwnedBaseAccount discards vesting details and converts the account to an EthOwnedBaseAccount
// This is mostly intended for testing where we might want to switch the account type.
func (a EthOwnedMultiContinuousVestingAccount) ToEthOwnedBaseAccount() *EthOwnedBaseAccount {
	return NewEthOwnedBaseAccount(a.BaseAccount, a.AccountOwner)
}
