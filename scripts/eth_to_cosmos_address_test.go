package scripts_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
	"github.com/stretchr/testify/require"
)

func TestConvertHexToCosmosAddressAndViceVersa(t *testing.T) {

	// Ethereum address in hex format
	ethereumAddress := "0x62d221db49aef5632f59b900b2ca90e52ecc0a80"
	ethereumAddress = strings.TrimPrefix(ethereumAddress, "0x")

	// Decode the hex encoded Ethereum address
	addrBz, err := hex.DecodeString(ethereumAddress)
	require.NoError(t, err)

	// Convert byte slice to hexadecimal string
	hexAddress := hex.EncodeToString(addrBz)
	fmt.Printf("Hex Address:     0x%s\n", hexAddress)

	valoperAddress, err := sdk.ValAddressFromHex(hexAddress)
	require.NoError(t, err)
	fmt.Printf("Valoper Address: %s\n", valoperAddress)

	accAddr, err := sdk.AccAddressFromHexUnsafe(hexAddress)
	require.NoError(t, err)
	fmt.Printf("Account Address: %s\n", accAddr.String())
}
