package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
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
