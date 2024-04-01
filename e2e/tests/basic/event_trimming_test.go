package basic_test

import (
	cmtypes "github.com/cometbft/cometbft/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *BasicTestSuite) TestEventTrimming() {
	s.Run("Bring up nodes and perform some queries and transactions", func() {

		// The max block bytes is not just for transactions, and must consider an
		// allocation for the header and other components that make up a block.
		numberOfValidators := len(testsuite.MNEMONICS)
		maxBytesForTransactions := int64(200)
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

		// TODO: submit events and check if trimmed
	})
}
