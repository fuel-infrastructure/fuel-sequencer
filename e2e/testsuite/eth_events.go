package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	BridgeCommitmentStoredEventName = "BridgeCommitmentStored"
)

var (
	BridgeCommitmentStoredEventHash = crypto.Keccak256Hash([]byte("BridgeCommitmentStored(uint256,uint64,uint64,bytes32)")).Hex()
)

type BridgeCommitmentStoredEvent struct {
	// Note: the other values are indexed, so they show up as topics not as data fields.
	ProofNonce *big.Int `json:"proofNonce"`
}
