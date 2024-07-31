package withdrawals_test

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *WithdrawalsTestSuite) TestWithdrawalWithCentralisedSolution_WithdrawalFromSequencer() {
	s.Run("Submit a deposit to the Sequencer so that the Ethereum contract escrows the tokens", func() {

		sender := s.EthKeys[0]

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is s.EthKeys[0]
		amount := big.NewInt(200)
		mintData := testsuite.PackMintToken(common.HexToAddress(sender.AddressHex), amount)
		_, err := s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(testsuite.SequencerInterfaceContractAddress, amount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(amount))
		s.PollForBalance(s.Ctx(), 10, sender.AddressSeq, amountCoin)
	})

	s.Run("Submit a withdrawal on the Sequencer and make sure it can be actioned on Ethereum", func() {

		// --------------------------------------- User withdraws on the Sequencer

		aliceWallet := s.SeqKeys[0].AddressSeq

		withdrawMsg := bridgemoduletypes.NewMsgWithdrawToEthereum(
			aliceWallet,
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			sdk.NewInt64Coin(testsuite.BridgeDenom, 100),
		)
		withdrawalResponse, err := s.SubmitMsgs(withdrawMsg)
		s.Require().NoError(err)
		s.Require().Zero(withdrawalResponse.Code)

		// The LastResultsHash is generated at the block right after the withdrawal.
		// We wait two Sequencer blocks to ensure that we can capture it in the Bridge Commitment.
		lastResultsHashHeight := withdrawalResponse.Height + 1
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, 10*time.Second))

		// Submit bridge commitment to FuelStreamX contract
		startBlock := uint64(1)
		endBlock := uint64(lastResultsHashHeight + 1)
		targetHeaderHash, bridgeCommitmentHash := s.GetDataForUpdateCommitHeaderRange(s.Ctx(), startBlock, endBlock)
		data := testsuite.PackUpdateCommitHeaderRangeMessage(endBlock, targetHeaderHash, bridgeCommitmentHash)
		receipt, err := s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)

		// The two events are: HeadUpdate, BridgeCommitmentStored
		s.Require().Len(receipt.Logs, 2)

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fuelStreamxABI, err := abi.JSON(strings.NewReader(testsuite.FuelStreamXContractABI))
		s.Require().NoError(err)

		// Check BridgeCommitmentStored event

		expectedBridgeCommitment := s.QueryBridgeCommitment(s.Ctx(), startBlock, endBlock)

		actualStartBlock, err := strconv.ParseUint(eventTopic1[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)
		actualTargetBlock, err := strconv.ParseUint(eventTopic2[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)

		s.Require().Equal(testsuite.BridgeCommitmentStoredEventHash, eventTopic0)
		s.Require().EqualValues(startBlock, actualStartBlock)
		s.Require().EqualValues(endBlock, actualTargetBlock)
		s.Require().Equal(expectedBridgeCommitment.String(), strings.ToUpper(eventTopic3[2:]))

		var event testsuite.BridgeCommitmentStoredEvent
		err = fuelStreamxABI.UnpackIntoInterface(&event, testsuite.BridgeCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the MsgIndex.
		txIndex := int64(1) // second tx
		bcLeaf, bcLeafProof, txResultMarshalled, txResultProof := s.GetDataForBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, endBlock,
		)

		// Submit transaction to Ethereum to process the withdrawal.
		data = testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce, bcLeaf, bcLeafProof, txResultMarshalled, txResultProof,
		)
		_, err = s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)
	})
}

