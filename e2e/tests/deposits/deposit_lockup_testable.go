package deposits

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

func SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate(s *testsuite.E2ETestSuite) {
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
	s.Sleep(vestingStartTime.Add(timeForUnlock).Sub(time.Now()))

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
	s.Require().True(sdk.NewCoins(delegateCoin).Equal(ethOwnedVestingAcc.DelegatedFree)) // delegation
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
	s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

	// Check account again
	ethOwnedVestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
	s.Require().NoError(err)
	s.Require().Equal(senderAddress, ethOwnedVestingAcc.AccountOwner)
	s.Require().Equal(vestingStartTime.Unix(), ethOwnedVestingAcc.StartTime)
	s.Require().Equal(vestingEndTime.Unix(), ethOwnedVestingAcc.EndTime)
	s.Require().True(sdk.NewCoins(amountCoin).Equal(ethOwnedVestingAcc.OriginalVesting))
	s.Require().Nil(ethOwnedVestingAcc.DelegatedFree) // back to zero
	s.Require().Nil(ethOwnedVestingAcc.DelegatedVesting)
}
