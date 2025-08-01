package multi_vesting_accounts_test

import (
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/multi_vesting_accounts"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

const (
	haltHeightDelta    = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	blocksAfterUpgrade = uint64(10) // will wait for this many blocks after the upgrade
	upgradeName        = multi_vesting_accounts.UpgradeName
	fromImageVersion   = "f7cb7f49" // seq-mainnet-1.3
	toImageVersion     = "735051c2" // sequencer with multi-vesting accounts upgrade (main:1.4.0-rc.1)
)

type UpgradesTestSuite struct {
	testsuite.E2ETestSuite
}

func TestUpgradesTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradesTestSuite))
}

func (s *UpgradesTestSuite) SetupTest() {

	s.FuelSequencerDockerImageTag = fromImageVersion

	s.E2ETestSuite.SetupTest()

	// Test suite will build them during test if not found, but better have them on hand
	s.EnsureAllDockerImagesExist(fromImageVersion, toImageVersion)
}

func (s *UpgradesTestSuite) TestUpgrade() {
	s.Run("Assert existing behaviour: create a vesting policy, and deposit funds to it twice", func() {
		senderAddress := s.EthKeys[0].AddressHex           // The depositor on Ethereum
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq // Deposit receiver; owned by the sender

		// Make sure that the balance of the receiver is as expected.
		expectedInitBalance := sdk.NewInt64Coin(testsuite.BridgeDenom, 0)
		balance, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		s.Require().Equal(expectedInitBalance.Amount, balance.Balances.AmountOf(testsuite.BridgeDenom))

		// First deposit with vesting
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer
		amountCoin1 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1)

		// Verify it's a continuous vesting account
		vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)

		// Second deposit with same vesting parameters
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer (doubled)
		amountCoin2 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, amountCoin1.Add(amountCoin2))

		// Verify it's still a continuous vesting account with doubled amount
		vestingAcc, err = s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().True(sdk.NewCoins(amountCoin1.Add(amountCoin2)).Equal(vestingAcc.OriginalVesting))
	})

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

	s.Run("Assert that vesting policy is still present", func() {
		senderAddress := s.EthKeys[0].AddressHex
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq

		// Verify the vesting account still exists and is accessible
		vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().NotNil(vestingAcc.OriginalVesting)
		s.Require().True(vestingAcc.OriginalVesting.AmountOf(testsuite.BridgeDenom).GT(sdkmath.ZeroInt()))
	})

	s.Run("Deposit funds to the vesting policy, and assert that the funds are summed accordingly", func() {
		senderAddress := s.EthKeys[0].AddressHex
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq

		// Get current balance before deposit
		balanceBefore, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		balanceBeforeAmount := balanceBefore.Balances.AmountOf(testsuite.BridgeDenom)

		// Third deposit with same vesting parameters
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer (tripled)
		amountCoin3 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		expectedBalance := balanceBeforeAmount.Add(amountCoin3.Amount)
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, sdk.NewCoin(testsuite.BridgeDenom, expectedBalance))

		// Verify the vesting account has the summed amount
		vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, vestingAcc.AccountOwner)
		s.Require().True(expectedBalance.Equal(vestingAcc.OriginalVesting.AmountOf(testsuite.BridgeDenom)))
	})

	s.Run("Change the vesting_start_time, and ensure vote proposal goes through with bridge param updated", func() {
		// Get current bridge params
		bridgeParams := s.QueryBridgeParams(s.Ctx())
		originalVestingStartTime := bridgeParams.VestingStartTime

		// Update vesting start time to now
		bridgeParams.VestingStartTime = time.Now()
		s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bridgeParams,
		})

		// Confirm the new value
		bridgeParamsAfterProposal := s.QueryBridgeParams(s.Ctx())
		// Compare times in UTC to avoid timezone issues
		s.Require().Equal(bridgeParams.VestingStartTime.UTC(), bridgeParamsAfterProposal.VestingStartTime.UTC())
		s.Require().NotEqual(originalVestingStartTime, bridgeParamsAfterProposal.VestingStartTime)
	})

	s.Run("Deposit a new amount to the vesting policy, and assert that the funds placed into a new vesting account. Ensure the previous policy is still present", func() {
		senderAddress := s.EthKeys[0].AddressHex
		ownedReceiverAddressSeq := s.EthKeys[0].AddressSeq

		// Get current balance before deposit
		balanceBefore, err := s.QueryAllBalances(s.Ctx(), ownedReceiverAddressSeq, nil)
		s.Require().NoError(err)
		balanceBeforeAmount := balanceBefore.Balances.AmountOf(testsuite.BridgeDenom)

		// Fourth deposit with different vesting start time (should create multi-vesting account)
		vestingSecondsBigInt := big.NewInt(int64(testsuite.VestingDuration2Years.Seconds()))
		sendAmount := new(big.Int).Quo(vestingSecondsBigInt, testsuite.MigrationRatio)
		_ = s.DepositTokenToSequencerFromMigrationNoDelegation(sendAmount, testsuite.VestingDuration2Years)

		// Match the expected balance for the receiver on the Sequencer (increased by new amount)
		amountCoin4 := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(vestingSecondsBigInt))
		expectedBalance := balanceBeforeAmount.Add(amountCoin4.Amount)
		s.PollForBalance(s.Ctx(), 10, ownedReceiverAddressSeq, sdk.NewCoin(testsuite.BridgeDenom, expectedBalance))

		// Verify it's now a multi-vesting account
		multiVestingAcc, err := s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ownedReceiverAddressSeq)
		s.Require().NoError(err)
		s.Require().Equal(senderAddress, multiVestingAcc.AccountOwner)
		s.Require().Len(multiVestingAcc.Infos, 2) // two vesting infos now

		// Verify the first vesting info (from before upgrade) is still present
		s.Require().True(multiVestingAcc.Infos[0].OriginalVesting.AmountOf(testsuite.BridgeDenom).GT(sdkmath.ZeroInt()))

		// Verify the second vesting info (from after upgrade) is present
		s.Require().True(multiVestingAcc.Infos[1].OriginalVesting.AmountOf(testsuite.BridgeDenom).GT(sdkmath.ZeroInt()))

		// Verify total original vesting equals expected balance
		totalOriginalVesting := multiVestingAcc.GetOriginalVesting()
		s.Require().True(expectedBalance.Equal(totalOriginalVesting.AmountOf(testsuite.BridgeDenom)))
	})
}
