package testsuite

import (
	"context"

	sdkmath "cosmossdk.io/math"
	minttypes "cosmossdk.io/x/mint/types"
)

func (s *E2ETestSuite) QueryMintParams(ctx context.Context) *minttypes.Params {
	queryClient := s.getGRPCClients().MintQueryClient
	res, err := queryClient.Params(ctx, &minttypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryMintInflation(ctx context.Context) sdkmath.LegacyDec {
	queryClient := s.getGRPCClients().MintQueryClient
	res, err := queryClient.Inflation(ctx, &minttypes.QueryInflationRequest{})
	s.Require().NoError(err)

	return res.Inflation
}
