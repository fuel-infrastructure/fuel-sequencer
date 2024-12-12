package upgrades_test

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/vesting_accounts_staking"
	e2etestsuite "github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	haltHeightDelta    = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	blocksAfterUpgrade = uint64(10) // will wait for this many blocks after the upgrade
	upgradeName        = vesting_accounts_staking.UpgradeName
	fromImageVersion   = "2e65f66" // this image needs to exist for this test to run
	toImageVersion     = "2019e20" // this image needs to exist for this test to run
)

type UpgradesTestSuite struct {
	e2etestsuite.E2ETestSuite
}

func TestUpgradesTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradesTestSuite))
}

func (s *UpgradesTestSuite) SetupTest() {

	s.FuelSequencerDockerImageTag = fromImageVersion

	s.E2ETestSuite.SetupTest()
}

func (s *UpgradesTestSuite) TestUpgradePowerReduction() {

	s.Run("Perform the upgrade", func() {
		height, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before submit upgrade proposal")

		haltHeight := height + haltHeightDelta
		s.Logger().Info("Submitting software upgrade proposal", zap.Uint64("halt_height", haltHeight))
		msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
			Authority: s.GetGovernanceAddress(),
			Plan: upgradetypes.Plan{
				Name:   upgradeName,
				Height: int64(haltHeight),
				Info:   "<dummy-info>",
			},
		}
		s.ExecuteGovProposal(msgUpgrade)

		height, err = s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before upgrade")

		// Wait until just before the upgrade
		err = s.WaitUntilSequencerBlock(s.Ctx(), int(haltHeight-1), time.Second*20)
		s.Require().NoError(err)

		// Hold Ethereum so the Sequencer doesn't get too out-of-sync
		s.PauseEthereum()

		// Ensure the nodes have reached the halt height
		time.Sleep(time.Second * 5)

		// Bring down nodes to prepare for upgrade.
		s.Logger().Info("Stopping all sequencer nodes...")
		s.StopAllSequencerNodes()
		s.Logger().Info("Removing all sequencer nodes...")
		s.RemoveAllSequencerNodes()

		// Resume Ethereum since we're about to resume the Sequencer
		s.UnpauseEthereum()

		// Upgrade version on all nodes and start them back up.
		s.Logger().Info("Starting nodes back up...")
		s.FuelSequencerDockerImageTag = toImageVersion
		s.RunSequencerValidators()

		err = s.WaitForSequencerBlocks(s.Ctx(), int(blocksAfterUpgrade), time.Second*20)
		s.Require().NoError(err, "chain did not produce blocks after upgrade")
	})

	s.Run("Submit deposits on Ethereum to Sequencer accounts that do not exist yet and check results", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(e2etestsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(e2etestsuite.BridgeDenom))

		// Deposit, using the vesting seconds as the amount, so that 1 token becomes spendable per second
		sendAmount := big.NewInt(int64(e2etestsuite.VestingDuration2Years.Seconds()))
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, e2etestsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin := sdk.NewCoin(e2etestsuite.BridgeDenom, sdkmath.NewIntFromBigInt(sendAmount))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin)

		// Calculate expected values
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		vestingStartTime := bridgeParams.VestingStartTime.Add(e2etestsuite.VestingStartTimeDelay)
		vestingEndTime := bridgeParams.VestingStartTime.Add(e2etestsuite.VestingDuration2Years)

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
		// Deposit "(now+20s)-vestingStartTime" worth of tokens, such that we expect these to be stakeable in 20s.
		timeForUnlock := time.Now().Add(time.Second * 20).Sub(vestingStartTime)
		delegateAmount := sdkmath.NewInt(int64(timeForUnlock.Seconds()))
		delegateCoin := sdk.NewCoin(e2etestsuite.BridgeDenom, delegateAmount)
		msgDelegateBz := s.E2ETestSuite.GenerateMsgDelegateBz(delegatorAddress, validator1Address, delegateCoin)
		authorizeData := e2etestsuite.PackAuthorize(msgDelegateBz)

		// Expect a failure because not enough time has elapsed yet
		delegation1, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
		s.Require().NoError(err)
		s.PollForLastEthereumBlockSynced(s.Ctx(), 10, delegation1.BlockNumber.Uint64()) // wait until tx processed
		s.PollForNoDelegation(s.Ctx(), 0, delegatorAddress, validator1Address)

		// Sleep the remaining time so that enough tokens will be spendable
		s.Sleep(vestingStartTime.Add(timeForUnlock).Sub(time.Now()))

		// Confirm that the delegation went through and is as expected.
		delegation2, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
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

		// Generate Authorize event wrapping a MsgUndelegate to validator1 with the amount previously delegated.
		undelegateCoin := delegateCoin
		msgUndelegateBz := s.E2ETestSuite.GenerateMsgUndelegateBz(delegatorAddress, validator1Address, undelegateCoin)
		authorizeData = e2etestsuite.PackAuthorize(msgUndelegateBz)

		// Confirm that the undelegation went through by checking for no delegation
		undelegation, err := s.SendEthTransactionToSequencerInterfaceContract(authorizeData)
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
	})
}
