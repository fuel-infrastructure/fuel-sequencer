package testsuite

import (
	"context"
	"math/big"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/ethereum/go-ethereum/common"

	commitmentstypes "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

// AuntsToHashes takes aunts from a Merkle proof and converts them to 32-byte hashes.
func AuntsToHashes(proof merkle.Proof) (hashes []common.Hash) {
	hashes = make([]common.Hash, len(proof.Aunts))
	for i, aunt := range proof.Aunts {
		hashes[i] = common.BytesToHash(aunt)
	}
	return
}

func (s *E2ETestSuite) QueryBridgeCommitment(ctx context.Context, start, end uint64) commitmentstypes.HexBytes {
	queryClient := s.getGRPCClients().CommitmentsQueryClient
	res, err := queryClient.BridgeCommitment(ctx,
		&commitmentstypes.QueryBridgeCommitmentRequest{
			Start: start,
			End:   end,
		},
	)
	s.Require().NoError(err)

	return res.BridgeCommitment
}

func (s *E2ETestSuite) QueryBridgeCommitmentInclusionProof(
	ctx context.Context, height, txIndex int64, start, end uint64,
) *commitmentstypes.QueryBridgeCommitmentInclusionProofResponse {
	queryClient := s.getGRPCClients().CommitmentsQueryClient
	res, err := queryClient.BridgeCommitmentInclusionProof(ctx,
		&commitmentstypes.QueryBridgeCommitmentInclusionProofRequest{
			Height:  height,
			TxIndex: txIndex,
			Start:   start,
			End:     end,
		},
	)
	s.Require().NoError(err)

	return res
}

func (s *E2ETestSuite) GetDataForUpdateCommitHeaderRange(
	ctx context.Context, start, end uint64,
) (targetHeaderHash, bridgeCommitmentHash common.Hash) {

	block, err := s.GetBlockByHeight(ctx, int64(end))
	s.Require().NoError(err)
	targetHeaderHash = common.BytesToHash(block.LastCommit.BlockID.Hash)

	bridgeCommitment := s.QueryBridgeCommitment(ctx, start, end)
	bridgeCommitmentHash = common.BytesToHash(bridgeCommitment)

	return
}

func (s *E2ETestSuite) GetDataForBridgeCommitmentInclusionProof(
	ctx context.Context, height, txIndex int64, start, end uint64,
) (
	bcLeaf BridgeCommitmentLeafForEthereum,
	bcLeafProof BinaryMerkleProofForEthereum,
	txResultMarshalled []byte,
	txResultProof BinaryMerkleProofForEthereum,
) {

	inclusionProof := s.QueryBridgeCommitmentInclusionProof(ctx, height, txIndex, start, end)

	// Construct BridgeCommitment leaf proof from the inclusion proof data.

	bridgeCommitmentMerkleProof := inclusionProof.BridgeCommitmentProof
	bcLeafProof = BinaryMerkleProofForEthereum{
		SideNodes: AuntsToHashes(*bridgeCommitmentMerkleProof.ToMerkleProof()),
		Key:       big.NewInt(bridgeCommitmentMerkleProof.Index),
		NumLeaves: big.NewInt(bridgeCommitmentMerkleProof.Total),
	}

	// Construct tx result proof from the inclusion proof data.

	lastResultsMerkleProof := inclusionProof.LastResultsProof
	txResultProof = BinaryMerkleProofForEthereum{
		SideNodes: AuntsToHashes(*lastResultsMerkleProof.ToMerkleProof()),
		Key:       big.NewInt(lastResultsMerkleProof.Index),
		NumLeaves: big.NewInt(lastResultsMerkleProof.Total),
	}

	// Construct BridgeCommitmentLeaf from the inclusion proof data.

	bcLeaf = BridgeCommitmentLeafForEthereum{
		Height:      big.NewInt(int64(inclusionProof.BridgeCommitmentLeaf.Height)),
		ResultsHash: common.BytesToHash(inclusionProof.BridgeCommitmentLeaf.LastResultsHash),
	}

	// Check that the proof is able to verify the marshalled tx result.

	lastResultsHash := inclusionProof.BridgeCommitmentLeaf.LastResultsHash
	txResultMarshalled = inclusionProof.TxResultMarshalled
	err := lastResultsMerkleProof.ToMerkleProof().Verify(lastResultsHash, txResultMarshalled)
	s.Require().NoError(err)

	return
}
