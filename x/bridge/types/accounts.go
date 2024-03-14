package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/ethereum/go-ethereum/common"
)

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
	if len(ethAddress) != 20 {
		return nil, ErrInvalidEthAddressLength.Wrapf("got %d", len(ethAddress))
	}

	return sdkaddress.Module(ModuleName, ethAddress), nil
}
