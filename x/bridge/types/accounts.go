package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	"github.com/ethereum/go-ethereum/common"
)

// GenerateSequencerAccountForEthereumAddress trims the 0x prefix from an Ethereum address, if any,
// and decodes it into bytes before passing it to GenerateSequencerAccountForEthereumAddressFromBz.
func GenerateSequencerAccountForEthereumAddress(ethAddress string) (sdk.AccAddress, error) {
	// TODO: We might want to verify checksum of address
	if !common.IsHexAddress(ethAddress) {
		return nil, errorsmod.Wrapf(ErrInvalidEthAddress, "invalid Ethereum to address format")
	}

	return GenerateSequencerAccountForEthereumAddressFromBz(common.FromHex(ethAddress))
}

// GenerateSequencerAccountForEthereumAddressFromBz derives a Sequencer address from the module name and
// the specified Ethereum address. The module name ensures we do not overlap with other modules' addresses.
func GenerateSequencerAccountForEthereumAddressFromBz(ethAddress []byte) (sdk.AccAddress, error) {
	if len(ethAddress) != 20 {
		return nil, ErrInvalidEthAddressLength.Wrapf("got %d", len(ethAddress))
	}

	return sdkaddress.Module(ModuleName, ethAddress), nil
}
