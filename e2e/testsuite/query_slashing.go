package testsuite

import (
	"context"

	slashingtypes "cosmossdk.io/x/slashing/types"
)

func (s *E2ETestSuite) QuerySlashingParams(ctx context.Context) *slashingtypes.Params {
	queryClient := s.getGRPCClients().SlashingQueryClient
	res, err := queryClient.Params(ctx, &slashingtypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}
