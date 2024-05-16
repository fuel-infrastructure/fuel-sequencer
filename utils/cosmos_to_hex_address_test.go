package utils_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	"github.com/stretchr/testify/require"
)

func TestConvertCosmosToHexAddressAndViceVersa(t *testing.T) {

	// Cosmos valoper address in bech32 format
	cosmosAddress := "fuelsequencervaloper1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5qn0wwpn"

	// Decode the bech32 encoded Cosmos address
	valAddr, err := sdk.ValAddressFromBech32(cosmosAddress)
	require.NoError(t, err)

	// Convert byte slice to hexadecimal string
	hexAddress := hex.EncodeToString(valAddr.Bytes())

	fmt.Printf("Hex Address: 0x%s\n", hexAddress)

	valoperAddress, err := sdk.ValAddressFromHex(hexAddress)
	require.NoError(t, err)

	fmt.Printf("valoper Address: %s\n", valoperAddress)

	accAddr, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)

	fmt.Printf("non-valoper Address: %s\n", accAddr.String())
}
