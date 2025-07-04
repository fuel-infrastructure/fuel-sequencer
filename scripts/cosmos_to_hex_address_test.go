package scripts_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // Required to load the right config for testing
)

func TestConvertCosmosToHexAddressAndViceVersa(t *testing.T) {

	// Cosmos account or valoper address in bech32 format
	cosmosAddress := "fuelsequencervaloper1vtfzrk6f4m6kxt6ehyqt9j5su5hvcz5qn0wwpn"

	// Decode the bech32 encoded Cosmos address
	var addrBz []byte
	valoper := sdk.PrefixValidator + sdk.PrefixOperator
	if strings.Contains(cosmosAddress, valoper) {
		valAddr, err := sdk.ValAddressFromBech32(cosmosAddress)
		require.NoError(t, err)
		addrBz = valAddr.Bytes()
	} else {
		accAddr, err := sdk.AccAddressFromBech32(cosmosAddress)
		require.NoError(t, err)
		addrBz = accAddr.Bytes()
	}

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
