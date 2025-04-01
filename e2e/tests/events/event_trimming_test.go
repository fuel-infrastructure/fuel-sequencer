package events_test

import (
	"fmt"
	"math/big"
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

		// All transactions will have a non-zero sequence
		nonZeroSequence := uint64(1)

		// Calculate size of transaction resulting from MsgIndex.
		typicalMsgIndex := &bridgetypes.MsgIndex{
			Authority:           s.GetGovernanceAddress(),
			NumInjectedEventTxs: 1,
			NewEthereumBlock:    false,
			BlockNumber:         1,
		}
		typicalMsgIndexBz, err := typicalMsgIndex.RawTxBytes(nonZeroSequence)
		s.Require().NoError(err)
		typicalMsgIndexSize := utils.TxSize(typicalMsgIndexBz)

		s.Logger().Info(fmt.Sprintf("Predicted size of MsgIndex: %d", typicalMsgIndexSize))

		// Generate a Transfer
		sendAmount := int64(10)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(sendAmount))
		from := s.EthKeys[0]
		validator := s.SeqKeys[0]

		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(from.AddressHex, validator.ValAddressHex, sendCoin)

		// Generate a deposit & delegate
		depositEvent := types.DepositEvent{
			Depositor: from.AddressHex,
			Recipient: from.AddressHex,
			Amount:    sendCoin.Amount.String(),
			Lockup:    "0",
		}
		delegateEvent := types.AuthorizeEvent{
			Sender: from.AddressHex,
			Data:   msgDelegateBz,
		}

		// Calculate size of transaction resulting from AuthorizeEvent.
		depositEventMsg, err := depositEvent.Messages(testsuite.TestCdc, s.GetGovernanceAddress())
		s.Require().NoError(err)
		depositEventMsgBz, err := utils.ValidRawTxBytesFromAnyMsgs(depositEventMsg, nonZeroSequence)
		s.Require().NoError(err)
		depositEventMsgSize := utils.TxSize(depositEventMsgBz)
		s.Logger().Info(fmt.Sprintf("Predicted size of tx from DepositEvent: %d", depositEventMsgSize))

		delegateEventMsg, err := delegateEvent.Messages(testsuite.TestCdc, s.GetGovernanceAddress())
		s.Require().NoError(err)
		delegateEventMsgBz, err := utils.ValidRawTxBytesFromAnyMsgs(delegateEventMsg, nonZeroSequence)
		s.Require().NoError(err)
		delegateEventMsgSize := utils.TxSize(delegateEventMsgBz)
		s.Logger().Info(fmt.Sprintf("Predicted size of tx from DelegateEvent: %d", delegateEventMsgSize))

		// Set a low max bytes for txs so that events are split across multiple blocks, with a buffer of 10 bytes.
		largestEventMsgSize := depositEventMsgSize
		if delegateEventMsgSize > depositEventMsgSize {
			largestEventMsgSize = delegateEventMsgSize
		}
		maxBytesForTransactions := int64(typicalMsgIndexSize + largestEventMsgSize + 10)

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

		// Send deposit and delegate events,
		_ = s.DepositAndDelegateTokenToSequencer(sendCoin.Amount.BigInt(), validator.ValAddressEth)
		s.PollForEthereumEventIndexOffset(s.Ctx(), 20, 1) // check for deposit event processed
		s.PollForEthereumEventIndexOffset(s.Ctx(), 2, 0)  // check for delegate event processed in the next block
	})
}

func (s *EventsTestSuite) TestMaxEthBlockUpdateDelay() {

	s.Run("Test validators can resume the chain if block is too large and max syncup delay exceeded", func() {

		// -------- Setup

		// All transactions will have a non-zero sequence
		nonZeroSequence := uint64(1)

		// Calculate size of transaction resulting from MsgIndex.
		typicalMsgIndex := &bridgetypes.MsgIndex{
			Authority:           s.GetGovernanceAddress(),
			NumInjectedEventTxs: 1, // matches the number of events emitted by AuthorizeMulti, per sequencer block
			NewEthereumBlock:    false,
			BlockNumber:         1,
		}
		typicalMsgIndexBz, err := typicalMsgIndex.RawTxBytes(nonZeroSequence)
		s.Require().NoError(err)
		typicalMsgIndexSize := utils.TxSize(typicalMsgIndexBz)

		s.Logger().Info(fmt.Sprintf("Predicted size of MsgIndex: %d", typicalMsgIndexSize))

		// Generate a Transfer
		sendAmount := int64(10)
		to := s.EthKeys[1].Address
		transfer := testsuite.PackTransfer(to, big.NewInt(sendAmount))

		// -------- Send transactions

		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(transfer)
		s.Require().NoError(err)

		// -------- Delay sync up

		s.PauseEthereum()
		// Wait MaxEthBlockUpdateDelay is no longer valid (padded some seconds due to caching on the sidecars)
		time.Sleep(time.Second * 40)
		s.UnpauseEthereum()

		// -------- Check that events are eventually processed

		s.PollForLastEthereumBlockSynced(s.Ctx(), 20, txReceipt.BlockNumber.Uint64())
	})
}
