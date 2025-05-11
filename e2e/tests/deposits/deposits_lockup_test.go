package deposits_test

import (
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/tests/deposits"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *DepositsTestSuite) TestDeposits_SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate() {
	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		deposits.SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate(&s.E2ETestSuite)
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
		s.Require().Nil(ethOwnedVestingAcc.DelegatedFree) // no delegation
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
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.DelegatedFree)) // delegation
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
		s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.DelegatedFree)) // delegation
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
