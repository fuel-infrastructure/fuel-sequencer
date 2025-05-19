package testsuite

import (
	"context"

	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

func (s *E2ETestSuite) QueryBondParams(ctx context.Context) *bondtypes.Params {
	queryClient := s.getGRPCClients().BondQueryClient
	res, err := queryClient.Params(ctx, &bondtypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryBondState(ctx context.Context) bondtypes.State {
	queryClient := s.getGRPCClients().BondQueryClient
	res, err := queryClient.State(ctx, &bondtypes.QueryStateRequest{})
	s.Require().NoError(err)

	return res.State
}
