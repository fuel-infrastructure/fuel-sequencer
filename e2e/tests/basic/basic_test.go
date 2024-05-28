package basic_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

func (s *BasicTestSuite) TestSequencerAndSidecarBasics() {
	s.Run("Test basic Sequencer queries and transactions", func() {

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
	})

	s.Run("Test Sidecar queries", func() {

		// Ensure we can query the sidecar (GRPC)
		events, err := s.QuerySidecarBlockEvents(s.Ctx(), 1)
		s.Require().NoError(err)
		s.Require().Empty(events)

		// We expect an error if we query a block that doesn't exist. Note, the block height is set to an arbitrarily
		// large number because it is impossible to reach that height with a one-second block time in this test.
		events, err = s.QuerySidecarBlockEvents(s.Ctx(), 1000000000)
		s.Require().Error(err)
	})

	s.Run("Check that Ethereum events get picked up by the Sidecar", func() {

		// Try getting height (RPC).
		ethHeight, err := s.GetEthereumHeight(s.Ctx())
		s.Require().NoError(err)
		s.Require().Greater(ethHeight, uint64(1))

		// Try generating some events via a transaction (RPC) - via deposit.
		toAddress := testsuite.ETH_ADDRESSES[1]
		amount1 := big.NewInt(200)
		amount2 := big.NewInt(300)
		depositData := testsuite.PackDeposit(amount1, common.HexToAddress(toAddress), amount2)
		depositTxReceipt, err := s.SendEthTransactionToFuelStreamXContract(depositData)
		s.Require().NoError(err)

		// Generate a MsgSend
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(testsuite.ETH_ADDRESSES[0], testsuite.ETH_ADDRESSES[1], sendCoins)

		// Try generating some events via a transaction (RPC) - via authorize.
		authorizeData := testsuite.PackAuthorize(msgSendBz)
		authorizeTxReceipt, err := s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// --------------------------------------- Ensure Sidecar got the new Events

		// Ensure deposit event is at the expected height.
		depositEvents, err := s.PollForSidecarBlockEvents(s.Ctx(), time.Second*20, int(depositTxReceipt.BlockNumber.Int64()))
		s.Require().NoError(err)
		s.Require().Len(depositEvents, 1)
		s.Require().Equal(sidecartypes.DepositEventName, depositEvents[0].EventType)

		publicKey := s.GetEthPublicKey()
		fromAddress := crypto.PubkeyToAddress(*publicKey).String()

		// Check deposit event data is as expected
		var depositEventData sidecartypes.DepositEvent
		err = depositEventData.Unmarshal(depositEvents[0].Data)
		s.Require().NoError(err)
		s.Require().True(depositEventData.Equal(&sidecartypes.DepositEvent{
			Depositor: fromAddress,
			Recipient: toAddress,
			Amount:    amount1.String(),
			Lockup:    amount2.String(),
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
			Sender: fromAddress,
			Data:   msgSendBz,
		}))
	})

	s.Run("Ensure LastEthereumBlockSynced is being updated", func() {

		// Get last Ethereum block synced
		lastEthereumBlockSyncedOld := s.QueryLastEthereumBlockSynced(s.Ctx())

		// Wait for some blocks
		err := s.WaitForSequencerBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)

		// Get last Ethereum block synced again and make sure it updated
		lastEthereumBlockSynced := s.QueryLastEthereumBlockSynced(s.Ctx())
		s.Require().Greater(lastEthereumBlockSynced, lastEthereumBlockSyncedOld)
	})

	s.Run("Ensure user transactions are still subject to a finite gas meter", func() {

		// Here we want to confirm that even though we're setting an infinite gas meter for MsgIndex, which is the first
		// transaction in all blocks, the gas meter gets reset for any new transaction.

		from := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[0])
		to := sdk.MustAccAddressFromBech32(testsuite.ADDRESSES[1])
		amount := sdk.NewCoins(sdk.NewInt64Coin(testsuite.BridgeDenom, 100))
		msg := banktypes.NewMsgSend(from, to, amount)

		// Submit message with gas=1
		resp, err := s.SubmitMsgsWithGas(1, msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, "out of gas")
	})
}
