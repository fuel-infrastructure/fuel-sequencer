package testsuite

import (
	"context"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

func (s *E2ETestSuite) QuerySlashingParams(ctx context.Context) *slashingtypes.Params {
	queryClient := s.getGRPCClients().SlashingQueryClient
	res, err := queryClient.Params(ctx, &slashingtypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}
