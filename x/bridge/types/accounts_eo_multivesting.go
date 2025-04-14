package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"gopkg.in/yaml.v2"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ authtypes.GenesisAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ banktypes.VestingAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ exported.VestingAccount  = (*EthOwnedMultiContinuousVestingAccount)(nil)
)

// NewEthOwnedMultiContinuousVestingAccount creates and returns a new EthOwnedMultiContinuousVestingAccount type
func NewEthOwnedMultiContinuousVestingAccount(
	cvas []*vestingtypes.ContinuousVestingAccount, owner string,
) *EthOwnedMultiContinuousVestingAccount {
	return &EthOwnedMultiContinuousVestingAccount{
		VestingAccounts: cvas,
		AccountOwner:    owner,
	}
}

// TrackDelegation TODO
func (a *EthOwnedMultiContinuousVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackDelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("delegation attempt with zero amount in coins %s", amount.String()))
	}

	// Calculate total locked.
	locked := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		locked = locked.Add(vacc.LockedCoins(blockTime)...)
	}

	// Calculate spendable coins.
	// Similar to https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/keeper/view.go#L212
	spendable, hasNeg := balance.SafeSub(locked...)
	if hasNeg {
		spendable = sdk.NewCoins()
	}

	// Delegation amount must be spendable
	if !spendable.IsAllGTE(amount) {
		panic(fmt.Sprintf("cannot delegate locked coins; max spendable is %s", spendable.String()))
	}

	// Always use the first vesting account to track delegations.
	a.VestingAccounts[0].DelegatedFree = a.VestingAccounts[0].DelegatedFree.Add(amount...)
}

// TrackUndelegation TODO
func (a *EthOwnedMultiContinuousVestingAccount) TrackUndelegation(amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackUndelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("undelegation attempt with zero amount in coins %s", amount.String()))
	}

	// Always use the first vesting account to track delegations.
	a.VestingAccounts[0].DelegatedFree = a.VestingAccounts[0].DelegatedFree.Sub(amount...)
}

// AddVestingCoins TODO
func (a *EthOwnedMultiContinuousVestingAccount) AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error) {

	startTimeUnix := startTime.Unix()
	endTimeUnix := endTime.Unix()

	for _, infos := range a.VestingAccounts {
		if infos.StartTime == startTimeUnix && infos.EndTime == endTimeUnix {
			infos.OriginalVesting = infos.OriginalVesting.Add(coins...)
			return a, nil
		}
	}

	baseAcc := a.VestingAccounts[0].BaseAccount
	newSubAcc, err := vestingtypes.NewContinuousVestingAccount(baseAcc, coins, startTimeUnix, endTimeUnix)
	if err != nil {
		return nil, err
	}
	a.VestingAccounts = append(a.VestingAccounts, newSubAcc)

	return a, nil
}

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedMultiContinuousVestingAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned multi continuous vesting account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedMultiContinuousVestingAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned multi continuous vesting account")
}

// Validate implements basic validation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	for _, vacc := range a.VestingAccounts {
		err := vacc.BaseAccount.Validate()
		if err != nil {
			return err
		}
	}
	return nil
}

// String returns a string representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) String() string {
	out, _ := a.MarshalYAML()
	return string(out)
}

// MarshalYAML returns the YAML representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) MarshalYAML() ([]byte, error) {
	baseAcc := a.VestingAccounts[0].BaseAccount
	accAddr, err := sdk.AccAddressFromBech32(baseAcc.Address)
	if err != nil {
		return nil, err
	}

	bz, err := yaml.Marshal(ethOwnedAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: baseAcc.AccountNumber,
		Sequence:      baseAcc.Sequence,
		AccountOwner:  a.AccountOwner,
	})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// MarshalJSON returns the JSON representation of the EthOwnedMultiContinuousVestingAccount
