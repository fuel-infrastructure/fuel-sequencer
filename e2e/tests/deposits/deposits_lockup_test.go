package deposits_test

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit, using the vesting seconds as the amount, so that 1 token becomes spendable per second. Note that
		// we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)

		// --------------------------------------- Delegate

		validator1Address := s.SeqKeys[0].ValAddressSeq
		validator1AddressEth := s.SeqKeys[0].ValAddressEth
		delegatorAddress := s.EthKeys[0].AddressHex

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		// Expect 60 seconds from vestingStartTime for 60 tokens to be available.
		timeForUnlock := time.Second * 60
		delegateAmount, ok := sdkmath.NewIntFromString("60")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		delegateData := testsuite.PackDelegate(delegateAmount.BigInt(), validator1AddressEth)

		// Expect a failure because not enough time has elapsed yet
		delegation1, err := s.SendEthTransactionToSequencerInterfaceContract(delegateData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation1.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validator1Address)

		// Sleep the remaining time so that enough tokens will be spendable
		s.Sleep(time.Until(vestingStartTime.Add(timeForUnlock)))

		// Confirm that the delegation went through and is as expected.
		delegation2, err := s.SendEthTransactionToSequencerInterfaceContract(delegateData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation2.BlockNumber.Uint64()) // wait until tx processed
		s.PollForDelegationBalance(s.Ctx(), 0, delegatorAddress, validator1Address, delegateCoin)

		// Check account again
		ethOwnedVestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)

		// --------------------------------------- Undelegate

		// Override unbonding time so that undelegation goes through immediately
		stakingParams := s.QueryStakingParams(s.Ctx())
		stakingParams.UnbondingTime = time.Second
		s.ExecuteGovProposal(&stakingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *stakingParams,
		})

		// Ensure value updated
		s.Require().Equal(time.Second, s.QueryStakingParams(s.Ctx()).UnbondingTime)

		// Delegate to validator1 with the amount previously delegated.
		undelegateCoin := delegateCoin
		unbondData := testsuite.PackUnbond(undelegateCoin.Amount.BigInt(), validator1AddressEth)

		// Confirm that the undelegation went through by checking for no delegation
		undelegation, err := s.SendEthTransactionToSequencerInterfaceContract(unbondData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, undelegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validator1Address)

		// Wait for undelegation to go through (Note: unbonding time is very small)
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10))

		// Check account again
		ethOwnedVestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup_WithDelegateInSameTx_FailsIfNotEnoughVested() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender
		validatorAddressHex := s.SeqKeys[0].ValAddressHex

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit and Delegate, using the vesting seconds as the amount, so that 1 token becomes spendable per second.
		// Note that we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		validatorAddress := common.HexToAddress(validatorAddressHex)
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		delegation := s.DepositTokenToSequencerFromMigrationAndDelegate(
			sendAmount, validatorAddress, testsuite.VestingDuration2Years,
		)

		// Expect a failure because not enough time has elapsed yet
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForBalance(s.Ctx(), 0, ownedReceiverAddressSeq, amountCoin)              // deposit successful
		s.PollForNoDelegation(s.Ctx(), 0, senderAddress, validatorAddressHex)          // delegation unsuccessful

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup_WithDelegateInSameTx_SuccessIfEnoughVested() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender
		validatorAddressHex := s.SeqKeys[0].ValAddressHex

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Set the vesting start time back by -2 years so that the delegation of the deposit amount passes
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		bridgeParams.VestingStartTime = bridgeParams.VestingStartTime.Add(-testsuite.VestingDuration2Years)
		s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		})

		// Confirm the new value
		bridgeParamsAfterProposal := s.QueryBridgeParams(s.Ctx())
		s.Require().Equal(bridgeParams.VestingStartTime, bridgeParamsAfterProposal.VestingStartTime)

		// Deposit and Delegate - the amount does not matter much here, but we'll use the same amount as other tests.
		// Note that we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		validatorAddress := common.HexToAddress(validatorAddressHex)
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		delegation := s.DepositTokenToSequencerFromMigrationAndDelegate(
			sendAmount, validatorAddress, testsuite.VestingDuration2Years,
		)

		// Expect delegation to go through
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForDelegationBalance(s.Ctx(), 0, senderAddress, validatorAddressHex, amountCoin)

		// Calculate expected values
		vestingStartTime := bridgeParamsAfterProposal.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParamsAfterProposal.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithNoVesting_WithLockup_WithDelegateInSameTx() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with no vesting and check results", func() {
		validatorAddressSeq := s.SeqKeys[0].AddressSeq     // Address of one of the validators
		validatorAddressHex := s.SeqKeys[0].ValAddressHex  // Address of one of the validators
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Create the account by transferring tokens to it
		from := sdk.MustAccAddressFromBech32(validatorAddressSeq)
		to := sdk.MustAccAddressFromBech32(ownedReceiverAddressSeq)
		initAmount := big.NewInt(1000)
		initBalanceCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(initAmount))
		initBalance := sdk.NewCoins(initBalanceCoin)
		msg := banktypes.NewMsgSend(from, to, initBalance)
		res, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Make sure that the balance of the receiver is as expected.
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(initBalance, balance.Balances)

		// Deposit and Delegate!
		//
		// The amount that we deposit and delegate here must be less than or equal to the amount already available.
		// Otherwise, the total spendable (i.e. stakeable) coins will not be enough to perform the delegation.
		// Note that we need to downscale using the v1 to v2 migration ratio.
		sendAmount := big.NewInt(1000)
		sendAmountScaledDown := new(big.Int).Quo(sendAmount, testsuite.MigrationRatio)
		validatorAddress := common.HexToAddress(validatorAddressHex)
		_ = s.DepositTokenToSequencerFromMigrationAndDelegate(
			sendAmountScaledDown, validatorAddress, testsuite.VestingDuration2Years,
		)

		// Match the expected delegation for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForDelegationBalance(s.Ctx(), 10, senderAddress, validatorAddressHex, amountCoin)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithNoVesting_WithLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with no vesting and check results", func() {
		validatorAddress := s.SeqKeys[0].AddressSeq        // Address of one of the validators
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Create the account by transferring tokens to it
		from := sdk.MustAccAddressFromBech32(validatorAddress)
		to := sdk.MustAccAddressFromBech32(ownedReceiverAddressSeq)
		initAmount := big.NewInt(100)
		initBalanceCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(initAmount))
		initBalance := sdk.NewCoins(initBalanceCoin)
		msg := banktypes.NewMsgSend(from, to, initBalance)
		res, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Make sure that the balance of the receiver is as expected.
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(initBalance, balance.Balances)

		// Deposit!
		//
		// Note that we need to downscale using the v1 to v2 migration ratio.
		sendAmount := big.NewInt(200)
		sendAmountScaledDown := new(big.Int).Quo(sendAmount, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmountScaledDown, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly deposited vesting tokens.
		expAmount := new(big.Int).Add(sendAmount, initAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// However, the OriginalVesting amount will be the deposited amount
		expOriginalVesting := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(expOriginalVesting).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithVesting_WithLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with vesting and check results", func() {
		validatorAddress := s.SeqKeys[0].AddressSeq        // Address of one of the validators
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Send tokens to the account and set it as a vesting account
		from := sdk.MustAccAddressFromBech32(validatorAddress)
		to := sdk.MustAccAddressFromBech32(ownedReceiverAddressSeq)
		initVestingAmount := big.NewInt(100)
		initVestingAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(initVestingAmount))
		initVestingAmountCoins := sdk.NewCoins(initVestingAmountCoin)
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		initVestingDuration := time.Second * time.Duration(94608000) // 3 years vesting
		initVestingEndTime := bridgeParams.VestingStartTime.Add(initVestingDuration).Unix()
		msg := vestingtypes.NewMsgCreateVestingAccount(from, to, initVestingAmountCoins, initVestingEndTime, false)
		res, err := s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Make sure that the balance of the receiver is as expected.
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(initVestingAmountCoins, balance.Balances)

		// Make sure that the account was created with the right vesting.
		continuousVestingAccount, err := s.QueryContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(initVestingEndTime, continuousVestingAccount.EndTime)
		s.Require().Equal(initVestingAmountCoins, continuousVestingAccount.OriginalVesting)

		// Deposit!
		//
		// Note that we need to downscale using the v1 to v2 migration ratio.
		sendAmount := big.NewInt(200)
		sendAmountScaledDown := new(big.Int).Quo(sendAmount, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmountScaledDown, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly deposited vesting tokens.
		expAmount := new(big.Int).Add(sendAmount, initVestingAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// However, the OriginalVesting amount will be the deposited amount
		expOriginalVesting := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))

		// Calculate expected values
		bridgeParams = s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(expOriginalVesting).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
	})
}

