package upgrades_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/bond_module"
	deposits "github.com/fuel-infrastructure/fuel-sequencer/e2e/tests/deposits"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bondtypes "github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const (
	bondHaltHeightDelta        = uint64(25) // will propose upgrade this many blocks in the future; must be > voting period
	bondBlocksAfterUpgrade     = uint64(10) // will wait for this many blocks after the upgrade
	bondModuleFromImageVersion = "7d60123"  // this image needs to exist for this test to run
	bondModuleToImageVersion   = "37183fa"  // this will be updated as work progresses
)

type BondModuleUpgradeTestSuite struct {
	testsuite.E2ETestSuite
}

func TestBondModuleUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(BondModuleUpgradeTestSuite))
}

func (s *BondModuleUpgradeTestSuite) SetupTest() {
	s.FuelSequencerDockerImageTag = bondModuleFromImageVersion
	s.E2ETestSuite.SetupTest()
}

func (s *BondModuleUpgradeTestSuite) TestBondModuleUpgrade() {
	s.Run("Perform the upgrade", func() {
		height, err := s.GetFuelSequencerHeight(s.Ctx())
		s.Require().NoError(err, "error fetching height before submit upgrade proposal")

		haltHeight := height + bondHaltHeightDelta
		s.Logger().Info("Submitting software upgrade proposal", zap.Uint64("halt_height", haltHeight))
		msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
			Authority: s.GetGovernanceAddress(),
			Plan: upgradetypes.Plan{
				Name:   bond_module.UpgradeName,
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
		s.FuelSequencerDockerImageTag = bondModuleToImageVersion
		s.RunSequencerValidators()

		err = s.WaitForSequencerBlocks(s.Ctx(), int(bondBlocksAfterUpgrade), time.Second*20)
		s.Require().NoError(err, "chain did not produce blocks after upgrade")
	})

	// Check that bond module params and state are now queryable
	// Upgrade handler should modify params on upgrade
	s.Run("Bond Params and State start as default, are updatable and queryable", func() {
		// Get the bond module params
		bondParams := s.QueryBondParams(s.Ctx())
		s.Require().NotNil(bondParams)

		// Check that the yield recipient is set to default value
		s.Require().Equal("", bondParams.YieldRecipient)

		// Check that the yield time is set to default value
		s.Require().Nil(bondParams.YieldTime)

		// Check that the yield amount is set to default value
		s.Require().True(bondParams.YieldAmount.IsNil())

		// Get the bond module state
		bondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(bondState)

		// Check that the yield mint height is set to default value
		s.Require().Equal(int64(0), bondState.YieldMintHeight)

		// Propose a new bond params via an expedited proposal
		newYieldRecipient := s.GetGovernanceAddress()
		newYieldTime := time.Now().Add(time.Hour * 24) // 24 hours from now
		newYieldAmount := sdkmath.NewInt(1000000)      // 1M tokens

		bondParams.YieldRecipient = newYieldRecipient
		bondParams.YieldTime = &newYieldTime
		bondParams.YieldAmount = newYieldAmount

		msgUpdateBond := &bondtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *bondParams,
		}
		s.ExecuteExpeditedGovProposal(msgUpdateBond)

		// Wait for 1 block to pass for the BeginBlocker to run
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// Verify params were updated
		updatedBondParams := s.QueryBondParams(s.Ctx())
		s.Require().Equal(newYieldRecipient, updatedBondParams.YieldRecipient)
		s.Require().Equal(newYieldTime.Unix(), updatedBondParams.YieldTime.Unix())
		s.Require().Equal(newYieldAmount, updatedBondParams.YieldAmount)

		// Verify state is still queryable and unchanged
		updatedBondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(updatedBondState)
		s.Require().Equal(int64(0), updatedBondState.YieldMintHeight, "state should remain unchanged after params update")
	})

	// With the updated params, confirm that the bond module:
	// 1. Mints the requested amount of tokens
	// 2. The total supply is increased by the minted amount
	// 3. Yields the requested amount of tokens to the yield recipient
	// 4. The State of the bond module is updated to show the block height of the yield time
	s.Run("Bond module yields as intended", func() {
		// Get initial state
		initialSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		initialRecipientBalance, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Verify bond params from previous test case
		bondParams := s.QueryBondParams(s.Ctx())
		s.Require().Equal(s.GetGovernanceAddress(), bondParams.YieldRecipient)
		s.Require().NotNil(bondParams.YieldTime)
		s.Require().Equal(sdkmath.NewInt(1000000), bondParams.YieldAmount)

		// Wait for yield time and a block to ensure yield is minted
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// Verify total supply increased by yield amount
		finalSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().Equal(initialSupply.Add(bondParams.YieldAmount), finalSupply)

		// Verify recipient balance increased by yield amount
		finalRecipientBalance, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Require().Equal(initialRecipientBalance.Balance.Amount.Add(bondParams.YieldAmount), finalRecipientBalance.Balance.Amount)

		// Verify bond module state is updated with yield mint height
		bondState := s.QueryBondState(s.Ctx())
		s.Require().NotZero(bondState.YieldMintHeight)
	})

	s.Run("Ensure deposit and delegate working as usual (regression check)", func() {
		deposits.SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate(&s.E2ETestSuite)
	})
}
