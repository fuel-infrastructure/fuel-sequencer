package testsuite

import (
	"context"

	commitmentstypes "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/types"
)

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
