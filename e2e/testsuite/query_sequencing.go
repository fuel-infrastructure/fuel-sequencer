package testsuite

import (
	"context"

	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (s *E2ETestSuite) QuerySequencingParams(ctx context.Context) *sequencingtypes.Params {
	queryClient := s.getGRPCClients().SequencingQueryClient
	res, err := queryClient.Params(ctx, &sequencingtypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}
