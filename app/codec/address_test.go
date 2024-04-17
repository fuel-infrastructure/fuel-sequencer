package codec_test

import (
	"testing"

	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	appcodec "github.com/fuel-infrastructure/fuel-sequencer/app/codec"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"
)

func TestStringToBytes_ComputesExpectedBytesForHex(t *testing.T) {
	codec := appcodec.NewFuelSequencerAddressCodec(sdkAddressCodec.NewBech32Codec(app.AccountAddressPrefix))

	// Compute test Ethereum address to bytes
	bz, err := codec.StringToBytes(testtypes.TestEthAddr1Str)
	require.NoError(t, err)

	// If the operation is reversed, the bytes should compute to the expected Sequencer address
	actualSequencerAddress, err := codec.BytesToString(bz)
	require.NoError(t, err)
	require.Equal(t, testtypes.TestSeqAddr1Str, actualSequencerAddress)
}

func TestStringToBytes_ComputesExpectedBytesForBech32(t *testing.T) {
	// Create a test codec
	codec := appcodec.NewFuelSequencerAddressCodec(sdkAddressCodec.NewBech32Codec(app.AccountAddressPrefix))

	// Compute test Sequencer address to bytes
	bz, err := codec.StringToBytes(testtypes.TestSeqAddr1Str)
	require.NoError(t, err)

	// If the operation is reversed, the bytes should compute to the original Sequencer address
	actualSequencerAddress, err := codec.BytesToString(bz)
	require.NoError(t, err)
	require.Equal(t, testtypes.TestSeqAddr1Str, actualSequencerAddress)
}
