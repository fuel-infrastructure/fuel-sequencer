package testsuite

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// ExecuteGovProposal submits a governance proposal using the provided message and uses
// all validators to vote yes on the proposal. It ensures the proposal successfully passes.
//
// Forked from https://github.com/cosmos/ibc-go/blob/3a04e955f24332da39a86d3968fba7b47710b9e8
func (s *E2ETestSuite) ExecuteGovProposal(msg sdk.Msg) {
	sender, err := sdk.AccAddressFromBech32(ADDRESSES[0])
	s.Require().NoError(err)

	msgs := []sdk.Msg{msg}
	msgSubmitProposal, err := govtypesv1.NewMsgSubmitProposal(
		msgs,
		sdk.NewCoins(sdk.NewCoin(BridgeDenom, govtypesv1.DefaultMinDepositTokens)),
		sender.String(),
		"",
		"title",
		"summary",
		false,
	)
	s.Require().NoError(err)

	// Submit proposal
	resp, err := s.SubmitMsgs(msgSubmitProposal)
	s.Require().NoError(err)
	s.AssertValidTxResponse(*resp)

	// Prevent account sequence mismatches by waiting 1 block
	err = s.WaitForBlocks(s.Ctx(), 1, time.Minute)
	s.Require().NoError(err)

	// Get and increment proposal ID counter
	proposalId := uint64(s.govProposalIdCounter)
	s.govProposalIdCounter += 1

	// Vote yes from all validators
	for _, val := range s.Chain.validators {
		msgVote := govtypesv1.NewMsgVote(val.address(), proposalId, govtypesv1.VoteOption_VOTE_OPTION_YES, "")
		resp, err = s.SubmitMsgsFrom(val, msgVote)
		s.Require().NoError(err)
		s.AssertValidTxResponse(*resp)
	}

	// Wait for proposal to pass.
	pass := govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED
	s.PollForProposalStatus(s.Ctx(), blocksToWaitForGovProposalToPass, 1, pass)
}
