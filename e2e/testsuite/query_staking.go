package testsuite

import (
	"context"
	"fmt"
	"strings"

	stakingtypes "cosmossdk.io/x/staking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *E2ETestSuite) QueryStakingParams(ctx context.Context) *stakingtypes.Params {
	queryClient := s.getGRPCClients().StakingQueryClient
	res, err := queryClient.Params(ctx, &stakingtypes.QueryParamsRequest{})
	s.Require().NoError(err)

	return &res.Params
}

func (s *E2ETestSuite) QueryDelegationRaw(
	ctx context.Context, delegatorAddr, validatorAddr string,
) (*stakingtypes.QueryDelegationResponse, error) {
	return s.getGRPCClients().StakingQueryClient.Delegation(ctx, &stakingtypes.QueryDelegationRequest{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
	})
}

func (s *E2ETestSuite) QueryDelegation(
	ctx context.Context, delegatorAddr, validatorAddr string,
) *stakingtypes.DelegationResponse {
	res, err := s.QueryDelegationRaw(ctx, delegatorAddr, validatorAddr)
	s.Require().NoError(err)

	return res.DelegationResponse
}

// PollForDelegationBalance polls until the delegation balance matches.
// Note: to poll for no delegation, use PollForNoDelegation instead.
func (s *E2ETestSuite) PollForDelegationBalance(
	ctx context.Context, deltaBlocks uint64, delegatorAddress, validatorAddress string, balance sdk.Coin,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf(
		"Polling for delegation balance %s of delegation from %s to %s",
		balance.String(),
		delegatorAddress,
		validatorAddress,
	))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		res, err := s.QueryDelegationRaw(ctx, delegatorAddress, validatorAddress)
		if err != nil {
			return nil, err
		}
		if !res.DelegationResponse.Balance.Equal(balance) {
			return nil, fmt.Errorf(
				"delegation balance (%s) does not match expected: (%s)", res.DelegationResponse.Balance, balance,
			)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "delegation balance not found in expected number of blocks")
}

// PollForNoDelegation polls until there is no delegation by the delegator to the validator.
func (s *E2ETestSuite) PollForNoDelegation(
	ctx context.Context, deltaBlocks uint64, delegatorAddress, validatorAddress string,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf(
		"Polling for no delegation from %s to %s",
		delegatorAddress,
		validatorAddress,
	))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		res, err := s.QueryDelegationRaw(ctx, delegatorAddress, validatorAddress)

		// We need to match the following error message to confirm that there are no delegations
		expectedErrMsg := fmt.Sprintf(
			"delegation with delegator %s not found for validator %s", delegatorAddress, validatorAddress,
		)

		// If the error is nil, it means that a delegation was found
		if err == nil {
			return nil, fmt.Errorf(
				"unexpected delegation balance found (%s)", res.DelegationResponse.Balance,
			)
		}

		// If a different error than what we expected is sent, we need to retry again as we are not sure if the
		// delegation is still there or not
		if !strings.Contains(err.Error(), expectedErrMsg) {
			return nil, err
		}

		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "delegation balance never missing in expected number of blocks")
}
