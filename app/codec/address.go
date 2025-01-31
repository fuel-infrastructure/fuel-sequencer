package codec

import (
	"errors"
	"strings"

	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdkAddress "github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
)

type FuelSequencerAddressCodec struct {
	*sdkAddressCodec.Bech32Codec
}

var _ address.Codec = FuelSequencerAddressCodec{}

func NewFuelSequencerAddressCodec(codec address.Codec) address.Codec {
	bech32Codec, ok := codec.(*sdkAddressCodec.Bech32Codec)
	if !ok {
		panic("base codec is not Bech32Codec when creating FuelSequencerAddressCodec")
	}

	return FuelSequencerAddressCodec{Bech32Codec: bech32Codec}
}

// StringToBytes encodes text to bytes
func (c FuelSequencerAddressCodec) StringToBytes(text string) ([]byte, error) {
	// Empty addresses strings are not allowed
	if len(strings.TrimSpace(text)) == 0 {
		return []byte{}, errors.New("empty address string is not allowed")
	}

	// If text is a Hex address attempt encoding the Ethereum address to bytes using Ethereum libraries
	if common.IsHexAddress(text) {
		bz := common.FromHex(text)

		// Copied from Bech32Codec
		if len(bz) > sdkAddress.MaxAddrLen {
			return nil, errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "address max length is %d, got %d", sdkAddress.MaxAddrLen, len(bz))
		}

		return bz, nil
	}

	// Otherwise default to the Cosmos SDK default implementation (bech32)
	return c.Bech32Codec.StringToBytes(text)
}
