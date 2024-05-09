package basic_test

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func (s *BasicTestSuite) TestStartUpAndBasicQueries() {
	s.Run("Bring up nodes and perform some queries and transactions", func() {

		// --------------------------------------- FuelSequencer queries

		// Try getting module params (GRPC).
		params := s.QueryBankParams(s.Ctx())
		s.Require().True(params.DefaultSendEnabled)

		// Try getting address balances (GRPC).
		balance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		s.Require().NoError(err)
		expected := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin)
		s.Require().True(expected.Amount.Equal(balance.Balances.AmountOf(testsuite.BridgeDenom)))

		// Try getting height (RPC).
		sequencerHeight, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err)
		s.Require().Greater(sequencerHeight, uint64(1))

		// Try querying block by height (GRPC).
		block, err := s.GetBlockByHeight(s.Ctx(), int64(sequencerHeight))
		s.Require().NoError(err)
		s.Require().Equal(sequencerHeight, uint64(block.Header.Height))

		// --------------------------------------- FuelSequencer transactions

		// Try transferring tokens (RPC).
		from := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[0])
		to := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[1])
		amount := sdk.NewCoins(sdk.NewInt64Coin(testsuite.BridgeDenom, 100))
		msg := banktypes.NewMsgSend(from, to, amount)
		res, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Wait for blocks (RPC).
		err = s.WaitForSequencerBlocks(s.Ctx(), 2, time.Minute)
		s.Require().NoError(err)

		// Ensure balance was reduced (GRPC)
		// Note: a fee was also charged.
		updatedBalance, err := s.QueryAllBalances(s.Ctx(), testsuite.ADDRESSES[0], nil)
		s.Require().NoError(err)
		s.Require().True(updatedBalance.Balances.IsAllLT(balance.Balances))

		// --------------------------------------- Sidecar queries

		// Ensure we can query the sidecar (GRPC)
		events, err := s.QuerySidecarBlockEvents(s.Ctx(), 1)
		s.Require().NoError(err)
		s.Require().Empty(events)

		// We expect an error if we query a block that doesn't exist. Note, the block height is set to an arbitrarily
		// large number because it is impossible to reach that height with a one-second block time in this test.
		events, err = s.QuerySidecarBlockEvents(s.Ctx(), 1000000000)
		s.Require().Error(err)

		// --------------------------------------- Ethereum queries and transactions

		// Try getting height (RPC).
		ethHeight, err := s.GetEthereumHeight(s.Ctx())
		s.Require().NoError(err)
		s.Require().Greater(ethHeight, uint64(1))

		// Try generating some events via a transaction (RPC) - via deposit.
		depositAmount := big.NewInt(200)
		mintData := testsuite.PackMint(testsuite.SIGNER_ETH_ADDRESS, depositAmount)
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(depositAmount)
		depositTxReceipt, err := s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Try generating some events via a transaction (RPC) - via authorize.
		someBytes := []byte("some bytes")
		authorizeData := testsuite.PackAuthorize(someBytes)
		authorizeTxReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)

		// --------------------------------------- Ensure Sidecar got the new Events

		// Ensure deposit event is at the expected height.
		depositEvents, err := s.PollForSidecarBlockEvents(s.Ctx(), time.Second*20, int(depositTxReceipt.BlockNumber.Int64()))
		s.Require().NoError(err)
		s.Require().Len(depositEvents, 1)
		s.Require().Equal(sidecartypes.DepositEventName, depositEvents[0].EventType)

		// Check deposit event data is as expected
		var depositEventData sidecartypes.DepositEvent
		err = depositEventData.Unmarshal(depositEvents[0].Data)
		s.Require().NoError(err)
		s.Require().True(depositEventData.Equal(&sidecartypes.DepositEvent{
			Depositor: testsuite.SIGNER_ETH_ADDRESS_HEX,
			Recipient: testsuite.SIGNER_ETH_ADDRESS_HEX,
			Amount:    depositAmount.String(),
			Lockup:    "0",
		}))

		// Ensure authorize event is at the expected height.
		authorizeEvents, err := s.PollForSidecarBlockEvents(
			s.Ctx(), time.Second*20, int(authorizeTxReceipt.BlockNumber.Int64()),
		)
		s.Require().NoError(err)
		s.Require().Len(authorizeEvents, 1)
		s.Require().Equal(sidecartypes.AuthorizeEventName, authorizeEvents[0].EventType)

		// Check authorize event data is as expected
		var authorizeEventData sidecartypes.AuthorizeEvent
		err = authorizeEventData.Unmarshal(authorizeEvents[0].Data)
		s.Require().NoError(err)
		s.Require().True(authorizeEventData.Equal(&sidecartypes.AuthorizeEvent{
			Sender: testsuite.SIGNER_ETH_ADDRESS_HEX,
			Data:   someBytes,
		}))

		// --------------------------------------- Ensure PreBlocker is updating LastEthereumBlockSynced

		// Get last Ethereum block synced
		lastEthereumBlockSyncedOld := s.QueryLastEthereumBlockSynced(s.Ctx())

		// Wait for some blocks
		err = s.WaitForSequencerBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)

		// Get last Ethereum block synced again and make sure it updated
		lastEthereumBlockSynced := s.QueryLastEthereumBlockSynced(s.Ctx())
		s.Require().Greater(lastEthereumBlockSynced, lastEthereumBlockSyncedOld)
	})
}
