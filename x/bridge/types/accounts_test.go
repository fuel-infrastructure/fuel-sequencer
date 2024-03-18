package types_test

import (
	"encoding/hex"
	"testing"

	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // ensure bech32 configs are set
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenerateSequencerAccountFromEthereumAddress(t *testing.T) {

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddress(testutiltypes.TestEthAddr1Str)
	require.NoError(t, err)

	require.Equal(t, testutiltypes.TestSeqAddr1Str, accAddress.String())
}

func TestGenerateSequencerAccountFromEthereumAddressFromBz(t *testing.T) {

	ethAddress := testutiltypes.TestEthAddr1Str
	ethAddressBz, err := hex.DecodeString(ethAddress[2:])
	require.NoError(t, err)

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddressFromBz(ethAddressBz)
	require.NoError(t, err)

	require.Equal(t, testutiltypes.TestSeqAddr1Str, accAddress.String())
}
