package authorize_transactions_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions() {
	s.Run("Submit an authorized MsgSend from Ethereum and check execution results on Sequencer", func() {
		// TODO: Fix this test, the problem is the signer on Ethereum has no balance on the sequencer. Maybe we need to
		//     : map the eth address to the fuelsequencer ones and give them a balance during startup. Or do deposit.
		// Make sure that the balance of the "From" user is as expected.
		expectedInitBalance := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin)
		balance, err := s.QueryAllBalances(s.Ctx(), testsuite.ETH_ADDRESSES[1], nil)
		s.Require().NoError(err)
		s.Require().True(expectedInitBalance.Amount.Equal(balance.Balances.AmountOf(testsuite.BridgeDenom)))

		// Make sure that the balance of the "To" user is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), testsuite.ETH_ADDRESSES[2], nil)
		s.Require().NoError(err)
		s.Require().True(expectedInitBalance.Amount.Equal(balance.Balances.AmountOf(testsuite.BridgeDenom)))

		// Generate Authorize event wrapping a MsgSend.
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBytes(
			testsuite.ETH_ADDRESSES[1], testsuite.ETH_ADDRESSES[2], sendCoins,
		)
		authorizeData := testsuite.PackAuthorizeMulti(msgSendBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Match the expected balances for each user depending on whether they are a sender or a receiver.
		s.PollForBalance(s.Ctx(), 10, testsuite.ADDRESSES[1], expectedInitBalance.Sub(sendCoin))
		s.PollForBalance(s.Ctx(), 10, testsuite.ADDRESSES[2], expectedInitBalance.Add(sendCoin))
	})
}
