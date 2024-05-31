package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

type BridgeCommitmentLeafForEthereum struct {
	Height      *big.Int
	ResultsHash common.Hash
}

type BinaryMerkleProofForEthereum struct {
	SideNodes []common.Hash
	Key       *big.Int
	NumLeaves *big.Int
}

func PackProcessSequencerWithdrawalMessage(
	proofNonce *big.Int,
	bridgeCommitmentLeaf BridgeCommitmentLeafForEthereum,
	bridgeCommitmentLeafProof BinaryMerkleProofForEthereum,
	txResultMarshalled []byte,
	txResultProof BinaryMerkleProofForEthereum,
) []byte {
	return packCall(
		sidecartypes.MockSequencerProxyContractABI,
		sidecartypes.MockProcessSequencerWithdrawalMessageFunctionName,
		[]interface{}{
			proofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		},
	)
}
