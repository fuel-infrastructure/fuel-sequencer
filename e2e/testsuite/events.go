package testsuite

import (
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	DataCommitmentStoredEventName = "DataCommitmentStored"
)

var (
	DataCommitmentStoredEventHash = crypto.Keccak256Hash([]byte("DataCommitmentStored(uint256,uint64,uint64,bytes32)")).Hex()
)

type DataCommitmentStoredEvent struct {
	// Note: the other values are indexed, so they show up as topics not as data fields.
	ProofNonce *big.Int `json:"proofNonce"`
}
