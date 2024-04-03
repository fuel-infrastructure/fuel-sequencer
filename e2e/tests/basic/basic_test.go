package basic_test

import (
	"strconv"
	"time"

	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
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
		currentHeight, err := s.Chain.FuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)

		// We need to wait some blocks so that we're at a height that is greater than UPDATE_DELAY_BLOCKS.
		// Note: UPDATE_DELAY_BLOCKS has to be greater than the height at which we submitted the withdrawal.
		err = s.WaitForBlocks(s.Ctx(), 5, time.Minute)

		requestId, startBlockString, targetBlockString := s.RunSuccinctXOperatorMockApi()
		startBlock, err := strconv.Atoi(startBlockString)
		s.Require().NoError(err)
		targetBlock, err := strconv.Atoi(targetBlockString)
		s.Require().NoError(err)

		// Make sure the transaction is included in the BridgeCommitment
		s.Require().GreaterOrEqual(uint64(targetBlock), currentHeight)

		// --------------------------------------- Run Relayer

		// Get genesis header
		genesisBlockHeaderHash, err := s.Chain.GetBlockHeaderHash(s.Ctx(), 1)
		s.Require().NoError(err)

		err = s.WaitForBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)

		s.RunSuccinctXRelayerMockApi(requestId, uint64(startBlock), uint64(targetBlock), genesisBlockHeaderHash)

		// --------------------------------------- Make a withdrawal on Ethereum
		// TODO:
	})
}
