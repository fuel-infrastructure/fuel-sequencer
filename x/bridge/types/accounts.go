// Inspired by https://github.com/cosmos/ibc-go/blob/v8.0.1/modules/apps/27-interchain-accounts/types/account.go

package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"gopkg.in/yaml.v2"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ authtypes.GenesisAccount = (*EthOwnedBaseAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedBaseAccount)(nil)

	_ authtypes.GenesisAccount = (*EthOwnedContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedContinuousVestingAccount)(nil)
	_ banktypes.VestingAccount = (*EthOwnedContinuousVestingAccount)(nil)

	_ authtypes.GenesisAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedMultiContinuousVestingAccount)(nil)
	_ banktypes.VestingAccount = (*EthOwnedMultiContinuousVestingAccount)(nil)
)

// EthOwnedAccountI wraps the sdk.AccountI interface
type EthOwnedAccountI interface {
	sdk.AccountI

	// AddVestingCoins adds new coins with the specified vesting start and end times.
	// If the account is already a vesting account, the start and end time should be ignored.
	AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error)
}

// ethOwnedAccountPretty defines an unexported struct used for encoding the EthOwnedAccount details
type ethOwnedAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	AccountOwner  string         `json:"account_owner" yaml:"account_owner"`
}

// --------------------- EthOwnedBaseAccount

// NewEthOwnedBaseAccount creates and returns a new EthOwnedBaseAccount type
func NewEthOwnedBaseAccount(ba *authtypes.BaseAccount, owner string) *EthOwnedBaseAccount {
	return &EthOwnedBaseAccount{
		BaseAccount:  ba,
		AccountOwner: owner,
	}
}

// NewEthOwnedBaseAccountWithAddress creates and returns a new EthOwnedBaseAccount type from an address
func NewEthOwnedBaseAccountWithAddress(address sdk.AccAddress, owner string) *EthOwnedBaseAccount {
	return &EthOwnedBaseAccount{
		BaseAccount:  authtypes.NewBaseAccountWithAddress(address),
		AccountOwner: owner,
	}
}

// AddVestingCoins converts the EthOwnedBaseAccount into an EthOwnedContinuousVestingAccount with the specified coins
// as the original vesting amount and the specified start and end times. Any coins that were already in this account
// will still remain available since we're not considering them when setting the vesting amount.
func (a EthOwnedBaseAccount) AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error) {
	newVestingAcc, err := vestingtypes.NewContinuousVestingAccount(
		a.BaseAccount, coins, startTime.Unix(), endTime.Unix(),
	)
	if err != nil {
		return nil, err
	}
	return NewEthOwnedContinuousVestingAccount(newVestingAcc, a.AccountOwner), nil
}

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedBaseAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedBaseAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned account")
}

// Validate implements basic validation of the EthOwnedBaseAccount
func (a EthOwnedBaseAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	return a.BaseAccount.Validate()
}

// String returns a string representation of the EthOwnedBaseAccount
func (a EthOwnedBaseAccount) String() string {
	out, _ := a.MarshalYAML()
	return string(out)
}

