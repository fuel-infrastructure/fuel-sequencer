package testsuite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEthFromMnemonic(t *testing.T) {
	mnemonic := "receive roof marine sure lady hundred sea enact exist place bean wagon kingdom betray science photo loop funny bargain floor suspect only strike endless"
	address := "0x14fdAC734De10065093C4Ed4a83C41638378005A"

	generatedKey, err := newEthereumKeyFromMnemonic(mnemonic)
	require.NoError(t, err, "error generating ethereum key")

	require.Equal(t, address, generatedKey.Address)
}

func TestEthFromMnemonicUsedInTestContractsRepo(t *testing.T) {
	mnemonic := "test test test test test test test test test test test junk"
	address := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	privateKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	generatedKey, err := newEthereumKeyFromMnemonic(mnemonic)
	require.NoError(t, err, "error generating ethereum key")

	require.Equal(t, address, generatedKey.Address)
	require.Equal(t, privateKey, generatedKey.PrivateKeyHex)
}
