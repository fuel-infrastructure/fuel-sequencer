package deposits_test

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup_AndAuthorizeDelegate() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit and delegate, using the vesting seconds as the amount, so that 1 token becomes spendable per second
		sendAmount := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
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
		delegatorAddress := s.EthKeys[0].AddressHex

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		timeForUnlock := time.Second * 60
		delegateAmount, ok := sdkmath.NewIntFromString("60")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := testsuite.PackAuthorize(msgDelegateBz)

		// Ensure that not too much time has elapsed, with a 20-second buffer
		timeUntilUnlock := vestingStartTime.Add(timeForUnlock).Sub(time.Now())
		if timeUntilUnlock < time.Second*20 {
			s.Fail(fmt.Sprintf("tokens will unlock in %s; not enough buffer!", timeUntilUnlock.String()))
		} else {
			s.Logger().Info(fmt.Sprintf("tokens will unlock in %s", timeUntilUnlock.String()))
		}

		// Expect a failure because not enough time has elapsed yet
		delegation1, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation1.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validator1Address)

		// Sleep the remaining time so that enough tokens will be spendable
		s.Sleep(vestingStartTime.Add(timeForUnlock).Sub(time.Now()))

		// Confirm that the delegation went through and is as expected.
		_, err = s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator1Address, delegateCoin)

		// Check account again
		ethOwnedVestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().True(sdk.NewCoins(delegateCoin).Equal(ethOwnedVestingAcc.DelegatedVesting)) // delegation
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
		sendAmount := big.NewInt(1000)
		validatorAddress := common.HexToAddress(validatorAddressHex)
		_ = s.DepositTokenToSequencerFromMigration(sendAmount, validatorAddress, testsuite.VestingDuration2Years)

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
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.DelegatedVesting)) // delegation
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockupShorterThanVestingStartTimeDelay() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// --------------------------------------- Deposit

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit and delegate, using the vesting seconds as the amount, so that 1 token becomes spendable per second
		sendAmount := big.NewInt(int64(testsuite.VestingDuration6Months.Seconds()))
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration6Months)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime // Note how no vesting start time delay applied here
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration6Months)

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
		delegatorAddress := s.EthKeys[0].AddressHex

		// Make sure that there is no pre-existing delegation between the delegator and validator1.
		delegationRaw, err := s.QueryDelegationRaw(s.Ctx(), delegatorAddress, validator1Address)
		s.Require().Nil(delegationRaw)
		s.Require().ErrorContains(
			err,
			fmt.Sprintf("delegation with delegator %s not found for validator %s", delegatorAddress, validator1Address),
		)

		// Generate Authorize event wrapping a MsgDelegate to validator1.
		timeForUnlock := time.Second * 60
		delegateAmount, ok := sdkmath.NewIntFromString("60")
		s.Require().True(ok)
		delegateCoin := sdk.NewCoin(testsuite.BridgeDenom, delegateAmount)
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := testsuite.PackAuthorize(msgDelegateBz)

		// Ensure that not too much time has elapsed, with a 20-second buffer
		timeUntilUnlock := vestingStartTime.Add(timeForUnlock).Sub(time.Now())
		if timeUntilUnlock < time.Second*20 {
			s.Fail(fmt.Sprintf("tokens will unlock in %s; not enough buffer!", timeUntilUnlock.String()))
		} else {
			s.Logger().Info(fmt.Sprintf("tokens will unlock in %s", timeUntilUnlock.String()))
		}

		// Expect a failure because not enough time has elapsed yet
		delegation1, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation1.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validator1Address)

		// Sleep the remaining time so that enough tokens will be spendable
		s.Sleep(vestingStartTime.Add(timeForUnlock).Sub(time.Now()))

		// Confirm that the delegation went through and is as expected.
		_, err = s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)
		s.PollForDelegationBalance(s.Ctx(), 10, delegatorAddress, validator1Address, delegateCoin)

		// Check account again
		ethOwnedVestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().True(sdk.NewCoins(delegateCoin).Equal(ethOwnedVestingAcc.DelegatedVesting)) // delegation
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
		sendAmount := big.NewInt(200)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

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
		sendAmount := big.NewInt(200)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

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
