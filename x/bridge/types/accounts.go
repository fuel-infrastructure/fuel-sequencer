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

func GenerateCrosschainAccountAddress(ethAddress string) (sdk.AccAddress, error) {
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
	return GenerateCrosschainAccountAddressFromBz(ethAddressBz)
}

func GenerateCrosschainAccountAddressFromBz(ethAddress []byte) (sdk.AccAddress, error) {
	if len(ethAddress) != 20 {
		return nil, ErrInvalidAddress.Wrapf("expected eth address to be 20 bytes long, got %d", len(ethAddress))
	}

	ethAccount := sdkaddress.Module(ModuleName, []byte(ethAccountsKey))
	return sdkaddress.Derive(ethAccount, []byte(ethAddress)), nil
}
