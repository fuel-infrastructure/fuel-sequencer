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
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *WithdrawalsTestSuite) TestWithdrawalWithCentralisedSolution_WithdrawalFromSequencer() {
	s.Run("Submit a deposit to the Sequencer so that the Ethereum contract escrows the tokens", func() {

		sender := testsuite.ETH_KEYS[0]

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		amount := big.NewInt(200)
		mintData := testsuite.PackMint(common.HexToAddress(sender.AddressHex), amount)
		_, err := s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(amount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(amount))
		s.PollForBalance(s.Ctx(), 10, sender.AddressSeq, amountCoin)
	})

	s.Run("Submit a withdrawal on the Sequencer and make sure it can be actioned on Ethereum", func() {

		// --------------------------------------- User withdraws on the Sequencer

		aliceWallet := testsuite.SEQ_ADDRESSES[0]

		withdrawMsg := bridgemoduletypes.NewMsgWithdrawToEthereum(
			aliceWallet,
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			sdk.NewInt64Coin(testsuite.BridgeDenom, 100),
		)
		withdrawalResponse, err := s.SubmitMsgs(withdrawMsg)
		s.Require().NoError(err)
		s.Require().Zero(withdrawalResponse.Code)

		// The LastResultsHash is generated at the block right after the withdrawal
		lastResultsHashHeight := withdrawalResponse.Height + 1
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, 5*time.Second)) // ensure result generated

		startBlockString, targetBlockString, _, _, receipt := s.RunFuelStreamXProcessForBlock(lastResultsHashHeight + 1)
		startBlock, err := strconv.ParseUint(startBlockString, 10, 64)
		s.Require().NoError(err)
		targetBlock, err := strconv.ParseUint(targetBlockString, 10, 64)
		s.Require().NoError(err)

		// Make sure the LastResultsHash due to the transaction is included in the BridgeCommitment.
		// Since the target block is exclusive, it has to be > not >=.
		// If this check fails, then we need to reconfigure FuelStreamX.
		s.Require().Greater(targetBlock, uint64(lastResultsHashHeight))

		// The two events are: HeadUpdate, DataCommitmentStored
		s.Require().Len(receipt.Logs, 2)

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fuelStreamxABI, err := abi.JSON(strings.NewReader(sidecartypes.FuelStreamXContractABI))
		s.Require().NoError(err)

		// Check DataCommitmentStored event

		expectedBridgeCommitment, err := s.GetBridgeCommitment(s.Ctx(), startBlock, targetBlock)
		s.Require().NoError(err)

		actualStartBlock, err := strconv.ParseUint(eventTopic1[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)
		actualTargetBlock, err := strconv.ParseUint(eventTopic2[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)

		s.Require().Equal(testsuite.DataCommitmentStoredEventHash, eventTopic0)
		s.Require().EqualValues(startBlock, actualStartBlock)
		s.Require().EqualValues(targetBlock, actualTargetBlock)
		s.Require().Equal(expectedBridgeCommitment.String(), strings.ToUpper(eventTopic3[2:]))

		var event testsuite.DataCommitmentStoredEvent
		err = fuelStreamxABI.UnpackIntoInterface(&event, testsuite.DataCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the MsgIndex.

		txIndex := int64(1) // second tx
		bridgeCommitmentInclusionProof, err := s.GetBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, targetBlock,
		)
		s.Require().NoError(err)

		// Construct BridgeCommitment leaf proof from the inclusion proof data.

		bridgeCommitmentMerkleProof := bridgeCommitmentInclusionProof.BridgeCommitmentMerkleProof
		bridgeCommitmentLeafProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: testsuite.AuntsToHashes(*bridgeCommitmentMerkleProof.ToMerkleProof()),
			Key:       big.NewInt(bridgeCommitmentMerkleProof.Index),
			NumLeaves: big.NewInt(bridgeCommitmentMerkleProof.Total),
		}

		// Construct tx result proof from the inclusion proof data.

		lastResultsMerkleProof := bridgeCommitmentInclusionProof.LastResultsMerkleProof
		txResultProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: testsuite.AuntsToHashes(*lastResultsMerkleProof.ToMerkleProof()),
			Key:       big.NewInt(lastResultsMerkleProof.Index),
			NumLeaves: big.NewInt(lastResultsMerkleProof.Total),
		}

		// Construct BridgeCommitmentLeaf from the inclusion proof data.

		bridgeCommitmentLeaf := testsuite.BridgeCommitmentLeafForEthereum{
			Height:      big.NewInt(int64(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.Height)),
			ResultsHash: common.BytesToHash(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash),
		}

		// Check that the proof is able to verify the marshalled tx result.

		lastResultsHash := bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash
		txResultMarshalled := bridgeCommitmentInclusionProof.TxResultMarshalled
		err = lastResultsMerkleProof.ToMerkleProof().Verify(lastResultsHash, txResultMarshalled)
		s.Require().NoError(err)

		// Submit transaction to Ethereum to process the withdrawal.

		data := testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		)
		_, err = s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)
	})
}

