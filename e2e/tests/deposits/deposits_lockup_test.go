package deposits_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// --------------------------------------- Account owned by sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Deposit!
		sendAmount := big.NewInt(200)
		_ = s.DepositTokenToSequencerFromMigration(sendAmount)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("E2E test ends here because we can only deposit to accounts owned by the sender at the moment")
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithNoVesting_WithLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with no vesting and check results", func() {
		validatorAddress := s.SeqKeys[0].AddressSeq        // Address of one of the validators
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// --------------------------------------- Account owned by sender

		// Create the account by transferring tokens to it
		from := sdk.MustAccAddressFromBech32(validatorAddress)
		to := sdk.MustAccAddressFromBech32(ownedReceiverAddressSeq)
		initAmount := big.NewInt(100)
		initBalanceCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, initAmount.Int64())
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
		_ = s.DepositTokenToSequencerFromMigration(sendAmount)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// However, the OriginalVesting amount will be the deposited amount
		expOriginalVesting := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(expOriginalVesting).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("E2E test ends here because we can only deposit to accounts owned by the sender at the moment")
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithVesting_WithLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with vesting and check results", func() {
		validatorAddress := s.SeqKeys[0].AddressSeq        // Address of one of the validators
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// --------------------------------------- Account owned by sender

		// Send tokens to the account and set it as a vesting account
		from := sdk.MustAccAddressFromBech32(validatorAddress)
		to := sdk.MustAccAddressFromBech32(ownedReceiverAddressSeq)
		initVestingAmount := big.NewInt(100)
		initVestingAmountCoin := sdk.NewInt64Coin(testsuite.BridgeDenom, initVestingAmount.Int64())
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
		_ = s.DepositTokenToSequencerFromMigration(sendAmount)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initVestingAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// However, the OriginalVesting amount will be the deposited amount
		expOriginalVesting := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))

		// Calculate expected values
		bridgeParams = s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(testsuite.VestingDuration)

		ethOwnedVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
		s.Require().True(sdk.NewCoins(expOriginalVesting).Equal(ethOwnedVestingAcc.OriginalVesting))
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree)
		s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("E2E test ends here because we can only deposit to accounts owned by the sender at the moment")
	})
}
