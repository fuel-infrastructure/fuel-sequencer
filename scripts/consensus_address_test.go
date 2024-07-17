package scripts

import (
	"encoding/base64"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestConsensusAddressesFromEd25519Key(t *testing.T) {

	rawEd25519PubKeyBase64 := "tI24yf5EG1JSMBBYRhGiBjFG8xMeim8cgWA2j9kTA8M="

	ed25519PubKeyBz, err := base64.StdEncoding.DecodeString(rawEd25519PubKeyBase64)
	require.NoError(t, err)

	ed25519PubKey := ed25519.PubKey{
		Key: ed25519PubKeyBz,
	}

	address := ed25519PubKey.Address()
	consensusAddress := sdk.ConsAddress(address)

	require.Equal(t, "CBAFEB96C72AD720974170407BC55A3AC22FD36D", address.String())
	require.Equal(t, "fuelsequencervalcons1ewh7h9k89ttjp96pwpq8h3268tpzl5md0lfpta", consensusAddress.String())
}
