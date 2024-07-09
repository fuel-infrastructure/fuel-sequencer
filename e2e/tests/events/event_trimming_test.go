package events_test

import (
	"fmt"
	"time"

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
			NumInjectedEventTxs: 1, // matches the number of events emitted by AuthorizeMulti, per sequencer block
			NewEthereumBlock:    false,
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
		from := s.EthKeys[0].AddressHex
		to := s.EthKeys[1].AddressHex
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(from, to, sendCoins)

		// Calculate size of transaction resulting from AuthorizeEvent.
		authorizeEvent := types.AuthorizeEvent{
			Sender: from,
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
		authorizeData := testsuite.PackBatchAuthorize([][]byte{msgSendBz, msgSendBz, msgSendBz, msgSendBz})
		_, err = s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
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

func (s *EventsTestSuite) TestMaxEthBlockUpdateDelay() {

	s.Run("Test validators can resume the chain if block is too large and max syncup delay exceeded", func() {

		// -------- Setup

		// Calculate size of transaction resulting from MsgIndex.
		typicalMsgIndex := &bridgetypes.MsgIndex{
			Authority:           s.GetGovernanceAddress(),
			NumInjectedEventTxs: 1, // matches the number of events emitted by AuthorizeMulti, per sequencer block
			NewEthereumBlock:    false,
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
		from := s.EthKeys[0].AddressHex
		to := s.EthKeys[1].AddressHex
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(from, to, sendCoins)

		// Calculate size of transaction resulting from AuthorizeEvent.
		authorizeEvent := types.AuthorizeEvent{
			Sender: from,
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

		// Calculate max block size - this is not just for txs and must consider
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
		authorizeData := testsuite.PackBatchAuthorize([][]byte{msgSendBz, msgSendBz, msgSendBz, msgSendBz})
		_, err = s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// -------- Delay sync up

		s.PauseEthereum()
		// Wait MaxEthBlockUpdateDelay is no longer valid (padded some seconds due to caching on the sidecars)
		time.Sleep(time.Second * 40)
		s.UnpauseEthereum()

		// -------- Check that events are eventually processed

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
