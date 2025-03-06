package types_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

func TestBinaryMerkleProofIsLossless(t *testing.T) {

	trs := []*abci.ExecTxResult{
		{Code: 0, Data: nil},
		{Code: 0, Data: []byte{}},
		{Code: 0, Data: []byte("one")},
		{Code: 14, Data: nil},
		{Code: 14, Data: []byte("foo")},
		{Code: 14, Data: []byte("bar")},
	}
	rs, err := abci.MarshalTxResults(trs)
	require.NoError(t, err)

	// Compute a set of proofs based on the above transactions.
	root, proofs := merkle.ProofsFromByteSlices(rs)

	// Ensure that converting the binary merkly proof back to a merkle
	// proof yields the original proof and that the merkle roots match.
	for _, proof := range proofs {
		binaryMerkleProof := types.NewBinaryMerkleProof(*proof)
		proofFromBinaryMerkleProof := binaryMerkleProof.ToMerkleProof()

		require.EqualValues(t, proof, proofFromBinaryMerkleProof)
		require.EqualValues(t, root, proofFromBinaryMerkleProof.ComputeRootHash())
	}
}
