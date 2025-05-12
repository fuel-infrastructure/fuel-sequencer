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
		s.Logger().Info("Current height before upgrade proposal", zap.Int64("height", int64(height)))

		haltHeight := height + bondHaltHeightDelta
		s.Logger().Info("Submitting software upgrade proposal",
			zap.Uint64("halt_height", haltHeight),
			zap.String("upgrade_name", bond_module.UpgradeName))
		msgUpgrade := &upgradetypes.MsgSoftwareUpgrade{
			Authority: s.GetGovernanceAddress(),
			Plan: upgradetypes.Plan{
				Name:   bond_module.UpgradeName,
				Height: int64(haltHeight),
				Info:   "<dummy-info>",
			},
		}
		s.ExecuteGovProposal(msgUpgrade)

		_, err = s.GetFuelSequencerHeight(s.Ctx())
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
		s.Logger().Info("Initial bond params",
			zap.String("yield_recipient", bondParams.YieldRecipient),
			zap.Any("yield_time", bondParams.YieldTime),
			zap.String("yield_amount", bondParams.YieldAmount.String()))

		// Check that the yield recipient is set to default value
		s.Require().Equal("", bondParams.YieldRecipient)

		// Check that the yield time is set to default value
		s.Require().Nil(bondParams.YieldTime)

		// Check that the yield amount is set to default value
		s.Require().True(bondParams.YieldAmount.IsZero())

		// Get the bond module state
		bondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(bondState)
		s.Logger().Info("Initial bond state",
			zap.Int64("yield_mint_height", bondState.YieldMintHeight))

		// Check that the yield mint height is set to default value
		s.Require().Equal(int64(0), bondState.YieldMintHeight)

		// Propose a new bond params via an expedited proposal
		newYieldRecipient := s.GetGovernanceAddress()
		newYieldTime := time.Now().Add(time.Minute) // 1 minute from now
		newYieldAmount := sdkmath.NewInt(1000000)   // 1M tokens

		s.Logger().Info("Proposing new bond params",
			zap.String("new_yield_recipient", newYieldRecipient),
			zap.Time("new_yield_time", newYieldTime),
			zap.String("new_yield_amount", newYieldAmount.String()))

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
		s.Logger().Info("Updated bond params",
			zap.String("yield_recipient", updatedBondParams.YieldRecipient),
			zap.Any("yield_time", updatedBondParams.YieldTime),
			zap.String("yield_amount", updatedBondParams.YieldAmount.String()))

		s.Require().Equal(newYieldRecipient, updatedBondParams.YieldRecipient)
		s.Require().Equal(newYieldTime.Unix(), updatedBondParams.YieldTime.Unix())
		s.Require().Equal(newYieldAmount, updatedBondParams.YieldAmount)

		// Verify state is still queryable and unchanged
		updatedBondState := s.QueryBondState(s.Ctx())
		s.Require().NotNil(updatedBondState)
		s.Logger().Info("Updated bond state",
			zap.Int64("yield_mint_height", updatedBondState.YieldMintHeight))
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

		s.Logger().Info("Initial state before yield",
			zap.String("initial_supply", initialSupply.String()),
			zap.String("initial_recipient_balance", initialRecipientBalance.Balance.Amount.String()))

		// Verify bond params from previous test case
		bondParams := s.QueryBondParams(s.Ctx())
		s.Logger().Info("Current bond params before yield",
			zap.String("yield_recipient", bondParams.YieldRecipient),
			zap.Any("yield_time", bondParams.YieldTime),
			zap.String("yield_amount", bondParams.YieldAmount.String()))

		// Verify params are set from previous test case
		s.Require().Equal(s.GetGovernanceAddress(), bondParams.YieldRecipient)
		s.Require().NotNil(bondParams.YieldTime)
		s.Require().Equal(sdkmath.NewInt(1000000), bondParams.YieldAmount)

		// Check if yield has already been minted
		bondState := s.QueryBondState(s.Ctx())
		s.Logger().Info("Current bond state",
			zap.Int64("yield_mint_height", bondState.YieldMintHeight))

		if bondState.YieldMintHeight > 0 {
			s.Logger().Info("Yield has already been minted at height",
				zap.Int64("height", bondState.YieldMintHeight))
		} else {
			// Wait until we reach or pass the yield time and yield occurs
			for {
				currentTime := time.Now()
				bondState = s.QueryBondState(s.Ctx())

				// If minting has occurred, verify it happened at the right time
				if bondState.YieldMintHeight > 0 {
					// Get the block time for the mint height
					block, err := s.GetBlockByHeight(s.Ctx(), bondState.YieldMintHeight)
					s.Require().NoError(err, "failed to get block at mint height")
					mintTime := block.Header.Time

					// Log detailed timing information
					s.Logger().Info("Yield minting occurred",
						zap.Int64("height", bondState.YieldMintHeight),
						zap.Time("mint_time", mintTime),
						zap.Time("yield_time", *bondParams.YieldTime),
						zap.String("time_difference", mintTime.Sub(*bondParams.YieldTime).String()),
						zap.Int64("current_block_height", bondState.YieldMintHeight),
						zap.Int64("yield_mint_height", bondState.YieldMintHeight),
					)

					// Verify that minting occurred at or after the intended yield time
					s.Require().True(bondParams.YieldTime.Before(mintTime) || bondParams.YieldTime.Equal(mintTime),
						"yield minting occurred before intended yield time")
					break
				}

				// If we haven't reached yield time yet, wait and continue
				if !currentTime.After(*bondParams.YieldTime) {
					s.Logger().Info("Current time before yield time",
						zap.Time("current_time", currentTime),
						zap.Time("yield_time", *bondParams.YieldTime))
					time.Sleep(time.Second)
					continue
				}

				// We've reached yield time, wait for two blocks to ensure minting occurs in the first block after yield time
				err = s.WaitForSequencerBlocks(s.Ctx(), 2, time.Second*10)
				s.Require().NoError(err, "failed to wait for blocks after yield time")

				// After expected yield time, final check after the upcoming block to assert that minting occurred
				bondState = s.QueryBondState(s.Ctx())
				if bondState.YieldMintHeight > 0 {
					// Get the block time for the mint height
					block, err := s.GetBlockByHeight(s.Ctx(), bondState.YieldMintHeight)
					s.Require().NoError(err, "failed to get block at mint height")
					mintTime := block.Header.Time

					// Log detailed timing information
					s.Logger().Info("Yield minting occurred after waiting",
						zap.Int64("height", bondState.YieldMintHeight),
						zap.Time("mint_time", mintTime),
						zap.Time("yield_time", *bondParams.YieldTime),
						zap.String("time_difference", mintTime.Sub(*bondParams.YieldTime).String()),
						zap.Int64("current_block_height", bondState.YieldMintHeight),
						zap.Int64("yield_mint_height", bondState.YieldMintHeight),
					)

					// Verify that minting occurred at or after the intended yield time
					s.Require().True(
						bondParams.YieldTime.Before(mintTime) || bondParams.YieldTime.Equal(mintTime),
						"yield minting occurred before intended yield time")
					break
				}

				// If we've waited a block after yield time and still no minting, fail
				s.Require().Fail("Yield minting did not occur within one block after yield time")
				return
			}
		}

		// Verify total supply increased by yield amount and log all supply data
		finalSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Logger().Info("Token supply data w.r.t mint",
			zap.String("initial_supply", initialSupply.String()),
			zap.String("yield_amount", bondParams.YieldAmount.String()),
			zap.String("expected_final_supply", initialSupply.Add(bondParams.YieldAmount).String()),
			zap.String("actual_final_supply", finalSupply.String()),
			zap.String("supply_difference", finalSupply.Sub(initialSupply).String()),
		)

		// Verify recipient balance increased by yield amount and log all balance data
		finalRecipientBalance, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.Logger().Info("Recipient balance data w.r.t yield",
			zap.String("initial_recipient_balance", initialRecipientBalance.Balance.Amount.String()),
			zap.String("expected_final_balance", initialRecipientBalance.Balance.Amount.Add(bondParams.YieldAmount).String()),
			zap.String("actual_final_balance", finalRecipientBalance.Balance.Amount.String()),
			zap.String("balance_difference", finalRecipientBalance.Balance.Amount.Sub(initialRecipientBalance.Balance.Amount).String()),
		)

		// Now perform assertions
		s.Require().Equal(initialSupply.Add(bondParams.YieldAmount), finalSupply,
			"total supply should increase by yield amount")
		s.Require().Equal(initialRecipientBalance.Balance.Amount.Add(bondParams.YieldAmount), finalRecipientBalance.Balance.Amount,
			"recipient balance should increase by yield amount")
	})

		// Run the deposit and delegate test
		deposits.SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate(&s.E2ETestSuite)
		s.Logger().Info("Completed deposit and delegate test")
	})
}
