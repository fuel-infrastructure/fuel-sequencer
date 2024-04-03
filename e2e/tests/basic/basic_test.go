package basic_test

import (
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *BasicTestSuite) TestStartUpAndBasicQueries() {
	s.Run("Bring up nodes and perform some queries and transactions", func() {

		//// --------------------------------------- FuelSequencer queries
		//
		//// Try getting module params (GRPC).
		//params := s.QueryBankParams(s.Ctx())
		//s.Require().True(params.DefaultSendEnabled)
		//
		//// Try getting address balances (GRPC).
		//balance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		//s.Require().NoError(err)
		//expected := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin)
		//s.Require().True(expected.Amount.Equal(balance.Balances.AmountOf(testsuite.BridgeDenom)))
		//
		//// Try getting height (RPC).
		//sequencerHeight, err := s.GetFuelSequencerHeight(s.Ctx())
		//s.Require().NoError(err)
		//s.Require().Greater(sequencerHeight, uint64(1))
		//
		//// Try querying block by height (GRPC).
		//block, err := s.GetBlockByHeight(s.Ctx(), sequencerHeight)
		//s.Require().NoError(err)
		//s.Require().Equal(sequencerHeight, uint64(block.Header.Height))
		//
		//// --------------------------------------- FuelSequencer transactions
		//
		//// Try transferring tokens (RPC).
		//from := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[0])
		//to := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[1])
		//amount := sdk.NewCoins(sdk.NewInt64Coin(testsuite.BridgeDenom, 100))
		//msg := banktypes.NewMsgSend(from, to, amount)
		//res, err := s.SubmitMsgs(msg)
		//s.Require().NoError(err)
		//s.Require().Zero(res.Code)
		//
		//// Wait for blocks (RPC).
		//err = s.WaitForBlocks(s.Ctx(), 2, time.Minute)
		//s.Require().NoError(err)
		//
		//// Ensure balance was reduced (GRPC)
		//// Note: a fee was also charged.
		//updatedBalance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		//s.Require().NoError(err)
		//s.Require().True(updatedBalance.Balances.IsAllLT(balance.Balances))
		//
		//// --------------------------------------- Sidecar queries
		//
		//// Ensure we can query the sidecar (GRPC)
		//events, err := s.QuerySidecarBlockEvents(s.Ctx(), 1)
		//s.Require().NoError(err)
		//s.Require().Empty(events)
		//
		//// We expect an error if we query a block that doesn't exist.
		//events, err = s.QuerySidecarBlockEvents(s.Ctx(), 1000)
		//s.Require().Error(err)
		//
		//// --------------------------------------- Ethereum queries
		//
		//// Try getting height (RPC).
		//ethHeight1, err := s.GetEthereumHeight(s.Ctx())
		//s.Require().NoError(err)
		//s.Require().Greater(ethHeight1, uint64(1))
		//
		//// Try generating some events via a transaction (RPC) - via deposit.
		//toAddress := testsuite.ADDRESSES[1]
		//amount1 := big.NewInt(200)
		//amount2 := big.NewInt(300)
		//depositData := testsuite.PackDeposit(amount1, toAddress, amount2)
		//err = s.SendEthTransactionToProxyContract(depositData)
		//s.Require().NoError(err)
		//
		//// Try generating some events via a transaction (RPC) - via authorize.
		//someBytes := []byte("some bytes")
		//authorizeData := testsuite.PackAuthorize(someBytes)
		//err = s.SendEthTransactionToProxyContract(authorizeData)
		//s.Require().NoError(err)
		//
		//// --------------------------------------- Ensure Sidecar got the new Events
		//
		//// Wait for Sidecar to get the events.
		//s.Sleep(time.Second * 5)
		//
		//// Get latest Ethereum height and check two blocks higher.
		//ethHeight2, err := s.GetEthereumHeight(s.Ctx())
		//s.Require().NoError(err)
		//s.Require().Equal(ethHeight2, ethHeight1+2)
		//
		//// Ensure deposit event is at ethHeight1+1
		//depositEvents, err := s.QuerySidecarBlockEvents(s.Ctx(), int(ethHeight1+1))
		//s.Require().NoError(err)
		//s.Require().Len(depositEvents, 1)
		//s.Require().Equal(sidecartypes.SendToSequencerEventName, depositEvents[0].EventType)
		//
		//publicKey := s.GetEthPublicKey()
		//fromAddress := crypto.PubkeyToAddress(*publicKey).String()
		//
		//// Check deposit event data is as expected
		//var depositEventData sidecartypes.SendToSequencerEvent
		//err = depositEventData.Unmarshal(depositEvents[0].Data)
		//s.Require().NoError(err)
		//s.Require().True(depositEventData.Equal(&sidecartypes.SendToSequencerEvent{
		//	From:     fromAddress,
		//	Amount:   amount1.String(),
		//	To:       toAddress,
		//	Duration: amount2.String(),
		//}))
		//
		//// Ensure authorize event is at ethHeight1+2
		//authorizeEvents, err := s.QuerySidecarBlockEvents(s.Ctx(), int(ethHeight1+2))
		//s.Require().NoError(err)
		//s.Require().Len(authorizeEvents, 1)
		//s.Require().Equal(sidecartypes.AuthorizeEventName, authorizeEvents[0].EventType)
		//
		//// Check authorize event data is as expected
		//var authorizeEventData sidecartypes.AuthorizeEvent
		//err = authorizeEventData.Unmarshal(authorizeEvents[0].Data)
		//s.Require().NoError(err)
		//s.Require().True(authorizeEventData.Equal(&sidecartypes.AuthorizeEvent{
		//	From:    fromAddress,
		//	Message: someBytes,
		//}))
		//
		//// --------------------------------------- PreBlocker
		//
		//err = s.WaitForBlocks(s.Ctx(), 5, time.Minute)
		//s.Require().NoError(err)
		//lastEthereumBlockSynced := s.QueryLastEthereumBlockSynced(s.Ctx())
		//s.Require().EqualValues(ethHeight2, lastEthereumBlockSynced)

		// --------------------------------------- User withdrawals on the Sequencer
		aliceWallet := testsuite.ADDRESSES[0]

		withdrawMsg := bridgemoduletypes.NewMsgWithdrawToEthereum(
			aliceWallet,
			"0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266",
			sdk.NewInt64Coin(testsuite.BridgeDenom, 100),
		)
		res, err := s.SubmitMsgs(withdrawMsg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

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

		var event1 testsuite.DataCommitmentStoredEvent
		err = fuelstreamxABI.UnpackIntoInterface(&event1, testsuite.DataCommitmentStoredEventName, eventData)
		s.Require().NoError(err)

		expectedProofNonce := 1
		s.Require().EqualValues(expectedProofNonce, event1.ProofNonce.Uint64())

		// --------------------------------------- Make a withdrawal on Ethereum
		// TODO:
	})
}
