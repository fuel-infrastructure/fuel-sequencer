package types_test

import (
	"encoding/hex"
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateSequencerAccountForEthereumAddress(t *testing.T) {

	ethAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"

	accAddress, err := types.GenerateSequencerAccountForEthereumAddress(ethAddress)
	require.NoError(t, err)

	require.Equal(t, "cosmos13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vq7xmqm6", accAddress.String())
}

func TestGenerateSequencerAccountForEthereumAddressFromBz(t *testing.T) {

	ethAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	ethAddressBz, err := hex.DecodeString(ethAddress[2:])
	require.NoError(t, err)

	accAddress, err := types.GenerateSequencerAccountForEthereumAddressFromBz(ethAddressBz)
	require.NoError(t, err)

	require.Equal(t, "cosmos13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vq7xmqm6", accAddress.String())
}
