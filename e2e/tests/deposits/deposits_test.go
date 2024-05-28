package deposits_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := testsuite.ETH_ADDRESSES[0]                // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_ADDRESSES[1]      // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[1] // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[0]    // Deposit receiver; owned by the sender

		// --------------------------------------- Account owned by sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		amount := big.NewInt(200)
		duration := big.NewInt(63072000) // 2 years vesting
		depositData := testsuite.PackDeposit(amount, common.HexToAddress(ownedReceiverAddressSeq), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(amount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		// Make sure that a new Eth owned vesting account was created with the correct details
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		durationNanoseconds := time.Second * time.Duration(duration.Int64())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay).Unix()
		vestingEndTime := bridgeParams.VestingStartTime.Add(durationNanoseconds).Unix()
		amountCoins := sdk.NewCoins(amountCoin)

		ethOwnedContinuousVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedContinuousVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime, ethOwnedContinuousVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime, ethOwnedContinuousVestingAcc.EndTime)
		s.Require().Equal(amountCoins, ethOwnedContinuousVestingAcc.OriginalVesting)

		// --------------------------------------- Account not owned by sender

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), notOwnedReceiverAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate a deposit to an account which is not owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		depositData = testsuite.PackDeposit(amount, common.HexToAddress(notOwnedReceiverAddress), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		s.PollForBalance(s.Ctx(), 10, notOwnedReceiverAddress, amountCoin)

		// Make sure that the created account is a simple base account
		baseAccount, err := s.QueryBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().NoError(err)
		s.Require().Equal(notOwnedReceiverAddressSeq, baseAccount.Address)

		// Try querying the account as a vesting account and assert failure to make sure that no vesting details were
		// stored
		_, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().Error(err)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithNoVesting() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with no vesting and check results", func() {
		validatorAddress := testsuite.ADDRESSES[0]                 // Address of one of the validators
		senderAddress := testsuite.ETH_ADDRESSES[0]                // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_ADDRESSES[1]      // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[1] // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[0]    // Deposit receiver; owned by the sender

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

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		sendAmount := big.NewInt(200)
		duration := big.NewInt(63072000) // 2 years vesting
		depositData := testsuite.PackDeposit(sendAmount, common.HexToAddress(ownedReceiverAddressSeq), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// Make sure that a new Eth owned vesting account was created with the correct details. Previously existing
		// funds should not be vested
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		durationNanoseconds := time.Second * time.Duration(duration.Int64())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay).Unix()
		vestingEndTime := bridgeParams.VestingStartTime.Add(durationNanoseconds).Unix()
		sendAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		vestingAmountCoins := sdk.NewCoins(sendAmountCoin)

		ethOwnedContinuousVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedContinuousVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime, ethOwnedContinuousVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime, ethOwnedContinuousVestingAcc.EndTime)
		s.Require().Equal(vestingAmountCoins, ethOwnedContinuousVestingAcc.OriginalVesting)

		// --------------------------------------- Account not owned by sender

		// Create the account by transferring tokens to it
		to = sdk.MustAccAddressFromBech32(notOwnedReceiverAddressSeq)
		msg = banktypes.NewMsgSend(from, to, initBalance)
		res, err = s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), notOwnedReceiverAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(initBalance, balance.Balances)

		// Generate a deposit to an account which is not owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		depositData = testsuite.PackDeposit(sendAmount, common.HexToAddress(notOwnedReceiverAddress), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly deposited tokens.
		s.PollForBalance(s.Ctx(), 10, notOwnedReceiverAddress, expAmountCoin)

		// Make sure that the created account is a simple base account
		baseAccount, err := s.QueryBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().NoError(err)
		s.Require().Equal(notOwnedReceiverAddressSeq, baseAccount.Address)

		// Try querying the account as a vesting account and assert failure to make sure that no vesting details were
		// stored
		_, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().Error(err)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithVesting() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with vesting and check results", func() {
		validatorAddress := testsuite.ADDRESSES[0]                 // Address of one of the validators
		senderAddress := testsuite.ETH_ADDRESSES[0]                // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_ADDRESSES[1]      // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[1] // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_ADDRESS_SEQ[0]    // Deposit receiver; owned by the sender

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

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		sendAmount := big.NewInt(200)
		duration := big.NewInt(63072000) // 2 years vesting
		depositData := testsuite.PackDeposit(sendAmount, common.HexToAddress(ownedReceiverAddressSeq), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initVestingAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		// Make sure that a new Eth owned vesting account was created with the correct details. Previously existing
		// funds should no longer be vested
		bridgeParams = s.QueryBridgeParams(s.Ctx())
		durationNanoseconds := time.Second * time.Duration(duration.Int64())
		vestingStartTime := bridgeParams.VestingStartTime.Add(testsuite.VestingStartTimeDelay).Unix()
		vestingEndTime := bridgeParams.VestingStartTime.Add(durationNanoseconds).Unix()
		sendAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		vestingAmountCoins := sdk.NewCoins(sendAmountCoin)

		ethOwnedContinuousVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedContinuousVestingAcc.AccountOwner)
		s.Require().Equal(vestingStartTime, ethOwnedContinuousVestingAcc.StartTime)
		s.Require().Equal(vestingEndTime, ethOwnedContinuousVestingAcc.EndTime)
		s.Require().Equal(vestingAmountCoins, ethOwnedContinuousVestingAcc.OriginalVesting)

		// --------------------------------------- Account not owned by sender

		// Send tokens to the account and set it as a vesting account
		to = sdk.MustAccAddressFromBech32(notOwnedReceiverAddressSeq)
		msg = vestingtypes.NewMsgCreateVestingAccount(from, to, initVestingAmountCoins, initVestingEndTime, false)
		res, err = s.SubmitMsgs(msg)
		s.Require().NoError(err)
		s.Require().Zero(res.Code)

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), notOwnedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(initVestingAmountCoins, balance.Balances)

		// Make sure that the account was created with the right vesting.
		continuousVestingAccount, err = s.QueryContinuousVestingAccount(s.Ctx(), notOwnedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(initVestingEndTime, continuousVestingAccount.EndTime)
		s.Require().Equal(initVestingAmountCoins, continuousVestingAccount.OriginalVesting)

		// Generate a deposit to an account which is not owned by the sender.
		// Note: by default the sender is testsuite.ETH_ADDRESSES[0]
		depositData = testsuite.PackDeposit(sendAmount, common.HexToAddress(notOwnedReceiverAddress), duration)
		_, err = s.SendEthTransactionToMockEthereumContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// vested balance and the newly deposited tokens.
		s.PollForBalance(s.Ctx(), 10, notOwnedReceiverAddress, expAmountCoin)

		// Make sure that the account's initial vesting details were not modified
		continuousVestingAccount, err = s.QueryContinuousVestingAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().NoError(err)
		s.Require().Equal(initVestingEndTime, continuousVestingAccount.EndTime)
		s.Require().Equal(initVestingAmountCoins, continuousVestingAccount.OriginalVesting)
	})
}
