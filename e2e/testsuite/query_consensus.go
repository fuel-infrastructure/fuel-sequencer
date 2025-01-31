package testsuite

import (
	"context"

	consensustypes "cosmossdk.io/x/consensus/types"
	cmtproto "github.com/cometbft/cometbft/api/cometbft/types/v1"
)

func (s *E2ETestSuite) QueryConsensusParams(ctx context.Context) *cmtproto.ConsensusParams {
	queryClient := s.getGRPCClients().ConsensusQueryClient
	res, err := queryClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return res.Params
}
