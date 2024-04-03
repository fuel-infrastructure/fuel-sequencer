package basic_test

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	cmtypes "github.com/cometbft/cometbft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *BasicTestSuite) TestWithdrawalWithMockedSuccinct() {
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

		// --------------------------------------- Run Operator

		heightAfterWithdrawal, err := s.Chain.FuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// We need to wait some blocks so that we're at a height that is greater than UPDATE_DELAY_BLOCKS.
		// Note: UPDATE_DELAY_BLOCKS has to be greater than the height at which we submitted the withdrawal.
		err = s.WaitForBlocks(s.Ctx(), 20, time.Minute)

		requestId, startBlockString, targetBlockString := s.RunSuccinctXOperatorMockApi()
		startBlock, err := strconv.ParseUint(startBlockString, 10, 64)
		s.Require().NoError(err)
		targetBlock, err := strconv.ParseUint(targetBlockString, 10, 64)
		s.Require().NoError(err)

		// Make sure the transaction is included in the BridgeCommitment
		s.Require().GreaterOrEqual(targetBlock, heightAfterWithdrawal)

		// --------------------------------------- Run Relayer

		// Get genesis header
		genesisBlockHeaderHash, err := s.Chain.GetBlockHeaderHash(s.Ctx(), 1)
		s.Require().NoError(err)

		err = s.WaitForBlocks(s.Ctx(), 5, time.Minute)
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

		fuelstreamxABI, err := abi.JSON(strings.NewReader(testsuite.FUEL_STREAM_X_ABI))
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

		lastResultsHashHeight := withdrawalResponse.Height + 1 // withdrawal's result is included in h+1's result hash
		txIndex := int64(1)                                    // block has 1 EthEventsTx followed by the withdrawal tx
		bridgeCommitmentInclusionProof, err := s.GetBridgeCommitmentInclusionProof(
			s.Ctx(), lastResultsHashHeight, txIndex, startBlock, targetBlock,
		)
		s.Require().NoError(err)

		block, err := s.GetBlockByHeight(s.Ctx(), lastResultsHashHeight)
		bridgeCommitmentLeaf := testsuite.BridgeCommitmentLeafForEthereum{
			Height:      big.NewInt(withdrawalResponse.Height),
			ResultsHash: common.BytesToHash(block.Header.LastResultsHash),
		}

		proof := bridgeCommitmentInclusionProof.BridgeCommitmentMerkleProof
		bridgeCommitmentLeafProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: []common.Hash{common.BytesToHash(proof.LeafHash)},
			Key:       big.NewInt(proof.Index),
			NumLeaves: big.NewInt(proof.Total),
		}

		blockResults, err := s.GetBlockResultsByHeight(s.Ctx(), withdrawalResponse.Height)
		abciResults := cmtypes.NewResults(blockResults.TxsResults)
		txResultMarshalled, err := abciResults[txIndex].Marshal()
		s.Require().NoError(err)

		proof = bridgeCommitmentInclusionProof.LastResultsMerkleProof
		txResultProof := testsuite.BinaryMerkleProofForEthereum{
			SideNodes: []common.Hash{common.BytesToHash(proof.LeafHash)},
			Key:       big.NewInt(proof.Index),
			NumLeaves: big.NewInt(proof.Total),
		}

		data := testsuite.PackProcessSequencerWithdrawalMessage(
			event.ProofNonce,
			bridgeCommitmentLeaf,
			bridgeCommitmentLeafProof,
			txResultMarshalled,
			txResultProof,
		)
		err = s.SendEthTransactionToFuelStreamXContract(data)
		s.Require().NoError(err)
	})
}