func (s *WithdrawalsTestSuite) TestWithdrawalWithCentralisedSolution_WithdrawalFromEthereum() {
	s.Run("Submit a deposit to the Sequencer so that the Ethereum contract escrows the tokens", func() {

		sender := s.EthKeys[0]

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is s.EthKeys[0]
		amount := big.NewInt(200)
		mintData := testsuite.PackMintToken(common.HexToAddress(sender.AddressHex), amount)
		_, err := s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(testsuite.SequencerInterfaceContractAddress, amount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(amount))
		s.PollForBalance(s.Ctx(), 10, sender.AddressSeq, amountCoin)
	})

	s.Run("Submit a withdrawal from Ethereum and make sure it can be actioned on Ethereum", func() {

		withdrawerAddress := s.EthKeys[0].AddressHex
		withdrawCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 100)

		// --------------------------------------- Fund Ethereum owned account that will withdraw

		msgSend := &banktypes.MsgSend{
			FromAddress: s.SeqKeys[0].AddressSeq,
			ToAddress:   withdrawerAddress,
			Amount:      sdk.NewCoins(withdrawCoin),
		}
		res, err := s.SubmitMsgs(msgSend)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// --------------------------------------- User withdraws on the Sequencer via Ethereum

		sequencerHeightBefore, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Generate Authorize event wrapping a MsgWithdrawToEthereum.
		msgWithdrawToEthereumBz := s.E2ETestSuite.GenerateMsgWithdrawToEthereumBz(
			withdrawerAddress, withdrawerAddress, withdrawCoin,
		)
		authorizeData := testsuite.PackAuthorize(msgWithdrawToEthereumBz)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// The LastResultsHash is generated at the block right after the withdrawal
		s.PollForLastEthereumBlockSynced(s.Ctx(), 20, txReceipt.BlockNumber.Uint64())
		sequencerHeightAfter, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// Get withdrawal event type
		withdrawalEvent, err := sdk.TypedEventToEvent(&bridgemoduletypes.EventWithdrawToEthereumReported{})
		s.Require().NoError(err)
		withdrawalEventType := withdrawalEvent.Type

		// Find the actual LastResultsHash height
		var found bool
		var lastResultsHashHeight int64
		for i := sequencerHeightBefore; i <= sequencerHeightAfter; i++ {
			_, found = s.SearchForEventInBlockResults(s.Ctx(), withdrawalEventType, int64(i))
			if found {
				lastResultsHashHeight = int64(i + 1)
				break
			}
		}
		s.Require().True(found)

		// We wait two Sequencer blocks to ensure that we can capture the last results hash in the Bridge Commitment.
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, 10*time.Second))

		// Submit bridge commitment to FuelStreamX contract
		startBlock := uint64(1)
		endBlock := uint64(lastResultsHashHeight + 1)
		targetHeaderHash, bridgeCommitmentHash := s.GetDataForUpdateCommitHeaderRange(s.Ctx(), startBlock, endBlock)
		data := testsuite.PackUpdateCommitHeaderRangeMessage(endBlock, targetHeaderHash, bridgeCommitmentHash)
		receipt, err := s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)

		// The two events are: HeadUpdate, BridgeCommitmentStored
		s.Require().Len(receipt.Logs, 2)

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fuelStreamxABI, err := abi.JSON(strings.NewReader(testsuite.FuelStreamXContractABI))
		s.Require().NoError(err)

		// Check BridgeCommitmentStored event

		expectedBridgeCommitment := s.QueryBridgeCommitment(s.Ctx(), startBlock, endBlock)

		actualStartBlock, err := strconv.ParseUint(eventTopic1[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)
		actualTargetBlock, err := strconv.ParseUint(eventTopic2[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)

		s.Require().Equal(testsuite.BridgeCommitmentStoredEventHash, eventTopic0)
		s.Require().EqualValues(startBlock, actualStartBlock)
		s.Require().EqualValues(endBlock, actualTargetBlock)
		s.Require().Equal(expectedBridgeCommitment.String(), strings.ToUpper(eventTopic3[2:]))

		var event testsuite.BridgeCommitmentStoredEvent
		err = fuelStreamxABI.UnpackIntoInterface(&event, testsuite.BridgeCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the MsgIndex.
		txIndex := int64(1) // second tx
		bcLeaf, bcLeafProof, txResultMarshalled, txResultProof := s.GetDataForBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, endBlock,
		)

		// Submit transaction to Ethereum to process the withdrawal.
		data = testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce, bcLeaf, bcLeafProof, txResultMarshalled, txResultProof,
		)
		_, err = s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)
	})
}
