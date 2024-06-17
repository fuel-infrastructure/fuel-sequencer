package testsuite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEthereumKeyFromMnemonic(t *testing.T) {
	mnemonic := "test test test test test test test test test test test junk"
	address := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	privateKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	generatedKey, err := newEthereumKeyFromMnemonic(mnemonic)
	require.NoError(t, err, "error generating ethereum key")

	require.Equal(t, address, generatedKey.AddressHex)
	require.Equal(t, privateKey, generatedKey.PrivateKeyHex)
}

func TestNewEthereumKeyFromPrivateKey(t *testing.T) {
	address := "0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266"
	privateKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

	generatedKey, err := newEthereumKeyFromPrivateKey(privateKey)
	require.NoError(t, err, "error generating ethereum key")

	require.Equal(t, address, generatedKey.AddressHex)
	require.Equal(t, privateKey, generatedKey.PrivateKeyHex)
}
