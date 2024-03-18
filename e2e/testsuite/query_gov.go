package testsuite

import (
	"context"

	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

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

// QueryGovVotingParams queries the on-chain governance voting params.
func (s *E2ETestSuite) QueryGovVotingParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamVoting,
	})
	s.Require().NoError(err)

	return res.Params
}

// QueryGovDepositParams queries the on-chain governance deposit params.
func (s *E2ETestSuite) QueryGovDepositParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamDeposit,
	})
	s.Require().NoError(err)

	return res.Params
}

// QueryGovTallyParams queries the on-chain governance tally params.
func (s *E2ETestSuite) QueryGovTallyParams(ctx context.Context) *govtypesv1.Params {
	queryClient := s.getGRPCClients().GovQueryClient
	res, err := queryClient.Params(ctx, &govtypesv1.QueryParamsRequest{
		ParamsType: govtypesv1.ParamTallying,
	})
	s.Require().NoError(err)

	return res.Params
}
