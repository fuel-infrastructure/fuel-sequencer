package testsuite

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (s *E2ETestSuite) QueryCommunityPool(ctx context.Context) sdk.DecCoins {
	queryClient := s.getGRPCClients().DistributionQueryClient
	res, err := queryClient.CommunityPool(ctx, &types.QueryCommunityPoolRequest{})
	s.Require().NoError(err)

	return res.Pool
}
