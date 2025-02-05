package authorize_test

import (
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func expectNoDelegation(s *AuthorizeTestSuite, delegatorAddress, validatorAddress string) {
	delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validatorAddress)
	s.Require().Nil(delegationRaw)
	s.Require().ErrorContains(
		err,
		fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validatorAddress),
	)
}

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
		expectNoDelegation(s, delegatorAddress, validator1Address)

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
		expectNoDelegation(s, delegatorAddress, validator2Address)

		redelegateData := testsuite.PackRedelegate(delegateAmount.BigInt(), validator1AddressEth, validator2AddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(redelegateData)
		s.Require().NoError(err)

		// Confirm that the redelegation to validator2 went through and was executed as expected.
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator2Address, delegateCoin)

		// Make sure that there are no delegations to validator1.
		expectNoDelegation(s, delegatorAddress, validator1Address)

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

func (s *AuthorizeTestSuite) TestAuthorizeEvents_AuthzClaimRewardsOperations_ViaSidecarEncoding() {
	// Helper function to create an exec message containing a send message
	mustExecClaimRewardsAsGranteeMsg := func(grantee, granter, validatorAddress string) (*authz.MsgExec, error) {
		msg := &distributiontypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: granter,
			ValidatorAddress: validatorAddress,
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

	s.Run("Grant and Revoke from Ethereum to Claim Rewards from the Sequencer on behalf of a granter", func() {
		/**
		*	Test runs Operations (either through Eth rollup or Sequencer; with Granter, Grantee, Validator):
		*
		*	```
		*	Delegate (Eth or Seq; Granter -> Validator)
		*	ExecGrantClaim (Seq; Grantee -> Granter -> Validator)
		*	GrantClaim (Eth; Granter -> Grantee)
		*	RevokeClaim (Eth; Granter -> Grantee)
		*	```
		*
		*	`s.EthKeys` are independent of `s.SeqKeys`, which means all actors need to opt for either one.
		*
		*	Delegate can be done from Eth or Sequencer, with the Granter, implying `granter = s.EthKeys[0]` or `s.SeqKeys[0]`.
		*	ExecGrantClaim requires to be executed from the Sequencer as the Grantee, implying `grantee = s.SeqKeys[0]`.
		*	GrantClaim and RevokeClaim need to be done from Eth, with the Granter, implying `granter = s.EthKeys[0]`.
		*
		*	Validator is set to `s.SeqKeys[1]`, as it is the only one that is registered on the Sequencer.
		* 	Using [1] to keep it distinct from the [0] associated with `granter` and `grantee`, for a clear distinction.
		**/

		granter, grantee := s.EthKeys[0], s.SeqKeys[0]
		validator := s.SeqKeys[1]

		delegateAmount, ok := sdkmath.NewIntFromString("110000000000")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)

		// Make sure that the balances are as expected, with no existing delegations
		expectedInitGranterBalance := testsuite.InitBalanceCoin
		granterBalance, err := s.QueryAllBalances(s.Ctx(), granter.AddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitGranterBalance.Amount, granterBalance.Balances.AmountOf(testsuite.BridgeDenom))
		expectNoDelegation(s, granter.AddressSeq, validator.ValAddressSeq)

		s.Require().True(delegateAmount.LT(expectedInitGranterBalance.Amount))

		expectedInitGranteeBalance := testsuite.InitBalanceCoin.Sub(testsuite.InitStakedCoin) // using seq keys with staked coin
		granteeBalance, err := s.QueryAllBalances(s.Ctx(), grantee.AddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitGranteeBalance.Amount, granteeBalance.Balances.AmountOf(testsuite.BridgeDenom))
		expectNoDelegation(s, grantee.AddressSeq, validator.ValAddressSeq)

		// Setup and confirm a delegation between the grantee and validator
		delegateData := testsuite.PackDelegate(delegateAmount.BigInt(), validator.ValAddressEth)
		_, err = s.SendEthTransactionToSequencerInterfaceContract(delegateData)
		s.Require().NoError(err)

		s.PollForDelegationBalance(s.Ctx(), 10, granter.AddressSeq, validator.ValAddressSeq, delegateCoin)

		// Prepare a claim rewards exec message for testing execution
		execClaimMsg, err := mustExecClaimRewardsAsGranteeMsg(grantee.AddressSeq, granter.AddressSeq, validator.ValAddressSeq)
		s.Require().NoError(err)

		// Attempt to execute a claim rewards exec message on Sequencer before authorisation is granted, to show failure as no authorization exists yet
		resp, err := s.SubmitMsgs(execClaimMsg)
		s.Require().NoError(err)
		s.Require().Equal(authz.ErrNoAuthorizationFound.ABCICode(), resp.Code, resp.RawLog)

		// Grant an authorisation from Ethereum, and wait for the grant to be processed
		grantData := testsuite.PackGrantClaimRewards(grantee.AddressEth, 0)
		txReceipt, err := s.SendEthTransactionToSequencerInterfaceContract(grantData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Wait some blocks for rewards to accrue, and query the rewards before execution, to compare later on...
		err = s.WaitForSequencerBlocks(s.Ctx(), 5, time.Minute)
		s.Require().NoError(err)
		rewardsBeforeClaim := s.QueryDelegationRewards(s.Ctx(), granter.AddressSeq, validator.ValAddressSeq)
		s.Require().False(rewardsBeforeClaim.AmountOf(testsuite.BridgeDenom).IsZero())

		// Execute the claim rewards exec message on Sequencer after authorisation is granted, to show success as authorization exists now
		resp, err = s.SubmitMsgs(execClaimMsg)
		s.Require().NoError(err)
		s.Require().Zero(resp.Code, resp.RawLog)

		// Verify the claim rewards was successful by checking balances
		//// Make sure that the rewards balance went down for the respective delegation
		claimedAmount := s.ParseWithdrawDelegatorRewardFromTxResponse(resp)
		s.Require().False(claimedAmount.IsZero(), "claimed amount should not be zero")
		denomFound, denomClaimedCoin := claimedAmount.Find(testsuite.BridgeDenom)
		s.Require().True(denomFound, "rewards with test denom should be found")

		rewardsAfterClaim := s.QueryDelegationRewards(s.Ctx(), granter.AddressSeq, validator.ValAddressSeq)
		s.Require().True(rewardsAfterClaim.AmountOf(testsuite.BridgeDenom).LT(rewardsBeforeClaim.AmountOf(testsuite.BridgeDenom)))

		//// Granter paid for the delegation and its fee, and the fee for the grant authorization
		// granterFeesCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 2*(testsuite.MinGasPricesFloat*testsuite.DefaultTxGas))
		// TODO: Why shouldn't .Sub(granterFeesCoin) not included? As it is matches correctly with the actual balance...
		s.PollForBalance(s.Ctx(), 10, granter.AddressSeq, expectedInitGranterBalance.Sub(delegateCoin).Add(denomClaimedCoin))

		//// Grantee paid for the failed and successful MsgExec of the ClaimRewards
		granteeFeesCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, 2*(testsuite.MinGasPricesFloat*testsuite.DefaultTxGas))
		s.PollForBalance(s.Ctx(), 10, grantee.AddressSeq, expectedInitGranteeBalance.Sub(granteeFeesCoin))

		// Revoke the authorisation from Ethereum
		revokeData := testsuite.PackRevokeClaimRewards(grantee.AddressEth)
		txReceipt, err = s.SendEthTransactionToSequencerInterfaceContract(revokeData)
		s.Require().NoError(err)

		// Wait for the revocation to be processed
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, txReceipt.BlockNumber.Uint64())

		// Attempt to execute a claim rewards exec message on Sequencer after authorisation is revoked, to show failure with no authorization
		resp, err = s.SubmitMsgs(execClaimMsg)
		s.Require().NoError(err)
		// While tx submission succeeds, response should show failure as no authorization exists now
		s.Require().Equal(authz.ErrNoAuthorizationFound.ABCICode(), resp.Code, resp.RawLog)
	})
}
