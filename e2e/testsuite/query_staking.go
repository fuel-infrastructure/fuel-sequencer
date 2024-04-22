package testsuite

import (
	"context"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *E2ETestSuite) QueryDelegation(
	ctx context.Context, delegatorAddr, validatorAddr string,
) *stakingtypes.DelegationResponse {
	queryClient := s.getGRPCClients().StakingQueryClient
	res, err := queryClient.Delegation(ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	})
	s.Require().NoError(err)

	return res.DelegationResponse
}

func (s *E2ETestSuite) QueryDelegationRaw(
	ctx context.Context, delegatorAddr, validatorAddr string,
) (*stakingtypes.QueryDelegationResponse, error) {
	return s.getGRPCClients().StakingQueryClient.Delegation(ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	})
}
