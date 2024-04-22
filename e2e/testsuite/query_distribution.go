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

func (s *E2ETestSuite) QueryDelegationRewards(
	ctx context.Context, delegatorAddress, validatorAddress string,
) sdk.DecCoins {
	queryClient := s.getGRPCClients().DistributionQueryClient
	res, err := queryClient.DelegationRewards(ctx, &types.QueryDelegationRewardsRequest{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: validatorAddress,
	})
	s.Require().NoError(err)

	return res.Rewards
}

func (s *E2ETestSuite) QueryDelegatorWithdrawAddress(ctx context.Context, delegatorAddress string) string {
	queryClient := s.getGRPCClients().DistributionQueryClient
	res, err := queryClient.DelegatorWithdrawAddress(ctx, &types.QueryDelegatorWithdrawAddressRequest{
		DelegatorAddress: delegatorAddress,
	})
	s.Require().NoError(err)

	return res.WithdrawAddress
}
