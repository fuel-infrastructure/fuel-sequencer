package events_test

import (
	cmtypes "github.com/cometbft/cometbft/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

// TestEventTrimming sets a reduced max bytes for blocks to showcase event trimming.
func (s *EventsTestSuite) TestEventTrimming() {

	s.T().Skip("We will probably have to disable this test anyways once we switch to repo with real contracts")

	s.Run("Run with reduced max bytes to showcase event trimming", func() {

		// Set a low max bytes for txs so that events are split across multiple blocks.
		maxBytesForTransactions := int64(150)

		// Calculate a max block size - this is not just for txs and must consider
		// the max size of the header and other components that make up a block.
		numberOfValidators := len(testsuite.MNEMONICS)
		maxBytes := maxBytesForTransactions +
			cmtypes.MaxOverheadForBlock +
			cmtypes.MaxHeaderBytes +
			cmtypes.MaxCommitBytes(numberOfValidators)
		s.Require().NotPanics(func() {
			_ = cmtypes.MaxDataBytesNoEvidence(maxBytes, numberOfValidators)
		})

		consensusParams := s.QueryConsensusParams(s.Ctx())
		consensusParams.Evidence.MaxBytes = 1
		consensusParams.Block.MaxBytes = maxBytes

		msgUpdateParams := consensustypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Block:     consensusParams.Block,
			Evidence:  consensusParams.Evidence,
			Validator: consensusParams.Validator,
			Abci:      consensusParams.Abci,
		}
		s.ExecuteGovProposal(&msgUpdateParams)

		// Check max block size was updated
		consensusParams = s.QueryConsensusParams(s.Ctx())
		s.Require().EqualValues(maxBytes, consensusParams.Block.MaxBytes)

		// Try generating some events via a transaction (RPC) - via authorize.
		someBytes := []byte("some bytes")
		authorizeData := testsuite.PackAuthorizeMulti(someBytes)
		_, err := s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// 1st event of 4 processed
		s.PollForEthereumEventIndexOffset(s.Ctx(), 20, 1)
		// 2nd event of 4 processed
		s.PollForEthereumEventIndexOffset(s.Ctx(), 2, 2)
		// 3rd event of 4 processed
		s.PollForEthereumEventIndexOffset(s.Ctx(), 2, 3)
		// 4th event of 4 processed
		s.PollForEthereumEventIndexOffset(s.Ctx(), 2, 0)
	})
}
