package events_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	cmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// TestEventTrimming sets a reduced max bytes for blocks to showcase event trimming.
func (s *EventsTestSuite) TestEventTrimming() {

	s.Run("Run with reduced max bytes to showcase event trimming", func() {

		// Calculate size of transaction resulting from MsgIndex.
		typicalMsgIndex := &bridgetypes.MsgIndex{
			Authority:           s.GetGovernanceAddress(),
			NumInjectedEventTxs: 4, // matches the number of events emitted by AuthorizeMulti
			NewEthereumBlock:    true,
			BlockNumber:         1,
		}
		typicalMsgIndexBz, err := typicalMsgIndex.RawTxBytes()
		s.Require().NoError(err)
		typicalMsgIndexSize := len(typicalMsgIndexBz)

		s.Logger().Info(fmt.Sprintf("Predicted size of MsgIndex: %d", typicalMsgIndexSize))

		// Generate a MsgSend
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(testsuite.ETH_ADDRESSES[0], testsuite.ETH_ADDRESSES[1], sendCoins)

		// Calculate size of transaction resulting from AuthorizeEvent.
		authorizeEvent := types.AuthorizeEvent{
			Sender: testsuite.ETH_ADDRESSES[0],
			Data:   msgSendBz,
		}
		authorizeEventMsg, err := authorizeEvent.Messages(testsuite.TestCdc, s.GetGovernanceAddress())
		s.Require().NoError(err)
		authorizeEventMsgBz, err := utils.ValidRawTxBytesFromAnyMsgs(authorizeEventMsg)
		s.Require().NoError(err)
		authorizeEventMsgSize := len(authorizeEventMsgBz)

		s.Logger().Info(fmt.Sprintf("Predicted size of tx from AuthorizeEvent: %d", authorizeEventMsgSize))

		// Set a low max bytes for txs so that events are split across multiple blocks, with a buffer of 10 bytes.
		maxBytesForTransactions := int64(typicalMsgIndexSize + authorizeEventMsgSize + 10)

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
		authorizeData := testsuite.PackAuthorizeMulti([][]byte{msgSendBz, msgSendBz, msgSendBz, msgSendBz})
		_, err = s.SendEthTransactionToMockEthereumContract(authorizeData)
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
