package types_test

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
	"github.com/stretchr/testify/require"
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
	_, proofs := merkle.ProofsFromByteSlices(rs)

	// Ensure that converting the binary merkly proof back to a merkle proof yields the original proof.
	// If the proofs are equivalent, we can assume that the roots are also equivalent.
	for _, proof := range proofs {
		binaryMerkleProof := types.NewBinaryMerkleProof(*proof)
		proofFromBinaryMerkleProof := binaryMerkleProof.ToMerkleProof()

		require.EqualValues(t, proof, proofFromBinaryMerkleProof)
	}
}
