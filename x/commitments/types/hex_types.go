package types

import (
	"github.com/cometbft/cometbft/crypto/merkle"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
)

type (
	HexBytes      = cmtbytes.HexBytes
	HexBytesSlice = []cmtbytes.HexBytes
)

// NewBinaryMerkleProof creates a BinaryMerkleProof from a merkle.Proof.
func NewBinaryMerkleProof(proof merkle.Proof) *BinaryMerkleProof {

	var newAunts []cmtbytes.HexBytes
	for _, aunt := range proof.Aunts {
		newAunts = append(newAunts, aunt)
	}

	return &BinaryMerkleProof{
		Total:    proof.Total,
		Index:    proof.Index,
		LeafHash: proof.LeafHash,
		Aunts:    newAunts,
	}
}

// ToMerkleProof converts BinaryMerkleProof into the original merkle.Proof.
func (proof *BinaryMerkleProof) ToMerkleProof() *merkle.Proof {

	var newAunts [][]byte
	for _, aunt := range proof.Aunts {
		newAunts = append(newAunts, aunt)
	}

	return &merkle.Proof{
		Total:    proof.Total,
		Index:    proof.Index,
		LeafHash: proof.LeafHash,
		Aunts:    newAunts,
	}
}
