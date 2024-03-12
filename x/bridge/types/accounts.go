package types

import (
	"encoding/hex"
	"regexp"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
)

// isValidEthAddr defines a regular expression to check if the provided
// string is a valid Ethereum address, excluding the 0x prefix.
var isValidEthAddr = regexp.MustCompile("^[a-fA-F0-9]{40}$").MatchString

// GenerateSequencerAccountForEthereumAddress trims the 0x prefix from an Ethereum address, if any,
// and decodes it into bytes before passing it to GenerateSequencerAccountForEthereumAddressFromBz.
func GenerateSequencerAccountForEthereumAddress(ethAddress string) (sdk.AccAddress, error) {
	if strings.HasPrefix(ethAddress, "0x") {
		ethAddress = ethAddress[2:] // trim 0x prefix
	}
	if !isValidEthAddr(ethAddress) {
		return nil, ErrInvalidAddress.Wrapf("string is not valid ethereum address: %s", ethAddress)
	}

	ethAddressBz, err := hex.DecodeString(ethAddress)
	if err != nil {
		return nil, err
	}
	return GenerateSequencerAccountForEthereumAddressFromBz(ethAddressBz)
}

// GenerateSequencerAccountForEthereumAddressFromBz derives a Sequencer address from the module name and
// the specified Ethereum address. The module name ensures we do not overlap with other modules' addresses.
func GenerateSequencerAccountForEthereumAddressFromBz(ethAddress []byte) (sdk.AccAddress, error) {
	if len(ethAddress) != 20 {
		return nil, ErrInvalidEthAddressLength.Wrapf("got %d", len(ethAddress))
	}

	return sdkaddress.Module(ModuleName, ethAddress), nil
}
