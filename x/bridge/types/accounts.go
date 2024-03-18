// Inspired by https://github.com/cosmos/ibc-go/blob/v8.0.1/modules/apps/27-interchain-accounts/types/account.go

package types

import (
	errorsmod "cosmossdk.io/errors"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/ethereum/go-ethereum/common"

	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var (
	_ authtypes.GenesisAccount = (*EthOwnedAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedAccount)(nil)

	_ authtypes.GenesisAccount = (*EthOwnedContinuousVestingAccount)(nil)
	_ EthOwnedAccountI         = (*EthOwnedContinuousVestingAccount)(nil)
)

// EthOwnedAccountI wraps the sdk.AccountI interface
type EthOwnedAccountI interface {
	sdk.AccountI
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

// --------------------- EthOwnedAccount

// NewEthOwnedAccount creates and returns a new EthOwnedAccount type
func NewEthOwnedAccount(ba *authtypes.BaseAccount) *EthOwnedAccount {
	return &EthOwnedAccount{
		BaseAccount: ba,
	}
}

// NewEthOwnedAccountWithAddress creates and returns a new EthOwnedAccount type from an address
func NewEthOwnedAccountWithAddress(address sdk.AccAddress) *EthOwnedAccount {
	return &EthOwnedAccount{
		BaseAccount: authtypes.NewBaseAccountWithAddress(address),
	}
}

// SetPubKey implements the authtypes.AccountI interface
func (EthOwnedAccount) SetPubKey(_ crypto.PubKey) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set public key for eth owned account")
}

// SetSequence implements the authtypes.AccountI interface
func (EthOwnedAccount) SetSequence(_ uint64) error {
	return errorsmod.Wrap(ErrUnsupported, "cannot set sequence number for eth owned account")
}

// --------------------- EthOwnedContinuousVestingAccount

// NewEthOwnedContinuousVestingAccount creates and returns a new EthOwnedVestingAccount type
func NewEthOwnedContinuousVestingAccount(cva *vestingtypes.ContinuousVestingAccount) *EthOwnedContinuousVestingAccount {
	return &EthOwnedContinuousVestingAccount{
		ContinuousVestingAccount: cva,
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

// ToEthOwnedAccount discards vesting details and converts the account to an EthOwnedAccount
func (a EthOwnedContinuousVestingAccount) ToEthOwnedAccount() *EthOwnedAccount {
	return NewEthOwnedAccount(a.BaseAccount)
}
