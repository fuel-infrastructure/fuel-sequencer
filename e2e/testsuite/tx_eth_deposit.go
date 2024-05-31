package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func PackDeposit(amount *big.Int, to common.Address, duration *big.Int) []byte {
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
