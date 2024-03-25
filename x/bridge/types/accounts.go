// Inspired by https://github.com/cosmos/ibc-go/blob/v8.0.1/modules/apps/27-interchain-accounts/types/account.go

package types

import (
	"encoding/json"
	"strings"

	errorsmod "cosmossdk.io/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/yaml.v2"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ authtypes.GenesisAccount = (*EthOwnedBaseAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedBaseAccount)(nil)

	_ authtypes.GenesisAccount = (*EthOwnedContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedContinuousVestingAccount)(nil)
)

// EthOwnedAccountI wraps the sdk.AccountI interface
type EthOwnedAccountI interface {
	sdk.AccountI
}

// ethOwnedAccountPretty defines an unexported struct used for encoding the EthOwnedAccount details
type ethOwnedAccountPretty struct {
	Address       sdk.AccAddress `json:"address" yaml:"address"`
	PubKey        string         `json:"public_key" yaml:"public_key"`
	AccountNumber uint64         `json:"account_number" yaml:"account_number"`
	Sequence      uint64         `json:"sequence" yaml:"sequence"`
	AccountOwner  string         `json:"account_owner" yaml:"account_owner"`
}

// GenerateSequencerAddressFromEthereumAddress trims the 0x prefix from an Ethereum address, if any,
// and decodes it into bytes before passing it to GenerateSequencerAddressFromEthereumAddressFromBz.
func GenerateSequencerAddressFromEthereumAddress(ethAddress string) (sdk.AccAddress, error) {
	// TODO: We might want to verify checksum of address
	if !common.IsHexAddress(ethAddress) {
		return nil, errorsmod.Wrapf(ErrInvalidEthAddress, "invalid Ethereum address format (%s)", ethAddress)
	}

	return GenerateSequencerAddressFromEthereumAddressFromBz(common.FromHex(ethAddress))
}

// GenerateSequencerAddressFromEthereumAddressFromBz derives a Sequencer address from the module name and
// the specified Ethereum address. The module name ensures we do not overlap with other modules' addresses.
func GenerateSequencerAddressFromEthereumAddressFromBz(ethAddress []byte) (sdk.AccAddress, error) {
	if len(ethAddress) != common.AddressLength {
		return nil, ErrInvalidEthAddressLength.Wrapf("got %d", len(ethAddress))
	}

	return sdkaddress.Module(ModuleName, ethAddress), nil
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

// NewEthOwnedContinuousVestingAccount creates and returns a new EthOwnedVestingAccount type
func NewEthOwnedContinuousVestingAccount(
	cva *vestingtypes.ContinuousVestingAccount, owner string,
) *EthOwnedContinuousVestingAccount {
	return &EthOwnedContinuousVestingAccount{
		ContinuousVestingAccount: cva,
		AccountOwner:             owner,
	}
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
func (a EthOwnedContinuousVestingAccount) ToEthOwnedBaseAccount() *EthOwnedBaseAccount {
	return NewEthOwnedBaseAccount(a.BaseAccount, a.AccountOwner)
}
