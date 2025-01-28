package authorize_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *AuthorizeTestSuite) TestAuthorizeEvents_MsgSend_ViaSidecarEncoding() {
	s.Run("Submit a transfer from Ethereum and check execution results on Sequencer", func() {
		senderAddress := s.EthKeys[0].AddressHex
		receiverAddress := s.EthKeys[1].AddressHex
		receiverAddressEth := s.EthKeys[1].Address

		// Make sure that the balance of the sender is as expected.
		expectedInitBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), senderAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), receiverAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		sendAmount, ok := sdkmath.NewIntFromString("10")
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)
		transferData := testsuite.PackTransfer(receiverAddressEth, sendAmount.BigInt())
		_, err = s.SendEthTransactionToSequencerInterfaceContract(transferData)
		s.Require().NoError(err)

		// Match the expected balances for each user depending on whether they are a sender or a receiver.
		s.PollForBalance(s.Ctx(), 10, senderAddress, expectedInitBalance.Sub(sendCoin))
		s.PollForBalance(s.Ctx(), 10, receiverAddress, expectedInitBalance.Add(sendCoin))
	})
}

func (s *AuthorizeTestSuite) TestAuthorizeEvents_StakingOperations_ViaSidecarEncoding() {
	s.Run("Submit staking operations from Ethereum and check execution results on Sequencer", func() {
		validator1Address := s.SeqKeys[0].ValAddressSeq
		validator1AddressEth := s.SeqKeys[0].ValAddressEth
		validator2Address := s.SeqKeys[1].ValAddressSeq
		validator2AddressEth := s.SeqKeys[1].ValAddressEth
		delegatorAddress := s.EthKeys[0].AddressHex
		withdrawAddressSeq := s.EthKeys[1].AddressSeq // alternate rewards withdrawal address (bech32)
		withdrawAddressEth := s.EthKeys[1].Address

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

		delegateAmount, ok := sdkmath.NewIntFromString("110000000000")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		delegateData := testsuite.PackDelegate(delegateAmount.BigInt(), validator1AddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(delegateData)
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

		redelegateData := testsuite.PackRedelegate(delegateAmount.BigInt(), validator1AddressEth, validator2AddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(redelegateData)
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

		// Wait some blocks for a significant amount of rewards to accrue, otherwise it's harder to perform some checks.
		err = s.WaitForSequencerBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)

		// Make sure that rewards have accrued.
		rewards := s.QueryDelegationRewards(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().NotZero(rewards.AmountOf(testsuite.BridgeDenom))

		claimRewardsData := testsuite.PackClaimRewards(validator2AddressEth)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(claimRewardsData)
		s.Require().NoError(err)

		// Wait for the withdrawal to be processed.
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// To make sure that the rewards withdrawal went through make sure that the rewards balance went down
		lessRewards := s.QueryDelegationRewards(s.Ctx(), delegatorAddress, validator2Address)
		s.Require().True(lessRewards.AmountOf(testsuite.BridgeDenom).LT(rewards.AmountOf(testsuite.BridgeDenom)))

		// ----------------------------------- Test MsgSetWithdrawAddress

		setRewardRecipientData := testsuite.PackSetRewardRecipient(withdrawAddressEth)
		txReceipt, err = s.SendEthTransactionToSequencerInterfaceContract(setRewardRecipientData)
		s.Require().NoError(err)

		// Wait for the change to be processed.
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Check that the withdrawals address was updated
		s.Require().Equal(withdrawAddressSeq, s.QueryDelegatorWithdrawAddress(s.Ctx(), delegatorAddress))

		// ----------------------------------- Test MsgUndelegate

		// Make sure that the delegation is still as originally set.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator2Address, delegateCoin)

		unbondData := testsuite.PackUnbond(delegateAmount.BigInt(), validator2AddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(unbondData)
		s.Require().NoError(err)

		// To make sure that the execution of MsgUndelegate went through check that all funds where withdrawn.
		s.PollForNoDelegation(s.Ctx(), 10, delegatorAddress, validator2Address)
	})
}

func (s *AuthorizeTestSuite) TestAuthorizeEvents_MsgWithdrawToEthereum_ViaSidecarEncoding() {
	s.Run("Submit a withdrawal to Ethereum from Ethereum and check execution results on Sequencer", func() {
		withdrawerAddress := s.EthKeys[0].AddressHex
		recipientAddressEth := s.EthKeys[1].Address

		// Make sure that the withdrawer's balance is as expected.
		expectedInitWithdrawerBalance := testsuite.InitBalanceCoin
		balance, err := s.QueryAllBalances(s.Ctx(), withdrawerAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitWithdrawerBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		withdrawAmount, ok := sdkmath.NewIntFromString("110000000000")
		s.Require().True(ok)
		withdrawCoin := sdk.NewCoin(testsuite.BridgeDenom, withdrawAmount)

		// Address withdraws to itself.
		withdrawData1 := testsuite.PackWithdraw(withdrawAmount.BigInt())
		_, err = s.SendEthTransactionToSequencerInterfaceContract(withdrawData1)
		s.Require().NoError(err)

		// Address withdraws to recipient.
		withdrawData2 := testsuite.PackWithdrawTo(withdrawAmount.BigInt(), recipientAddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(withdrawData2)
		s.Require().NoError(err)

		// Make sure that the withdrawals were executed by checking the withdrawers' balance
		postWithdrawalBalance := expectedInitWithdrawerBalance.Sub(withdrawCoin).Sub(withdrawCoin)
		s.PollForBalance(s.Ctx(), 10, withdrawerAddress, postWithdrawalBalance)
	})
}

func (s *AuthorizeTestSuite) TestAuthorizeEvents_MsgVote_ViaSidecarEncoding() {
	s.Run("Submit a vote from Ethereum and check execution results on Sequencer", func() {

		// Create a new dummy proposal to vote on
		consensusParams := s.QueryConsensusParams(s.Ctx())
		msgUpdateParams := consensustypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Block:     consensusParams.Block,
			Evidence:  consensusParams.Evidence,
			Validator: consensusParams.Validator,
			Abci:      consensusParams.Abci,
		}
		proposalId := s.SubmitGovProposal(&msgUpdateParams)

		// Make sure that the voting period started
		s.PollForProposalStatus(s.Ctx(), 10, proposalId, govtypesv1.ProposalStatus_PROPOSAL_STATUS_VOTING_PERIOD)

		// Confirm that the proposal has no votes yet
		votes := s.QueryVotes(s.Ctx(), proposalId)
		s.Require().Empty(votes)

		voteData := testsuite.PackVote(proposalId, uint32(govtypesv1.VoteOption_VOTE_OPTION_YES), "")
		_, err := s.SendEthTransactionToSequencerInterfaceContract(voteData)
		s.Require().NoError(err)

		// Make sure that the vote gets submitted by checking that the votes tally has increased from 0 to 1
		s.PollForNumberOfVotes(s.Ctx(), 10, proposalId, 1)
	})
}

func (s *AuthorizeTestSuite) TestAuthorizeEvents_AuthzOperations_ViaSidecarEncoding() {
	s.Run("Grant and Revoke an account authorisation from Ethereum and check Exec respectively succeeds and fails on Sequencer", func() {
		granterAddress := s.EthKeys[0].AddressHex

		granteeAddress := s.SeqKeys[1].AddressHex
		granteeAddressEth := s.SeqKeys[1].AddressEth

		// Make sure that the balances are as expected.
		expectedInitGranterBalance := testsuite.InitBalanceCoin
		granterBalance, err := s.QueryAllBalances(s.Ctx(), granterAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitGranterBalance.Amount, granterBalance.Balances.AmountOf(testsuite.BridgeDenom))

		expectedInitGranteeBalance := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin)
		granteeBalance, err := s.QueryAllBalances(s.Ctx(), granteeAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitGranteeBalance.Amount, granteeBalance.Balances.AmountOf(testsuite.BridgeDenom))

		// Prepare a send amount for testing execution
		sendAmount, ok := sdkmath.NewIntFromString("451")
		sendAmountUint := sendAmount.Uint64()
		s.Require().True(ok)
		sendCoin := sdk.NewCoin(testsuite.BridgeDenom, sendAmount)

		execMsg, err := mustPackGrantExecSendMsg(granteeAddress, granterAddress, sendCoin)
		s.Require().NoError(err)

		// Attempt to execute a MsgExec on Sequencer before authorisation is granted
		resp, err := s.SubmitMsgsFromValidatorN(1, execMsg)
		s.Require().NoError(err)
		// While tx submission succeeds, response should show failure as no authorization exists yet
		s.Require().Equal(authz.ErrNoAuthorizationFound.ABCICode(), resp.Code)

		// Grant an authorisation from Ethereum
		grantData := testsuite.PackGrant(granteeAddressEth, "/cosmos.bank.v1beta1.MsgSend", 0)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(grantData)
		s.Require().NoError(err)

		// Wait for the grant to be processed
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Execute a MsgExec on Sequencer after authorisation is granted
		resp, err = s.SubmitMsgsFromValidatorN(1, execMsg)
		s.Require().NoError(err)
		// Response should show success as authorisation is granted
		s.Require().Zero(resp.Code)

		// Verify the send was successful by checking balances
		s.PollForBalance(s.Ctx(), sendAmountUint, granterAddress, expectedInitGranterBalance.Sub(sendCoin))
		// Account for the fees of the 2 txs that previously took place...
		feeCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 2*(testsuite.MinGasPricesFloat*testsuite.DefaultTxGas))
		s.PollForBalance(s.Ctx(), sendAmountUint, granteeAddress, expectedInitGranteeBalance.Sub(feeCoin).Add(sendCoin))

		// Revoke the authorisation from Ethereum
		revokeData := testsuite.PackRevoke(granteeAddressEth, "/cosmos.bank.v1beta1.MsgSend")
		txReceipt, err = s.SendEthTransactionToSequencerInterfaceContract(revokeData)
		s.Require().NoError(err)

		// Wait for the revocation to be processed
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Attempt to execute a MsgExec on Sequencer after authorisation is revoked
		resp, err = s.SubmitMsgsFromValidatorN(1, execMsg)
		s.Require().NoError(err)
		// While tx submission succeeds, response should show failure as no authorization exists yet
		s.Require().Equal(authz.ErrNoAuthorizationFound.ABCICode(), resp.Code)
	})
}

// Helper function to create an exec message containing a send message
func mustPackGrantExecSendMsg(grantee, granter string, amount sdk.Coin) (*authz.MsgExec, error) {
	msg := &banktypes.MsgSend{
		FromAddress: granter,
		ToAddress:   grantee,
		Amount:      sdk.NewCoins(amount),
	}
	anyMsg, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, err
	}
	return &authz.MsgExec{
		Grantee: grantee,
		Msgs:    []*codectypes.Any{anyMsg},
	}, nil
}
