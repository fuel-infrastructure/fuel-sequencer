package types_test

import (
	"encoding/hex"
	"testing"

	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // ensure bech32 configs are set
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateSequencerAccountFromEthereumAddress(t *testing.T) {

	ethAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddress(ethAddress)
	require.NoError(t, err)

	require.Equal(t, "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k", accAddress.String())
}

func TestGenerateSequencerAccountFromEthereumAddressFromBz(t *testing.T) {

	ethAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	ethAddressBz, err := hex.DecodeString(ethAddress[2:])
	require.NoError(t, err)

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddressFromBz(ethAddressBz)
	require.NoError(t, err)

	require.Equal(t, "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k", accAddress.String())
}
