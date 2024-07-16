package testsuite

import (
	"github.com/ethereum/go-ethereum/common"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func PackUpdateCommitHeaderRangeMessage(
	targetBlock uint64,
	targetHeader common.Hash,
	bridgeCommitment common.Hash,
) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockUpdateCommitHeaderRangeFunctionName,
		[]interface{}{
			targetBlock,
			targetHeader,
			bridgeCommitment,
		},
	)
}
