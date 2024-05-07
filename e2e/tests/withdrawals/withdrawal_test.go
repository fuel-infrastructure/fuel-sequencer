package withdrawals_test

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *WithdrawalsTestSuite) TestWithdrawalWithMockedSuccinct() {
	s.Run("Submit a withdrawal on the Sequencer and make sure it can be actioned on Ethereum", func() {

		// --------------------------------------- User withdraws on the Sequencer

		aliceWallet := testsuite.ADDRESSES[0]

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

		// --------------------------------------- Run Operator

		// We need to wait some blocks so that we're at a height that is greater than the operator UPDATE_DELAY_BLOCKS.
		// Note: UPDATE_DELAY_BLOCKS has to be greater than the height at which we submitted the withdrawal.
		err = s.WaitUntilSequencerBlock(s.Ctx(), testsuite.UPDATE_DELAY_BLOCKS+1, time.Minute)

		requestId, startBlockString, targetBlockString := s.RunSuccinctXOperatorMockApi()
		startBlock, err := strconv.ParseUint(startBlockString, 10, 64)
		s.Require().NoError(err)
		targetBlock, err := strconv.ParseUint(targetBlockString, 10, 64)
		s.Require().NoError(err)

		// Make sure the LastResultsHash due to the transaction is included in the BridgeCommitment.
		// Since the target block is exclusive, it has to be > not >=.
		s.Require().Greater(targetBlock, uint64(lastResultsHashHeight))

		// --------------------------------------- Run Relayer

		// Get genesis header
		genesisBlockHeaderHash, err := s.Chain.GetBlockHeaderHash(s.Ctx(), 1)
		s.Require().NoError(err)

		err = s.WaitForSequencerBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)

		receipt := s.RunSuccinctXRelayerMockApi(requestId, startBlock, targetBlock, genesisBlockHeaderHash)
		s.Require().Len(receipt.Logs, 3) // The three events are: HeadUpdate, DataCommitmentStored, Call

		// Extract log 1's topics and data
		s.Require().Len(receipt.Logs[1].Topics, 4)
		eventTopic0 := receipt.Logs[1].Topics[0].Hex()
		eventTopic1 := receipt.Logs[1].Topics[1].Hex()
		eventTopic2 := receipt.Logs[1].Topics[2].Hex()
		eventTopic3 := receipt.Logs[1].Topics[3].Hex()
		eventData := receipt.Logs[1].Data

		fuelstreamxABI, err := abi.JSON(strings.NewReader(sidecartypes.MockSequencerProxyContractABI))
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
		err = fuelstreamxABI.UnpackIntoInterface(&event, testsuite.DataCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum

		// Get BridgeCommitment inclusion proof
		// - The 'last result hash' incorporating the withdrawal result is at h+1.
		// - The withdrawal is assumed to be the second transaction in the block, following the EthEventsTx.

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
