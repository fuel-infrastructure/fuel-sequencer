package testsuite

import (
	"math/big"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func PackDeposit(amount *big.Int, to string, duration *big.Int) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockDepositFunctionName,
		[]interface{}{
			amount,
			to,
			duration,
		},
	)
}