// MarshalYAML returns the YAML representation of the EthOwnedBaseAccount
func (a EthOwnedBaseAccount) MarshalYAML() ([]byte, error) {
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

// MarshalJSON returns the JSON representation of the EthOwnedBaseAccount
func (a EthOwnedBaseAccount) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON unmarshals raw JSON bytes into the EthOwnedBaseAccount
func (a *EthOwnedBaseAccount) UnmarshalJSON(bz []byte) error {
	var alias ethOwnedAccountPretty
	if err := json.Unmarshal(bz, &alias); err != nil {
		return err
	}

	a.BaseAccount = authtypes.NewBaseAccount(alias.Address, nil, alias.AccountNumber, alias.Sequence)
	a.AccountOwner = alias.AccountOwner

	return nil
}

// --------------------- EthOwnedContinuousVestingAccount

// NewEthOwnedContinuousVestingAccount creates and returns a new EthOwnedContinuousVestingAccount type
func NewEthOwnedContinuousVestingAccount(
	cva *vestingtypes.ContinuousVestingAccount, owner string,
) *EthOwnedContinuousVestingAccount {
	return &EthOwnedContinuousVestingAccount{
		ContinuousVestingAccount: cva,
		AccountOwner:             owner,
	}
}

// TrackDelegation overrides the ContinuousVestingAccount TrackDelegation (which uses the BaseVestingAccount one) to
// ensure that the amount being delegated is spendable. The delegated amount is added to the DelegatedFree entry.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L59
//
// A note about the use of DelegatedVesting in LockedCoinsFromVesting: this function is calculating how many of the
// tokens from the address balance are locked. DelegatedVesting tokens are not in the balance, but vestingCoins are.
// If the address has (i) 100 tokens in the balance, (ii) 200 tokens vesting, and (iii) 150 tokens DelegatedVesting,
// then only 50 (200-150) of the vesting tokens are considered "locked", from the 100 tokens that are in the balance.
//
// Due to our implementation below, we expect that all the vesting tokens will be considered "locked". If we dry-run
// LockedCoinsFromVesting, we will find that the result will always be equal to vestingCoins for DelegatedVesting = 0.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L45
func (a *EthOwnedContinuousVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {

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

// TrackUndelegation overrides the CointinuousVestingAccount TrackUndelegation (which uses the BaseVestingAccount one)
// to mirror the overridden TrackDelegation function. The undelegated amount is subtracted from the DelegatedFree entry.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L99
func (a *EthOwnedContinuousVestingAccount) TrackUndelegation(amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackUndelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("undelegation attempt with zero amount in coins %s", amount.String()))
	}

	a.DelegatedFree = a.DelegatedFree.Sub(amount...)
}

// AddVestingCoins TODO
func (a *EthOwnedContinuousVestingAccount) AddVestingCoins(coins sdk.Coins, startTime, endTime time.Time) (EthOwnedAccountI, error) {

	if a.StartTime == startTime.Unix() && a.EndTime == endTime.Unix() {
		a.OriginalVesting = a.OriginalVesting.Add(coins...)
	} else {
		multiVestingAcc := &EthOwnedMultiContinuousVestingAccount{
			VestingAccounts: []*vestingtypes.ContinuousVestingAccount{a.ContinuousVestingAccount},
			AccountOwner:    a.AccountOwner,
		}
		return multiVestingAcc.AddVestingCoins(coins, startTime, endTime)
	}
	return a, nil
}

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedContinuousVestingAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned continuous vesting account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedContinuousVestingAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned continuous vesting account")
}

// Validate implements basic validation of the EthOwnedContinuousVestingAccount
func (a EthOwnedContinuousVestingAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	return a.BaseAccount.Validate()
}

// String returns a string representation of the EthOwnedContinuousVestingAccount
func (a EthOwnedContinuousVestingAccount) String() string {
	out, _ := a.MarshalYAML()
	return string(out)
}

// MarshalYAML returns the YAML representation of the EthOwnedContinuousVestingAccount
func (a EthOwnedContinuousVestingAccount) MarshalYAML() ([]byte, error) {
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

// MarshalJSON returns the JSON representation of the EthOwnedContinuousVestingAccount
func (a EthOwnedContinuousVestingAccount) MarshalJSON() ([]byte, error) {
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

// UnmarshalJSON unmarshals raw JSON bytes into the EthOwnedContinuousVestingAccount
func (a *EthOwnedContinuousVestingAccount) UnmarshalJSON(bz []byte) error {
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
func (a EthOwnedContinuousVestingAccount) ToEthOwnedBaseAccount() *EthOwnedBaseAccount {
	return NewEthOwnedBaseAccount(a.BaseAccount, a.AccountOwner)
}

// --------------------- EthOwnedContinuousVestingAccount

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

	for _, vacc := range a.VestingAccounts {
		// Calculate spendable coins, where balance will only ever be an amount in FUEL.
		// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/bank/keeper/view.go#L212
		locked := vacc.LockedCoins(blockTime)
		spendable, hasNeg := balance.SafeSub(locked...)
		if hasNeg {
			spendable = sdk.NewCoins()
		}

		// The delegateable amount is whatever we can spend from the spendable.
		canDelegate := spendable.Min(amount)

		vacc.DelegatedFree = vacc.DelegatedFree.Add(canDelegate...)
		amount = amount.Sub(canDelegate...)

		// If no more to delegate, done
		if amount.IsZero() {
			break
		}
	}

	if !amount.IsZero() {
		panic(fmt.Sprintf("expected %s more to be spendable", amount.String()))
	}
}

// TrackUndelegation TODO
func (a *EthOwnedMultiContinuousVestingAccount) TrackUndelegation(amount sdk.Coins) {

	// Sanity check delegation amount - inspired by overridden TrackUndelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("undelegation attempt with zero amount in coins %s", amount.String()))
	}

	for _, vacc := range a.VestingAccounts {
		canUndelegate := vacc.DelegatedFree.Min(amount)
		vacc.DelegatedFree = vacc.DelegatedFree.Sub(canUndelegate...)
		amount = amount.Sub(canUndelegate...)

		if amount.IsZero() {
			break
		}
	}

	if !amount.IsZero() {
		panic(fmt.Sprintf("expected %s more to be delegated", amount.String()))
	}
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
