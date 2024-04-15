package utils_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/stretchr/testify/require"
)

func TestConvertCosmosToEthAddress(t *testing.T) {
	app.InitSDKConfig()

	// Cosmos valoper address in bech32 format
	cosmosAddress := "fuelsequencervaloper1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5qn0wwpn"

	// Decode the bech32 encoded Cosmos address
	valAddr, err := sdk.ValAddressFromBech32(cosmosAddress)
	require.NoError(t, err)

	// Convert byte slice to hexadecimal string
	hexAddress := hex.EncodeToString(valAddr.Bytes())

	// Ensure the Ethereum address is exactly 20 bytes (40 characters in hex)
	if len(hexAddress) > 40 {
		hexAddress = hexAddress[:40]
	}

	fmt.Printf("Ethereum Address: 0x%s\n", hexAddress)

	valoperAddress, err := sdk.ValAddressFromHex(hexAddress)
	require.NoError(t, err)

	fmt.Printf("valoper Address: %s\n", valoperAddress)

	accAddr, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	fmt.Printf("non-valoper Address: %s\n", accAddr.String())
}
