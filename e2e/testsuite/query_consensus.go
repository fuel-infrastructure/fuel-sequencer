package testsuite

import (
	"context"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
)

func (s *E2ETestSuite) QueryConsensusParams(ctx context.Context) *cmtproto.ConsensusParams {
	queryClient := s.getGRPCClients().ConsensusQueryClient
	res, err := queryClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return res.Params
}
