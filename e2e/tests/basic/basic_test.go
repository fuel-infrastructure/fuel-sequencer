package basic_test

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"go.uber.org/zap"
)

func (s *BasicTestSuite) TestSequencerAndSidecarBasics() {
	s.Run("Test basic Sequencer queries and transactions", func() {

		// Try getting module params (GRPC).
		params := s.QueryBankParams(s.Ctx())
		s.Require().True(params.DefaultSendEnabled)

		// Try getting address balances (GRPC).
		balance, err := s.QueryAllBalances(s.Ctx(), s.SeqKeys[0].AddressSeq, nil)
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
		from := s.SeqKeys[0].Address
		to := s.SeqKeys[1].Address
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
		updatedBalance, err := s.QueryAllBalances(s.Ctx(), s.SeqKeys[0].AddressSeq, nil)
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
		depositAmount := big.NewInt(200)
		depositTxReceipt := s.DepositTokenToSequencer(depositAmount)

		// Generate a MsgSend
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		from := s.EthKeys[0].AddressHex
		to := s.EthKeys[1].AddressHex
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(from, to, sendCoins)

		// Try generating some events via a transaction (RPC) - via authorize.
		authorizeData := testsuite.PackAuthorize(msgSendBz)
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
			Depositor: s.EthKeys[0].AddressHex,
			Recipient: s.EthKeys[0].AddressHex, // sender == recipient unless otherwise specified
			Amount:    depositAmount.String(),
			Lockup:    "0", // the deposit initiated from Ethereum has no lockup
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
			Sender: s.EthKeys[0].AddressHex,
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

	s.Run("Ensure we can subscribe to fully synced Ethereum blocks", func() {

		// Get last Ethereum block synced
		lastEthereumBlockSynced := s.QueryLastEthereumBlockSynced(s.Ctx())
		lookOutFor := lastEthereumBlockSynced + 2

		// Subscribe for 2 Ethereum blocks from now
		blockNumberKey := "fuelsequencer.bridge.EventEthereumBlockSynced.block_number"
		fullSyncKey := "fuelsequencer.bridge.EventEthereumBlockSynced.full_sync"
		query := fmt.Sprintf(`%s='"%d"' AND %s='true'`, blockNumberKey, lookOutFor, fullSyncKey)
		s.Logger().Info("subscribing to new Ethereum block", zap.String("query", query))
		sub, err := s.Chain.SubscribeToSequencer(s.Ctx(), query)
		s.Require().NoError(err)

		// Check that we eventually detect the event
		s.Require().Eventually(func() bool {
			select {
			case res := <-sub:
				s.Require().Equal([]string{fmt.Sprintf(`"%d"`, lookOutFor)}, res.Events[blockNumberKey])
				s.Require().Equal([]string{"true"}, res.Events[fullSyncKey])
				return true // found
			default:
				return false // not found yet
			}
		}, time.Minute/2, time.Second)
	})

	s.Run("Ensure user transactions are still subject to a finite gas meter", func() {

		// Here we want to confirm that even though we're setting an infinite gas meter for MsgIndex, which is the first
		// transaction in all blocks, the gas meter gets reset for any new transaction.

		from := s.SeqKeys[0].Address
		to := s.SeqKeys[1].Address
		amount := sdk.NewCoins(sdk.NewInt64Coin(testsuite.BridgeDenom, 100))
		msg := banktypes.NewMsgSend(from, to, amount)

		// Submit message with gas=1
		resp, err := s.SubmitMsgsWithGas(1, msg)
		s.Require().NoError(err)
		s.Require().Contains(resp.RawLog, "out of gas")
	})

	s.Run("Ensure that we can run an expedited proposal to change the inflation rate", func() {

		params := s.QueryMintParams(s.Ctx())
		inflationRate := s.QueryMintInflation(s.Ctx())
		newInflationRate := sdkmath.LegacyMustNewDecFromStr("0.5")

		// InflationMin and InflationMax are currently unequal
		s.Require().False(params.InflationMin.Equal(params.InflationMax))

		// Inflation and the new inflation rates are also unequal
		s.Require().False(inflationRate.Equal(newInflationRate))

		// Propose a new inflation rate via an expedited proposal
		params.InflationMin = newInflationRate
		params.InflationMax = newInflationRate
		msg := &minttypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *params,
		}
		s.ExecuteExpeditedGovProposal(msg)

		// Wait for 1 block to pass for the BeginBlocker to run
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10))

		// Check that the inflation rate was updated to the new inflation rate
		inflationRate = s.QueryMintInflation(s.Ctx())
		s.Require().True(inflationRate.Equal(newInflationRate))
	})
}
