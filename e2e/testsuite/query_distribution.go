package testsuite

import (
	"context"
	"fmt"

	"cosmossdk.io/x/distribution/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// PollForDelegationRewards polls until the rewards balance matches
func (s *E2ETestSuite) PollForDelegationRewards(
	ctx context.Context, deltaBlocks uint64, delegatorAddress, validatorAddress string, balance sdk.DecCoins,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf(
		"Polling for rewards balance %s of delegation from %s to %s",
		balance.String(),
		delegatorAddress,
		validatorAddress,
	))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		res, err := s.Chain.grpcClients.DistributionQueryClient.DelegationRewards(
			ctx, &types.QueryDelegationRewardsRequest{
				DelegatorAddress: delegatorAddress,
				ValidatorAddress: validatorAddress,
			},
		)
		if err != nil {
			return nil, err
		}
		if !res.Rewards.Equal(balance) {
			return nil, fmt.Errorf(
				"rewards balance (%s) does not match expected: (%s)", res.Rewards, balance,
			)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "rewards balance not found in expected number of blocks")
}
