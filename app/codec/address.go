package codec

import (
	"cosmossdk.io/core/address"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/ethereum/go-ethereum/common"
)

type FuelSequencerAddressCodec struct {
	sdkAddressCodec.Bech32Codec
}

var _ address.Codec = FuelSequencerAddressCodec{}

func NewFuelSequencerAddressCodec(codec address.Codec) address.Codec {
	bech32Codec, ok := codec.(sdkAddressCodec.Bech32Codec)
	if !ok {
		panic("could not create FuelSequencerAddressCodec")
	}

	return FuelSequencerAddressCodec{Bech32Codec: bech32Codec}
}

// StringToBytes encodes text to bytes
func (c FuelSequencerAddressCodec) StringToBytes(text string) ([]byte, error) {
	// If text is a Hex address attempt encoding the Ethereum address to bytes using Ethereum libraries
	if common.IsHexAddress(text) {
		return common.FromHex(text), nil
	}

	// Otherwise default to the Cosmos SDK default implementation (bech32)
	return c.Bech32Codec.StringToBytes(text)
}

// BytesToString decodes bytes to text
func (c FuelSequencerAddressCodec) BytesToString(bz []byte) (string, error) {

	// Bytes will always be decoded to bech32
	return c.Bech32Codec.BytesToString(bz)
}