func (a EthOwnedMultiContinuousVestingAccount) MarshalJSON() ([]byte, error) {
	baseAcc := a.VestingAccounts[0].BaseAccount
	accAddr, err := sdk.AccAddressFromBech32(baseAcc.Address)
	if err != nil {
		return nil, err
	}

	bz, err := json.Marshal(ethOwnedAccountPretty{
		Address:       accAddr,
		PubKey:        "",
		AccountNumber: baseAcc.AccountNumber,
		Sequence:      baseAcc.Sequence,
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

	a.VestingAccounts = []*vestingtypes.ContinuousVestingAccount{
		{
			BaseVestingAccount: &vestingtypes.BaseVestingAccount{
				BaseAccount: authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence),
			},
		},
	}
	a.AccountOwner = alias.AccountOwner

	return nil
}

// ToEthOwnedBaseAccount discards vesting details and converts the account to an EthOwnedBaseAccount
// This is mostly intended for testing where we might want to switch the account type.
func (a EthOwnedMultiContinuousVestingAccount) ToEthOwnedBaseAccount() *EthOwnedBaseAccount {
	return NewEthOwnedBaseAccount(a.VestingAccounts[0].BaseAccount, a.AccountOwner)
}

func (a *EthOwnedMultiContinuousVestingAccount) GetAddress() sdk.AccAddress {
	return a.VestingAccounts[0].GetAddress()
}

func (a *EthOwnedMultiContinuousVestingAccount) SetAddress(address sdk.AccAddress) error {
	for _, vacc := range a.VestingAccounts {
		err := vacc.SetAddress(address)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *EthOwnedMultiContinuousVestingAccount) GetPubKey() crypto.PubKey {
	return a.VestingAccounts[0].GetPubKey()
}

func (a *EthOwnedMultiContinuousVestingAccount) GetAccountNumber() uint64 {
	return a.VestingAccounts[0].GetAccountNumber()
}

func (a *EthOwnedMultiContinuousVestingAccount) SetAccountNumber(u uint64) error {
	for _, vacc := range a.VestingAccounts {
		err := vacc.SetAccountNumber(u)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *EthOwnedMultiContinuousVestingAccount) GetSequence() uint64 {
	return a.VestingAccounts[0].GetSequence()
}

func (a *EthOwnedMultiContinuousVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	lockedCoins := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		lockedCoins = lockedCoins.Add(vacc.LockedCoins(blockTime)...)
	}
	return lockedCoins
}

func (a *EthOwnedMultiContinuousVestingAccount) GetOriginalVesting() sdk.Coins {
	originalVesting := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		originalVesting = originalVesting.Add(vacc.OriginalVesting...)
	}
	return originalVesting
}

func (a *EthOwnedMultiContinuousVestingAccount) GetDelegatedFree() sdk.Coins {
	delegatedFree := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		delegatedFree = delegatedFree.Add(vacc.DelegatedFree...)
	}
	return delegatedFree
}

func (a *EthOwnedMultiContinuousVestingAccount) GetDelegatedVesting() sdk.Coins {
	delegatedVesting := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		delegatedVesting = delegatedVesting.Add(vacc.DelegatedVesting...)
	}
	return delegatedVesting
}

func (a *EthOwnedMultiContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	vestedCoins := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		vestedCoins = vestedCoins.Add(vacc.GetVestedCoins(blockTime)...)
	}
	return vestedCoins
}

func (a *EthOwnedMultiContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	vestingCoins := sdk.NewCoins()
	for _, vacc := range a.VestingAccounts {
		vestingCoins = vestingCoins.Add(vacc.GetVestingCoins(blockTime)...)
	}
	return vestingCoins
}

func (a *EthOwnedMultiContinuousVestingAccount) GetStartTime() int64 {
	panic("cannot get start time for eth owned multi continuous vesting account")
}

func (a *EthOwnedMultiContinuousVestingAccount) GetEndTime() int64 {
	panic("cannot get end time for eth owned multi continuous vesting account")
}
