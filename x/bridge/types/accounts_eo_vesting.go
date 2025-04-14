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
	_ sdk.AccountI             = (*EthOwnedContinuousVestingAccount)(nil)
	_ authtypes.GenesisAccount = (*EthOwnedContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedContinuousVestingAccount)(nil)
	_ banktypes.VestingAccount = (*EthOwnedContinuousVestingAccount)(nil)
	_ exported.VestingAccount  = (*EthOwnedContinuousVestingAccount)(nil)
)

// NewEthOwnedContinuousVestingAccount creates and returns a new EthOwnedContinuousVestingAccount type
func NewEthOwnedContinuousVestingAccount(
	cva *vestingtypes.ContinuousVestingAccount, owner string,
) *EthOwnedContinuousVestingAccount {
	return &EthOwnedContinuousVestingAccount{
		ContinuousVestingAccount: cva,
		AccountOwner:             owner,
	}
}

// ------------------------------------ EthOwnedAccountI implementations

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

// ------------------------------------ VestingAccount implementations

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

// ------------------------------------ AccountI implementations

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedContinuousVestingAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned continuous vesting account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedContinuousVestingAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned continuous vesting account")
}

// ------------------------------------ GenesisAccount implementations

// Validate implements basic validation of the EthOwnedContinuousVestingAccount
func (a EthOwnedContinuousVestingAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	return a.BaseAccount.Validate()
}

// ------------------------------------ Miscellaneous implementations

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
