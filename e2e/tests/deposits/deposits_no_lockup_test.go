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

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_NoLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := testsuite.ETH_KEYS[0].AddressHex              // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_KEYS[1].AddressHex    // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_KEYS[1].AddressSeq // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_KEYS[0].AddressSeq    // Deposit receiver; owned by the sender

		// --------------------------------------- Account owned by sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate a deposit to an account owned by the sender.
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		amount := big.NewInt(200)
		mintData := testsuite.PackMint(common.HexToAddress(senderAddress), amount)
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(amount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(amount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		ethOwnedBaseAcc, err := s.QueryEthOwnedBaseAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedBaseAcc.AccountOwner)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("Cannot proceed in E2E test because we can only deposit to account owned by the sender!")
		return

		// Make sure that the balance of the receiver is as expected.
		balance, err = s.QueryAllBalances(s.Ctx(), notOwnedReceiverAddress, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// Generate a deposit to an account which is not owned by the sender.
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		mintData = testsuite.PackMint(common.HexToAddress(senderAddress), amount)
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData = testsuite.PackTransferAndCall(amount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer
		s.PollForBalance(s.Ctx(), 10, notOwnedReceiverAddress, amountCoin)

		// Make sure that the created account is a simple base account
		baseAccount, err := s.QueryBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().NoError(err)
		s.Require().Equal(notOwnedReceiverAddressSeq, baseAccount.Address)

		// Try querying the account and assert failure to make sure that the account is not an EthOwnedBaseAccount
		_, err = s.QueryEthOwnedBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().Error(err)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithNoVesting_NoLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with no vesting and check results", func() {
		validatorAddress := testsuite.SEQ_ADDRESSES[0]                 // Address of one of the validators
		senderAddress := testsuite.ETH_KEYS[0].AddressHex              // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_KEYS[1].AddressHex    // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_KEYS[1].AddressSeq // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_KEYS[0].AddressSeq    // Deposit receiver; owned by the sender

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
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		sendAmount := big.NewInt(200)
		mintData := testsuite.PackMint(common.HexToAddress(senderAddress), sendAmount)
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(sendAmount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		ethOwnedBaseAcc, err := s.QueryEthOwnedBaseAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedBaseAcc.AccountOwner)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("Cannot proceed in E2E test because we can only deposit to account owned by the sender!")
		return

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
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		mintData = testsuite.PackMint(common.HexToAddress(senderAddress), sendAmount) // TODO!
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData = testsuite.PackTransferAndCall(sendAmount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly deposited tokens.
		s.PollForBalance(s.Ctx(), 10, notOwnedReceiverAddress, expAmountCoin)

		// Make sure that the created account is a simple base account
		baseAccount, err := s.QueryBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().NoError(err)
		s.Require().Equal(notOwnedReceiverAddressSeq, baseAccount.Address)

		// Try querying the account and assert failure to make sure that the account is not an EthOwnedBaseAccount
		_, err = s.QueryEthOwnedBaseAccount(s.Ctx(), notOwnedReceiverAddress)
		s.Require().Error(err)
	})
}

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsExistWithVesting_NoLockup() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that exist with vesting and check results", func() {
		validatorAddress := testsuite.SEQ_ADDRESSES[0]                 // Address of one of the validators
		senderAddress := testsuite.ETH_KEYS[0].AddressHex              // The depositor on Ethereum
		notOwnedReceiverAddress := testsuite.ETH_KEYS[1].AddressHex    // Deposit receiver; not owned by the sender
		notOwnedReceiverAddressSeq := testsuite.ETH_KEYS[1].AddressSeq // Seq addr corresponding to notOwnedReceiverAddress
		ownedReceiverAddressSeq := testsuite.ETH_KEYS[0].AddressSeq    // Deposit receiver; owned by the sender

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
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		sendAmount := big.NewInt(200)
		mintData := testsuite.PackMint(common.HexToAddress(senderAddress), sendAmount)
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData := testsuite.PackTransferAndCall(sendAmount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
		s.Require().NoError(err)

		// Match the expected balance for the receiver on the Sequencer. This should be the summation of the initial
		// balance and the newly vested tokens.
		expAmount := new(big.Int).Add(sendAmount, initVestingAmount)
		expAmountCoin := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(expAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, expAmountCoin)

		ethOwnedBaseAcc, err := s.QueryEthOwnedBaseAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, ethOwnedBaseAcc.AccountOwner)

		// --------------------------------------- Account not owned by sender

		s.Logger().Warn("Cannot proceed in E2E test because we can only deposit to account owned by the sender!")
		return

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
		// Note: by default the sender is testsuite.ETH_KEYS[0]
		mintData = testsuite.PackMint(common.HexToAddress(senderAddress), sendAmount) // TODO!
		_, err = s.SendEthTransactionToTokenContract(mintData)
		s.Require().NoError(err)
		depositData = testsuite.PackTransferAndCall(sendAmount)
		_, err = s.SendEthTransactionToTokenContract(depositData)
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
