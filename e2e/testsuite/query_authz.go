package testsuite

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/authz"
)

func (s *E2ETestSuite) QueryGranterGrants(ctx context.Context, granter string) []*authz.GrantAuthorization {
	queryClient := s.getGRPCClients().AuthZQueryClient
	res, err := queryClient.GranterGrants(ctx, &authz.QueryGranterGrantsRequest{
		Granter: granter,
	})
	s.Require().NoError(err)

	return res.Grants
}

func (s *E2ETestSuite) QueryGranteeGrants(ctx context.Context, grantee string) []*authz.GrantAuthorization {
	queryClient := s.getGRPCClients().AuthZQueryClient
	res, err := queryClient.GranteeGrants(ctx, &authz.QueryGranteeGrantsRequest{
		Grantee: grantee,
	})
	s.Require().NoError(err)

	return res.Grants
}
