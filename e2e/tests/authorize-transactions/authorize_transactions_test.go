package authorize_transactions_test

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions_MsgSend() {
	s.Run("Submit an authorized MsgSend from Ethereum and check execution results on Sequencer", func() {
		senderAddress := testsuite.ETH_ADDRESSES[0]
		receiverAddress := testsuite.ETH_ADDRESSES[1]

		// Make sure that the balance of the sender is as expected.
		expectedInitBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), senderAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), receiverAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate Authorize event wrapping a MsgSend.
		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		sendCoins := sdk.NewCoins(sendCoin)
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBz(senderAddress, receiverAddress, sendCoins)
		authorizeData := testsuite.PackAuthorize(msgSendBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Match the expected balances for each user depending on whether they are a sender or a receiver.
		s.PollForBalance(s.Ctx(), 10, senderAddress, expectedInitBalance.Sub(sendCoin))
		s.PollForBalance(s.Ctx(), 10, receiverAddress, expectedInitBalance.Add(sendCoin))
	})
}

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions_StakingOperations() {
	s.Run("Submit authorized delegation messages from Ethereum and check execution results on Sequencer", func() {
		validator1Acc, err := sdk.AccAddressFromBech32(testsuite.ADDRESSES[0])
		s.Require().NoError(err)
		validator2Acc, err := sdk.AccAddressFromBech32(testsuite.ADDRESSES[1])
		s.Require().NoError(err)
		validator1Address := sdk.ValAddress(validator1Acc.Bytes()).String()
		validator2Address := sdk.ValAddress(validator2Acc.Bytes()).String()
		delegatorAddress := testsuite.ETH_ADDRESSES[0]

		// Make sure that the delegator's balance is as expected.
		expectedInitDelegatorBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), delegatorAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitDelegatorBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// ----------------------------------- Test MsgDelegate

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		delegateAmount, ok := sdkmath.NewIntFromString("110000000000")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := testsuite.PackAuthorize(msgDelegateBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Confirm that the delegation went through and is as expected.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator1Address, delegateCoin)

		// ----------------------------------- Test MsgRedelegate

		// Make sure that there is no pre-existing delegation between the delegator and validator2.
		delegationRaw, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator2Address),
		)

		// Generate Authorize event wrapping a MsgBeginRedelegate to validator2.
		msgBeginRedelegateBz := s.E2ETestSuite.GenerateMsgBeginRedelegateBz(
			delegatorAddress, validator1Address, validator2Address, delegateCoin,
		)
		authorizeData = testsuite.PackAuthorize(msgBeginRedelegateBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Confirm that the redelegation to validator2 went through and was executed as expected.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator2Address, delegateCoin)

		// Make sure that there are no delegations to validator1.
		delegationRaw, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// ----------------------------------- Test MsgWithdrawDelegatorReward

		// Make sure that rewards have accrued.
		rewards := s.QueryDelegationRewards(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().NotZero(rewards.AmountOf(testsuite.BridgeDenom))

		// Generate Authorize event wrapping a MsgWithdrawDelegatorReward.
		msgWithdrawDelegatorRewardBz := s.E2ETestSuite.GenerateMsgWithdrawDelegatorRewardBz(
			delegatorAddress, validator2Address,
		)
		authorizeData = testsuite.PackAuthorize(msgWithdrawDelegatorRewardBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// To make sure that the rewards withdrawal went through make sure that the rewards balance resets to zero/
		zeroRewards := sdk.NewDecCoins(sdk.NewDecCoin(testsuite.BridgeDenom, sdkmath.ZeroInt()))
		s.PollForDelegationRewards(s.Ctx(), 10, delegatorAddress, validator2Address, zeroRewards)

		// ----------------------------------- Test MsgUndelegate

		// Make sure that the delegation is still as originally set.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator2Address, delegateCoin)

		// Generate Authorize event wrapping a MsgUndelegate.
		msgUndelegateBz := s.E2ETestSuite.GenerateMsgUndelegateBz(delegatorAddress, validator2Address, delegateCoin)
		authorizeData = testsuite.PackAuthorize(msgUndelegateBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// To make sure that the execution of MsgUndelegate went through check that all funds where withdrawn.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator2Address, delegateCoin)
	})
}

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions_MsgWithdrawToEthereum() {
	s.Run("Submit authorized withdraw to Ethereum from Ethereum and check execution results on Sequencer", func() {
		withdrawerAddress := testsuite.ETH_ADDRESSES[0]

		// Make sure that the withdrawer's balance is as expected.
		expectedInitWithdrawerBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), withdrawerAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitWithdrawerBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate Authorize event wrapping a MsgWithdrawToEthereum.
		withdrawAmount, ok := sdkmath.NewIntFromString("110000000000")
		s.Require().True(ok)
		withdrawCoin := sdk.NewCoin(testsuite.BridgeDenom, withdrawAmount)
		msgWithdrawToEthereumBz := s.E2ETestSuite.GenerateMsgWithdrawToEthereumBz(
			withdrawerAddress, withdrawerAddress, withdrawCoin,
		)
		authorizeData := testsuite.PackAuthorize(msgWithdrawToEthereumBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Make sure that the withdrawal was executed by checking the withdrawers' balance
		postWithdrawalBalance := expectedInitWithdrawerBalance.Sub(withdrawCoin)
		s.PollForBalance(s.Ctx(), 10, withdrawerAddress, postWithdrawalBalance)
	})
}

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions_MsgVote() {
	s.Run("Submit authorized vote from Ethereum and check execution results on Sequencer", func() {
		// TODO: Submit a new proposal, hopefully manage to vote in due time by sending an authorize msgvote and check
		//     : that the vote went through by checking the list of votes. We might need to increase the voting period
		//     : in genesis if needed
		// TODO: Tomorrow use query_gov and tx_go to formulate tests and finish spec.

		//withdrawerAddress := testsuite.ETH_ADDRESSES[0]
		//
		//// Make sure that the withdrawer's balance is as expected.
		//expectedInitWithdrawerBalance := testsuite.InitBalanceCoin
		//balance, err := s.QueryAllBalances(s.Ctx(), withdrawerAddress, nil)
		//s.Require().NoError(err)
		//s.Require().Equal(expectedInitWithdrawerBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))
		//
		//// Generate Authorize event wrapping a MsgWithdrawToEthereum.
		//withdrawAmount, ok := sdkmath.NewIntFromString("110000000000")
		//s.Require().True(ok)
		//withdrawCoin := sdk.NewCoin(testsuite.BridgeDenom, withdrawAmount)
		//msgWithdrawToEthereumBz := s.E2ETestSuite.GenerateMsgWithdrawToEthereumBz(
		//	withdrawerAddress, withdrawerAddress, withdrawCoin,
		//)
		//authorizeData := testsuite.PackAuthorize(msgWithdrawToEthereumBz)
		//err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		//s.Require().NoError(err)
		//
		//// Make sure that the withdrawal was executed by checking the withdrawers' balance
		//postWithdrawalBalance := expectedInitWithdrawerBalance.Sub(withdrawCoin)
		//s.PollForBalance(s.Ctx(), 10, withdrawerAddress, postWithdrawalBalance)
	})
}