func (s *DepositsTestSuite) TestDeposits_WithChangingVestingStartTimeAndLockupPeriods() {
	s.Run("Submit multiple deposits with different vesting details and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// -----------------------------------------------------------------------------------------------
		// Deposit 1:
		// - VestingStartTime = Genesis time
		// - VestingDuration = 2 years

		// Deposit, using the vesting seconds as the amount, so that 1 token becomes spendable per second. Note that
		// we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin1 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime1 := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime1 := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime1.Unix(), vestingAcc.StartTime)
		s.Require().Equal(vestingEndTime1.Unix(), vestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin1).Equal(vestingAcc.OriginalVesting))
		s.Require().Nil(vestingAcc.DelegatedFree)
		s.Require().Nil(vestingAcc.DelegatedVesting)

		// -----------------------------------------------------------------------------------------------
		// Deposit 2:
		// - VestingStartTime = Genesis time
		// - VestingDuration = 2 years
		// Note: same as previous

		// Deposit, using the vesting seconds as the amount, so that now 2 tokens becomes spendable per second. Note
		// that we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt = big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount = new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer. It's double the amount since it's deposit 2.
		amountCoin2 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1.Add(amountCoin2))

		vestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime1.Unix(), vestingAcc.StartTime) // same vesting start time
		s.Require().Equal(vestingEndTime1.Unix(), vestingAcc.EndTime)     // same vesting end time
		s.Require().True(sdk.NewCoins(amountCoin1.Add(amountCoin2)).Equal(vestingAcc.OriginalVesting))
		s.Require().Nil(vestingAcc.DelegatedFree)
		s.Require().Nil(vestingAcc.DelegatedVesting)

		// -----------------------------------------------------------------------------------------------
		// Deposit 3:
		// - VestingStartTime = Genesis time
		// - VestingDuration = 4 years
		// Note: vesting duration is lower now

		// Deposit, using the vesting seconds as the amount, so that now 3 tokens becomes spendable per second. Note
		// that we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt = big.NewInt(int64(testsuite.VestingDuration4Years.Seconds()))
		sendAmount = new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration4Years)

		// Match the expected balance for the receiver on the Sequencer. We take into consideration previous deposits.
		amountCoin3 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1.Add(amountCoin2).Add(amountCoin3))

		// Calculate expected values
		bridgeParams = s.QueryBridgeParams(s.Ctx())

		expectVestingInfo1 := bridgetypes.NewVestingInfo(
			sdk.NewCoins(amountCoin1.Add(amountCoin2).Add(amountCoin3)), vestingStartTime1.Unix(), vestingEndTime1.Unix(),
		)

		vestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime1.Unix(), vestingAcc.StartTime) // same vesting start time
		s.Require().Equal(vestingEndTime1.Unix(), vestingAcc.EndTime)     // same vesting end time
		s.Require().True((vestingAcc.OriginalVesting).Equal(expectVestingInfo1.OriginalVesting))
		s.Require().Nil(vestingAcc.DelegatedFree)
		s.Require().Nil(vestingAcc.DelegatedVesting)

		// -----------------------------------------------------------------------------------------------
		// Deposit 4:
		// - VestingStartTime = Now()
		// - VestingDuration = 2 year
		// Note: vesting start time is different now

		// Update vesting start time to now
		bridgeParams.VestingStartTime = time.Now()
		s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		})

		// Deposit, using the vesting seconds as the amount, so that now 4 tokens becomes spendable per second. Note
		// that we need to downscale using the v1 to v2 migration ratio.
		vestingSecondsBigInt = big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount = new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer. We take into consideration previous deposits.
		amountCoin4 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1.Add(amountCoin2).Add(amountCoin3).Add(amountCoin4))

		// Calculate expected values
		bridgeParams = s.QueryBridgeParams(s.Ctx())
		vestingStartTime4 := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime4 := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration2Years)

		expectVestingInfo2 := bridgetypes.NewVestingInfo(
			sdk.NewCoins(amountCoin4), vestingStartTime4.Unix(), vestingEndTime4.Unix(),
		)

		multiVestingAcc, err := s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, multiVestingAcc.AccountOwner)
		s.Require().Len(multiVestingAcc.Infos, 2) // two vesting infos now
		s.Require().Equal(multiVestingAcc.Infos[0], expectVestingInfo1)
		s.Require().Equal(multiVestingAcc.Infos[1], expectVestingInfo2)

		// -----------------------------------------------------------------------------------------------
		// Check the resultant vesting account's spendable coins

		// We expect 4 tokens to be spendable per second; four (1, 1, 2) from the first vesting info, and one each from the other.
		now := time.Now()
		expSpendable1 := (now.Unix() - multiVestingAcc.Infos[0].StartTime) * 4
		expSpendable2 := now.Unix() - multiVestingAcc.Infos[1].StartTime
		expTotalSpendable := expSpendable1 + expSpendable2

		// Use a buffer of 18 seconds or ~3 blocks (18*4 = 72 tokens)
		buffer := int64(72)

		// Wait for 2 block to make sure we're past the 'now'
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*20))

		// Spendable should be greater than the amount calculated above, but lower than the buffered amount.
		spendable, err := s.QuerySpendableBalance(s.Ctx(), senderAddress, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable)
		s.Require().LessOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable+buffer)

		// -----------------------------------------------------------------------------------------------
		// Delegate and undelegate the spendable coins

		// Delegate

		delegateCoin := *spendable.Balance
		validator1Address := s.SeqKeys[0].ValAddressSeq
		validator1AddressEth := s.SeqKeys[0].ValAddressEth

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), senderAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", senderAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1, with the spendable coins as a delegation.
		delegateData := testsuite.PackDelegate(delegateCoin.Amount.BigInt(), validator1AddressEth)

		// Confirm that the delegation went through and is as expected.
		delegation, err := s.SendEthTransactionToSequencerInterfaceContract(delegateData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForDelegationBalance(s.Ctx(), 0, senderAddress, validator1Address, delegateCoin)

		// Check account again
		multiVestingAcc, err = s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, multiVestingAcc.AccountOwner)
		s.Require().Len(multiVestingAcc.Infos, 2)
		s.Require().Equal(multiVestingAcc.Infos[0], expectVestingInfo1)
		s.Require().Equal(multiVestingAcc.Infos[1], expectVestingInfo2)

		// --------------------------------------- Re-check spendable

		// We expect 4 tokens to be spendable per second; four (1, 1, 2) from the first vesting info, and one each from the other.
		now = time.Now()
		expSpendable1 = (now.Unix() - multiVestingAcc.Infos[0].StartTime) * 4
		expSpendable2 = now.Unix() - multiVestingAcc.Infos[1].StartTime
		expTotalSpendable = expSpendable1 + expSpendable2 - delegateCoin.Amount.Int64() // subtract delegation

		// Use a buffer of 18 seconds or ~3 blocks (18*4 = 72 tokens)
		buffer = int64(72)

		// Wait for 2 block to make sure we're past the 'now'
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*20))

		// Spendable should be greater than the amount calculated above, but lower than the buffered amount.
		spendable, err = s.QuerySpendableBalance(s.Ctx(), senderAddress, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable)
		s.Require().LessOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable+buffer)

		// --------------------------------------- Undelegate

		// Override unbonding time so that undelegation goes through immediately
		stakingParams := s.QueryStakingParams(s.Ctx())
		stakingParams.UnbondingTime = time.Second
		s.ExecuteGovProposal(&stakingtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *stakingParams,
		})

		// Ensure value updated
		s.Require().Equal(time.Second, s.QueryStakingParams(s.Ctx()).UnbondingTime)

		// Delegate to validator1 with the amount previously delegated.
		undelegateCoin := delegateCoin
		unbondData := testsuite.PackUnbond(undelegateCoin.Amount.BigInt(), validator1AddressEth)

		// Confirm that the undelegation went through by checking for no delegation
		undelegation, err := s.SendEthTransactionToSequencerInterfaceContract(unbondData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, undelegation.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, senderAddress, validator1Address)

		// Wait for undelegation to go through (Note: unbonding time is very small)
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10))

		// Check account again
		multiVestingAcc, err = s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, multiVestingAcc.AccountOwner)
		s.Require().Len(multiVestingAcc.Infos, 2)
		s.Require().Equal(multiVestingAcc.Infos[0], expectVestingInfo1)
		s.Require().Equal(multiVestingAcc.Infos[1], expectVestingInfo2)

		// --------------------------------------- Re-check spendable

		// We expect 4 tokens to be spendable per second; four (1, 1, 2) from the first vesting info, and one each from the other.
		now = time.Now()
		expSpendable1 = (now.Unix() - multiVestingAcc.Infos[0].StartTime) * 4
		expSpendable2 = now.Unix() - multiVestingAcc.Infos[1].StartTime
		expTotalSpendable = expSpendable1 + expSpendable2

		// Use a buffer of 18 seconds or ~3 blocks (18*4 = 72 tokens)
		buffer = int64(72)

		// Wait for 2 block to make sure we're past the 'now'
		s.Require().NoError(s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*20))

		// Spendable should be greater than the amount calculated above, but lower than the buffered amount.
		spendable, err = s.QuerySpendableBalance(s.Ctx(), senderAddress, testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().GreaterOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable)
		s.Require().LessOrEqual(spendable.Balance.Amount.Int64(), expTotalSpendable+buffer)
	})
}
