package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
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

// LockedCoinsFromVesting is identical to the BaseVestingAccount implementation.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L45
func (a *EthOwnedMultiContinuousVestingAccount) LockedCoinsFromVesting(vestingCoins sdk.Coins) sdk.Coins {
	lockedCoins := vestingCoins.Sub(vestingCoins.Min(a.DelegatedVesting)...)
	if lockedCoins == nil {
		return sdk.Coins{}
	}
	return lockedCoins
}

// LockedCoins is identical to the ContinuousVestingAccount implementation.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L231
func (a *EthOwnedMultiContinuousVestingAccount) LockedCoins(blockTime time.Time) sdk.Coins {
	return a.LockedCoinsFromVesting(a.GetVestingCoins(blockTime))
}

// TrackDelegation implements the VestingAccount interface's TrackDelegation function to ensure that the amount being
// delegated is spendable. The delegated amount is added to the DelegatedFree entry.
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
func (a *EthOwnedMultiContinuousVestingAccount) TrackDelegation(blockTime time.Time, balance, amount sdk.Coins) {

	// Sanity check delegation amount - inspired by BaseVestingAccount's TrackDelegation function
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

// TrackUndelegation implements the VestingAccount interface's TrackUndelegation function to mirror the implemented
// TrackDelegation function. The undelegated amount is subtracted from the DelegatedFree entry.
//
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L99
func (a *EthOwnedMultiContinuousVestingAccount) TrackUndelegation(amount sdk.Coins) {

	// Sanity check delegation amount - inspired by BaseVestingAccount's TrackUndelegation function
	if !amount.IsAllPositive() {
		panic(fmt.Sprintf("undelegation attempt with zero amount in coins %s", amount.String()))
	}

	a.DelegatedFree = a.DelegatedFree.Sub(amount...)
}

// GetVestedCoins calculates the sum of all vested coins reported by the vesting infos.
func (a *EthOwnedMultiContinuousVestingAccount) GetVestedCoins(blockTime time.Time) sdk.Coins {
	var vestedCoins sdk.Coins
	for _, info := range a.Infos {
		vestedCoins = vestedCoins.Add(info.GetVestedCoins(blockTime)...)
	}
	return vestedCoins
}

// GetVestingCoins is identical to the ContinuousVestingAccount implementation but gets the aggregate OriginalVesting
// and aggregate VestedCoins from each vesting info.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.10/x/auth/vesting/types/vesting_account.go#L225
func (a *EthOwnedMultiContinuousVestingAccount) GetVestingCoins(blockTime time.Time) sdk.Coins {
	return a.GetOriginalVesting().Sub(a.GetVestedCoins(blockTime)...)
}

// GetStartTime is not expected to be used in production. In the Cosmos SDK it is only used in testing.
func (a *EthOwnedMultiContinuousVestingAccount) GetStartTime() int64 {
	panic("cannot get start time for eth owned multi continuous vesting account")
}

// GetEndTime is not expected to be used in production. In the Cosmos SDK it is only used in testing.
func (a *EthOwnedMultiContinuousVestingAccount) GetEndTime() int64 {
	panic("cannot get end time for eth owned multi continuous vesting account")
}

// GetOriginalVesting calculates the sum of all original vesting coins reported by the vesting infos.
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
