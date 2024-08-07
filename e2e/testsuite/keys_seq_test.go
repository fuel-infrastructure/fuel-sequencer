package testsuite

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSequencerKeyFromMnemonic(t *testing.T) {
	mnemonic := "test test test test test test test test test test test junk"
	address := "fuelsequencer15yk64u7zc9g9k2yr2wmzeva5qgwxps6y3z4xeu"

	generatedKey, err := newSequencerKeyFromMnemonic(mnemonic)
	require.NoError(t, err, "error generating sequencer key")

	require.Equal(t, address, generatedKey.AddressSeq)
}
