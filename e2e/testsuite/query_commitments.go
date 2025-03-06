package testsuite

import (
	"context"
	"math/big"

	"github.com/cometbft/cometbft/crypto/merkle"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	commitmentstypes "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

// publicValues is a set of values required when calling updateCommitHeaderRange.
type publicValues struct {
	TrustedBlock      uint64
	TrustedHeaderHash common.Hash
	TargetBlock       uint64
	TargetHeaderHash  common.Hash
	BridgeCommitment  common.Hash
}

func (pv *publicValues) paddedBytes() ([]byte, error) {
	var bz []byte

	trustedBlockBz, err := utils.To32PaddedHexBytes(pv.TrustedBlock)
	if err != nil {
		return nil, err
	}

	targetBlockBz, err := utils.To32PaddedHexBytes(pv.TargetBlock)
	if err != nil {
		return nil, err
	}

	bz = append(bz, trustedBlockBz...)
	bz = append(bz, pv.TrustedHeaderHash.Bytes()...)
	bz = append(bz, targetBlockBz...)
	bz = append(bz, pv.TargetHeaderHash.Bytes()...)
	bz = append(bz, pv.BridgeCommitment.Bytes()...)

	return bz, nil
}

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

// GetDataForUpdateCommitHeaderRange generates the data needed for the updateCommitHeaderRange function to pass.
//
// It includes an option to get the trusted header hash from the contract, since we do not have permission to set the
// initial trusted header, since it's reserved to the contract admin. Thankfully it's not that important to test this.
func (s *E2ETestSuite) GetDataForUpdateCommitHeaderRange(
	ctx context.Context, start, end uint64, trustedHeaderHashFromContract bool,
) (proof, publicValuesBz []byte) {

	bridgeCommitment := s.QueryBridgeCommitment(ctx, start, end)
	bridgeCommitmentHash := common.BytesToHash(bridgeCommitment)

	var trustedHeaderHash common.Hash
	var err error
	if trustedHeaderHashFromContract {
		trustedHeaderHash, err = s.QueryBlockHeightToHeaderHashQueryName(s.Ctx(), start)
		s.Require().NoError(err)
	} else {
		startBlock, err := s.GetBlockByHeight(ctx, int64(end))
		s.Require().NoError(err)
		trustedHeaderHash = common.BytesToHash(startBlock.LastCommit.BlockID.Hash)
	}

	endBlock, err := s.GetBlockByHeight(ctx, int64(end))
	s.Require().NoError(err)
	targetHeaderHash := common.BytesToHash(endBlock.LastCommit.BlockID.Hash)

	pv := &publicValues{
		TrustedBlock:      start,
		TrustedHeaderHash: trustedHeaderHash,
		TargetBlock:       end,
		TargetHeaderHash:  targetHeaderHash,
		BridgeCommitment:  bridgeCommitmentHash,
	}
	bz, err := pv.paddedBytes()
	s.Require().NoError(err)

	// Note: the mock SP1 verifier expects an empty proof. This is why we return nil.
	return nil, bz
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
