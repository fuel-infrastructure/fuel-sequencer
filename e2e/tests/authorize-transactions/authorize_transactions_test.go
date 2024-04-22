package authorize_transactions_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTransactionsTestSuite) TestAuthorizedTransactions() {
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
		msgSendBz := s.E2ETestSuite.GenerateMsgSendBytes(senderAddress, receiverAddress, sendCoins)
		authorizeData := testsuite.PackAuthorize(msgSendBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		// Match the expected balances for each user depending on whether they are a sender or a receiver.
		s.PollForBalance(s.Ctx(), 10, senderAddress, expectedInitBalance.Sub(sendCoin))
		s.PollForBalance(s.Ctx(), 10, receiverAddress, expectedInitBalance.Add(sendCoin))
	})

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
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBytes(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := testsuite.PackAuthorize(msgDelegateBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		err = s.WaitForBlocks(s.Ctx(), 10, time.Minute)

		// Confirm that the delegation went through and is as expected.
		delegation := s.QueryDelegation(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Equal(delegateCoin, delegation.Balance)

		// ----------------------------------- Test MsgRedelegate

		// Make sure that there is no pre-existing delegation between the delegator and validator2.
		delegationRaw, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator2Address),
		)

		// Generate Authorize event wrapping a MsgBeginRedelegate to validator2.
		msgBeginRedelegateBz := s.E2ETestSuite.GenerateMsgBeginRedelegateBytes(
			delegatorAddress, validator1Address, validator2Address, delegateCoin,
		)
		authorizeData = testsuite.PackAuthorize(msgBeginRedelegateBz)
		err = s.SendEthTransactionToFuelStreamXContract(authorizeData)
		s.Require().NoError(err)

		err = s.WaitForBlocks(s.Ctx(), 10, time.Minute)

		// Confirm that the redelegation to validator2 went through and was executed as expected.
		delegation = s.QueryDelegation(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().Equal(delegateCoin, delegation.Balance)

		// Make sure that there are no delegations to validator1.
		delegationRaw, err = s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// ----------------------------------- Test MsgWithdrawDelegatorReward

		// TODO: Problem with rewards staying 0

		withdrawAddress := s.QueryDelegatorWithdrawAddress(s.Ctx(), delegatorAddress)
		fmt.Println(withdrawAddress)

		balance, err = s.QueryAllBalances(s.Ctx(), delegatorAddress, nil)
		fmt.Println(balance)

		// Make sure that rewards have accrued.
		rewards := s.QueryDelegationRewards(s.Ctx(), delegatorAddress, validator2Address)
		fmt.Println(rewards)

		// ----------------------------------- Test MsgUndelegate

		// TODO: Optimize with poll for delegation query instead of waiting for 10 blocks every time
	})
}
