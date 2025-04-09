package testsuite

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (s *E2ETestSuite) SubmitGovProposal(msg sdk.Msg) uint64 {
	return s.submitGovProposal(msg, false)
}

func (s *E2ETestSuite) SubmitExpeditedGovProposal(msg sdk.Msg) uint64 {
	return s.submitGovProposal(msg, true)
}

// SubmitGovProposal submits a governance proposal using the provided message. It also returns the proposal's ID.
//
// Inspired from https://github.com/cosmos/ibc-go/blob/3a04e955f24332da39a86d3968fba7b47710b9e8
func (s *E2ETestSuite) submitGovProposal(msg sdk.Msg, expedited bool) uint64 {
	msgs := []sdk.Msg{msg}
	msgSubmitProposal, err := govtypesv1.NewMsgSubmitProposal(
		msgs,
		sdk.NewCoins(sdk.NewCoin(BridgeDenom, govtypesv1.DefaultMinDepositTokens)),
		s.SeqKeys[0].AddressSeq,
		"",
		"title",
		"summary",
		expedited,
	)
	s.Require().NoError(err)

	// Submit proposal
	resp, err := s.SubmitMsgs(msgSubmitProposal)
	s.Require().NoError(err)
	s.AssertValidTxResponse(*resp)

	// Prevent account sequence mismatches by waiting 1 block
	err = s.WaitForSequencerBlocks(s.Ctx(), 1, time.Minute)
	s.Require().NoError(err)

	// Get and increment proposal ID counter
	proposalId := uint64(s.govProposalIdCounter)
	s.govProposalIdCounter += 1

	return proposalId
}

func (s *E2ETestSuite) ExecuteGovProposal(msg sdk.Msg) {
	s.executeGovProposal(msg, false)
}

func (s *E2ETestSuite) ExecuteExpeditedGovProposal(msg sdk.Msg) {
	s.executeGovProposal(msg, true)
}

// executeGovProposal submits a governance proposal using the provided message and uses
// all validators to vote yes on the proposal. It ensures the proposal successfully passes.
//
// Inspired from https://github.com/cosmos/ibc-go/blob/3a04e955f24332da39a86d3968fba7b47710b9e8
func (s *E2ETestSuite) executeGovProposal(msg sdk.Msg, expedited bool) {
	// Submit proposal and get its ID
	proposalId := s.submitGovProposal(msg, expedited)

	// Vote yes from all validators
	for _, val := range s.Chain.Validators {
		msgVote := govtypesv1.NewMsgVote(val.Address(), proposalId, govtypesv1.VoteOption_VOTE_OPTION_YES, "")
		resp, err := s.SubmitMsgsFrom(val, msgVote)
		s.Require().NoError(err)
		s.AssertValidTxResponse(*resp)
	}

	// Wait for proposal to pass.
	pass := govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED
	s.PollForProposalStatus(s.Ctx(), blocksToWaitForGovProposalToPass, proposalId, pass)
}
