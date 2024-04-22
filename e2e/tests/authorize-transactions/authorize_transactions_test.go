package authorize_transactions_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions() {
	s.Run("Submit an authorized MsgSend from Ethereum and check execution results on Sequencer", func() {
		// Make sure that the balance of the "From" user is as expected.
		expectedInitBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), testsuite.ETH_ADDRESSES[0], nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Make sure that the balance of the "To" user is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), testsuite.ETH_ADDRESSES[1], nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate Authorize event wrapping a MsgSend.
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBytes(
			testsuite.ETH_ADDRESSES[0], testsuite.ETH_ADDRESSES[1], sendCoins,
		)
		authorizeData := testsuite.PackAuthorize(msgSendBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Match the expected balances for each user depending on whether they are a sender or a receiver.
		s.PollForBalance(s.Ctx(), 10, testsuite.ETH_ADDRESSES[0], expectedInitBalance.Sub(sendCoin))
		s.PollForBalance(s.Ctx(), 10, testsuite.ETH_ADDRESSES[1], expectedInitBalance.Add(sendCoin))
	})
}
