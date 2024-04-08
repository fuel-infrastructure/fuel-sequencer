package testsuite

import (
	"context"
	"fmt"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (s *E2ETestSuite) GetGovernanceAddress() string {
	return authtypes.NewModuleAddress(govtypes.ModuleName).String()
}

func (s *E2ETestSuite) QueryProposal(ctx context.Context, proposalID uint64) (govtypesv1.Proposal, error) {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Proposal(ctx, &govtypesv1.QueryProposalRequest{
		ProposalId: proposalID,
	})
	if err != nil {
		return govtypesv1.Proposal{}, err
	}

	return *res.Proposal, nil
}

// QueryGovVotingParams queries the on-Chain governance voting params.
func (s *E2ETestSuite) QueryGovVotingParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamVoting,
	})
	s.Require().NoError(err)

	return res.Params
}

// QueryGovDepositParams queries the on-Chain governance deposit params.
func (s *E2ETestSuite) QueryGovDepositParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamDeposit,
	})
	s.Require().NoError(err)

	return res.Params
}

// QueryGovTallyParams queries the on-Chain governance tally params.
func (s *E2ETestSuite) QueryGovTallyParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamTallying,
	})
	s.Require().NoError(err)

	return res.Params
}

// PollForProposalStatus polls until the proposal status matches
func (s *E2ETestSuite) PollForProposalStatus(
	ctx context.Context, deltaBlocks uint64, proposalID uint64, status govtypesv1.ProposalStatus,
) {
	h, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)

	s.T().Log(fmt.Sprintf("Polling for status %s of proposal %d", status, proposalID))

	doPoll := func(ctx context.Context, height uint64) (any, error) {
		proposal, err := s.QueryProposal(ctx, proposalID)
		if err != nil {
			return nil, err
		}

		if proposal.Status != status {
			return nil, fmt.Errorf("status (%s) does not match expected: (%s)", proposal.Status, status)
		}
		return nil, nil
	}

	bp := BlockPoller[any]{CurrentHeight: s.Chain.FuelSequencerHeight, PollFunc: doPoll}
	_, err = bp.DoPoll(ctx, h, h+deltaBlocks)
	s.Require().NoError(err, "status not found in expected number of blocks")
}