func (s *WithdrawalsTestSuite) TestWithdrawalWithCentralisedSolution_WithdrawalFromEthereum() {
	s.Run("Submit a withdrawal from Ethereum and make sure it can be actioned on Ethereum", func() {

		withdrawerAddress := testsuite.ETH_KEYS[0].AddressHex
		withdrawCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 100)

		// --------------------------------------- Fund Ethereum owned account that will withdraw

		msgSend := &banktypes.MsgSend{
			FromAddress: testsuite.SEQ_ADDRESSES[0],
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
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, 5*time.Second)) // ensure result generated

		startBlockString, targetBlockString, _, _, receipt := s.RunFuelStreamXProcessForBlock(lastResultsHashHeight + 1)
		startBlock, err := strconv.ParseUint(startBlockString, 10, 64)
		s.Require().NoError(err)
		targetBlock, err := strconv.ParseUint(targetBlockString, 10, 64)
		s.Require().NoError(err)

		// Make sure the LastResultsHash due to the transaction is included in the BridgeCommitment.
		// Since the target block is exclusive, it has to be > not >=.
		// If this check fails, then we need to reconfigure FuelStreamX.
		s.Require().Greater(targetBlock, uint64(lastResultsHashHeight))

		// The two events are: HeadUpdate, DataCommitmentStored
		s.Require().Len(receipt.Logs, 2)

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fuelStreamxABI, err := abi.JSON(strings.NewReader(sidecartypes.FuelStreamXContractABI))
		s.Require().NoError(err)

		// Check DataCommitmentStored event

		expectedBridgeCommitment, err := s.GetBridgeCommitment(s.Ctx(), startBlock, targetBlock)
		s.Require().NoError(err)

		actualStartBlock, err := strconv.ParseUint(eventTopic1[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)
		actualTargetBlock, err := strconv.ParseUint(eventTopic2[2:], 16, 64) // hex to uint64
		s.Require().NoError(err)

		s.Require().Equal(testsuite.DataCommitmentStoredEventHash, eventTopic0)
		s.Require().EqualValues(startBlock, actualStartBlock)
		s.Require().EqualValues(targetBlock, actualTargetBlock)
		s.Require().Equal(expectedBridgeCommitment.String(), strings.ToUpper(eventTopic3[2:]))

		var event testsuite.DataCommitmentStoredEvent
		err = fuelStreamxABI.UnpackIntoInterface(&event, testsuite.DataCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the MsgIndex.

		txIndex := int64(1) // second tx
		bridgeCommitmentInclusionProof, err := s.GetBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, targetBlock,
		)
		s.Require().NoError(err)

		// Construct BridgeCommitment leaf proof from the inclusion proof data.

		bridgeCommitmentMerkleProof := bridgeCommitmentInclusionProof.BridgeCommitmentMerkleProof
		bridgeCommitmentLeafProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: testsuite.AuntsToHashes(*bridgeCommitmentMerkleProof.ToMerkleProof()),
			Key:       big.NewInt(bridgeCommitmentMerkleProof.Index),
			NumLeaves: big.NewInt(bridgeCommitmentMerkleProof.Total),
		}

		// Construct tx result proof from the inclusion proof data.

		lastResultsMerkleProof := bridgeCommitmentInclusionProof.LastResultsMerkleProof
		txResultProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: testsuite.AuntsToHashes(*lastResultsMerkleProof.ToMerkleProof()),
			Key:       big.NewInt(lastResultsMerkleProof.Index),
			NumLeaves: big.NewInt(lastResultsMerkleProof.Total),
		}

		// Construct BridgeCommitmentLeaf from the inclusion proof data.

		bridgeCommitmentLeaf := testsuite.BridgeCommitmentLeafForEthereum{
			Height:      big.NewInt(int64(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.Height)),
			ResultsHash: common.BytesToHash(bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash),
		}

		// Check that the proof is able to verify the marshalled tx result.

		lastResultsHash := bridgeCommitmentInclusionProof.BridgeCommitmentLeaf.LastResultsHash
		txResultMarshalled := bridgeCommitmentInclusionProof.TxResultMarshalled
		err = lastResultsMerkleProof.ToMerkleProof().Verify(lastResultsHash, txResultMarshalled)
		s.Require().NoError(err)

		// Submit transaction to Ethereum to process the withdrawal.

		data := testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		)
		_, err = s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)
	})
}
