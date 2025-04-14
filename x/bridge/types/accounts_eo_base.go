package types

import (
	"encoding/json"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"gopkg.in/yaml.v2"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ sdk.AccountI             = (*EthOwnedBaseAccount)(nil)
	_ authtypes.GenesisAccount = (*EthOwnedBaseAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedBaseAccount)(nil)
)

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

// ------------------------------------ EthOwnedAccountI implementations

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

// ------------------------------------ AccountI implementations

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedBaseAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedBaseAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned account")
}

// ------------------------------------ GenesisAccount implementations

// Validate implements basic validation of the EthOwnedBaseAccount
func (a EthOwnedBaseAccount) Validate() error {
	if strings.TrimSpace(a.AccountOwner) == "" {
		return errorsmod.Wrap(ErrInvalidAccountAddress, "AccountOwner cannot be empty")
	}
	return a.BaseAccount.Validate()
}

// ------------------------------------ Miscellaneous implementations

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
