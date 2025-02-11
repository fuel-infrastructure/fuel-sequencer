package testutil

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/stretchr/testify/require"
)

func SequencerProxyContractABI(t *testing.T) abi.ABI {
	var contractAbi abi.ABI
	err := contractAbi.UnmarshalJSON([]byte(sidecartypes.SequencerProxyContractABI))
	require.NoError(t, err)
	return contractAbi
}
