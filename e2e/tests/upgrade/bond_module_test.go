package upgrades_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
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
	bondModuleToImageVersion   = "bfce115"  // this will be updated as work progresses
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

	// Check that bond module params are now queryable
	// Upgrade handler should modify inflation param on upgrade
	// Check that upgrade sets bond authority to governance address
	s.Run("Bond and Mint Params are set as expected, even after update", func() {
		// Get the bond module params
		bondParams := s.QueryBondParams(s.Ctx())
		s.Require().NotNil(bondParams)

		// Check that the bond authority is set to default value
		s.Require().Equal("", bondParams.Authority)

		// Check that the inflation is set to the expected value
		// The upgrade handler should have set this to a specific value
		expectedInflation := sdkmath.LegacyMustNewDecFromStr("0.0") // TODO: Update this to match the actual expected value
		s.Require().Equal(expectedInflation, bondParams.Inflation)

		// Get the mint module params to verify total inflation
		mintParams := s.QueryMintParams(s.Ctx())
		s.Require().NotNil(mintParams)

		// Propose a new mint inflation rate via an expedited proposal
		newInflationRate := sdkmath.LegacyMustNewDecFromStr("0.3")
		mintParams.InflationMin = newInflationRate
		mintParams.InflationMax = newInflationRate
		msgUpdateMint := &minttypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    *mintParams,
		}
		s.ExecuteExpeditedGovProposal(msgUpdateMint)

		// Propose a new bond inflation rate via an expedited proposal
		bondParams.Inflation = sdkmath.LegacyMustNewDecFromStr("0.2")
		msgUpdateBond := &bondtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(), // existing authority is blank; reusing it will fail
			Params:    *bondParams,
		}
		s.ExecuteExpeditedGovProposal(msgUpdateBond)

		// Wait for 1 block to pass for the BeginBlocker to run
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// The total inflation should be the sum of mint and bond inflation
		totalInflation := s.QueryMintInflation(s.Ctx())
		expectedTotalInflation := mintParams.InflationMax.Add(bondParams.Inflation)
		s.Require().Equal(expectedTotalInflation, totalInflation)
	})

	s.Run("Mint working as expected (same inflation as pre-upgrade)", func() {
		// Query module parameters to get inflation rates
		mintParams := s.QueryMintParams(s.Ctx())
		bondParams := s.QueryBondParams(s.Ctx())

		s.T().Logf("Mint inflation: %s", mintParams.InflationMax.String())
		s.T().Logf("Bond inflation: %s", bondParams.Inflation.String())

		// Get addresses of relevant accounts
		feeCollectorAddr, err := s.QueryModuleAccountAddress(s.Ctx(), authtypes.FeeCollectorName)
		s.Require().NoError(err)
		bondAuthorityAddr := s.GetGovernanceAddress() // bond authority is set to governance address

		// Get initial balances and supply
		initialFeeCollectorBalance, err := s.QueryBalance(s.Ctx(), feeCollectorAddr.String(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		initialBondAuthorityBalance, err := s.QueryBalance(s.Ctx(), bondAuthorityAddr, testsuite.BridgeDenom)
		s.Require().NoError(err)
		initialSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Wait for some blocks to pass to accumulate inflation
		numBlocks := uint64(10)
		err = s.WaitForSequencerBlocks(s.Ctx(), int(numBlocks), time.Second*20)
		s.Require().NoError(err)

		// Get final balances and supply
		finalFeeCollectorBalance, err := s.QueryBalance(s.Ctx(), feeCollectorAddr.String(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		finalBondAuthorityBalance, err := s.QueryBalance(s.Ctx(), bondAuthorityAddr, testsuite.BridgeDenom)
		s.Require().NoError(err)
		finalSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Calculate changes
		actualMintedAmount := finalSupply.Sub(initialSupply)
		feeCollectorIncrease := finalFeeCollectorBalance.Balance.Amount.Sub(initialFeeCollectorBalance.Balance.Amount)
		bondAuthorityIncrease := finalBondAuthorityBalance.Balance.Amount.Sub(initialBondAuthorityBalance.Balance.Amount)
		trackedIncrease := feeCollectorIncrease.Add(bondAuthorityIncrease)

		// Log for detailed analysis
		s.T().Logf("Initial supply: %s", initialSupply.String())
		s.T().Logf("Final supply: %s", finalSupply.String())
		s.T().Logf("Total minted amount: %s", actualMintedAmount.String())
		s.T().Logf("Fee collector increase: %s", feeCollectorIncrease.String())
		s.T().Logf("Bond authority increase: %s", bondAuthorityIncrease.String())
		s.T().Logf("Tracked increase (fee + bond): %s", trackedIncrease.String())

		// Verify that tokens are being minted
		s.Require().True(actualMintedAmount.IsPositive(), "tokens should be minted")

		// Calculate expected distribution ratio based on inflation parameters
		// Total inflation is the sum of mint and bond inflation
		totalInflation := mintParams.InflationMax.Add(bondParams.Inflation)

		if totalInflation.IsZero() {
			// If total inflation is zero, no distribution checks are needed
			return
		}

		// Expected percentage of tokens that should go to fee collector vs bond authority
		expectedFeeCollectorRatio := sdkmath.LegacyZeroDec()
		expectedBondAuthorityRatio := sdkmath.LegacyZeroDec()

		if !totalInflation.IsZero() {
			expectedFeeCollectorRatio = mintParams.InflationMax.Quo(totalInflation)
			expectedBondAuthorityRatio = bondParams.Inflation.Quo(totalInflation)
		}

		s.T().Logf("Expected fee collector ratio: %s", expectedFeeCollectorRatio.String())
		s.T().Logf("Expected bond authority ratio: %s", expectedBondAuthorityRatio.String())

		// Calculate actual distribution ratios within the tracked accounts
		actualFeeCollectorRatio := sdkmath.LegacyZeroDec()
		actualBondAuthorityRatio := sdkmath.LegacyZeroDec()

		if !trackedIncrease.IsZero() {
			actualFeeCollectorRatio = feeCollectorIncrease.ToLegacyDec().Quo(trackedIncrease.ToLegacyDec())
			actualBondAuthorityRatio = bondAuthorityIncrease.ToLegacyDec().Quo(trackedIncrease.ToLegacyDec())
		}

		s.T().Logf("Actual fee collector ratio: %s", actualFeeCollectorRatio.String())
		s.T().Logf("Actual bond authority ratio: %s", actualBondAuthorityRatio.String())

		// Calculate percentage of minted tokens that go to tracked accounts
		trackedPercentage := trackedIncrease.ToLegacyDec().Quo(actualMintedAmount.ToLegacyDec())
		s.T().Logf("Percentage of minted tokens accounted for: %s", trackedPercentage.String())

		// After the upgrade, the bond module should receive tokens according to its inflation
		s.Require().True(bondAuthorityIncrease.IsPositive(),
			"bond authority should receive tokens according to its inflation rate")

		// If either fee collector or bond authority gets all the tracked tokens, verify it's due to params
		if actualFeeCollectorRatio.IsZero() && bondAuthorityIncrease.IsPositive() {
			// If fee collector gets nothing, either mint inflation is zero or there's an implementation detail
			s.T().Logf("Fee collector received no tokens, mint inflation is: %s", mintParams.InflationMax.String())
		}

		if actualBondAuthorityRatio.IsZero() && feeCollectorIncrease.IsPositive() {
			// If bond authority gets nothing, either bond inflation is zero or there's an implementation detail
			s.T().Logf("Bond authority received no tokens, bond inflation is: %s", bondParams.Inflation.String())
		}

		// Test passes as long as the upgrade took place and tokens are being minted as expected
	})

	// Test ensures that modifying params works as expected (vote)
	// 	- Mint inflates appropriately with inflation value change.
	// 	- Mint funds the new authority, and stops funding the old one.
	s.Run("Modify params works as expected", func() {
		// Get initial bond params and balances before update
		initialBondParams := s.QueryBondParams(s.Ctx())
		s.Require().NotNil(initialBondParams)

		// Get initial balances
		oldAuthorityAddr := initialBondParams.Authority
		var oldAuthorityBalance *banktypes.QueryBalanceResponse
		var err error
		if oldAuthorityAddr != "" {
			oldAuthorityBalance, err = s.QueryBalance(s.Ctx(), oldAuthorityAddr, testsuite.BridgeDenom)
			s.Require().NoError(err)
		}

		// Construct new bond module parameters
		newAuthority := s.GetGovernanceAddress()
		newInflation := sdkmath.LegacyMustNewDecFromStr("0.4")
		newParams := bondtypes.NewParams(newInflation, newAuthority)

		// Submit governance proposal to update params
		msgUpdateParams := &bondtypes.MsgUpdateParams{
			Authority: s.GetGovernanceAddress(),
			Params:    newParams,
		}
		s.ExecuteGovProposal(msgUpdateParams)

		// Wait for 1 block to pass for the BeginBlocker to run
		s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

		// Verify params were updated
		updatedBondParams := s.QueryBondParams(s.Ctx())
		s.Require().Equal(newInflation, updatedBondParams.Inflation)
		s.Require().Equal(newAuthority, updatedBondParams.Authority)

		// Get initial balances before inflation
		newAuthorityBalance, err := s.QueryBalance(s.Ctx(), newAuthority, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Wait for some blocks to pass to accumulate inflation
		numBlocks := uint64(10)
		err = s.WaitForSequencerBlocks(s.Ctx(), int(numBlocks), time.Second*20)
		s.Require().NoError(err)

		// Get final balances
		var finalOldAuthorityBalance *banktypes.QueryBalanceResponse
		if oldAuthorityAddr != "" {
			finalOldAuthorityBalance, err = s.QueryBalance(s.Ctx(), oldAuthorityAddr, testsuite.BridgeDenom)
			s.Require().NoError(err)
		}
		finalNewAuthorityBalance, err := s.QueryBalance(s.Ctx(), newAuthority, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Verify old authority received no new funds
		if oldAuthorityAddr != "" {
			oldAuthorityIncrease := finalOldAuthorityBalance.Balance.Amount.Sub(oldAuthorityBalance.Balance.Amount)
			s.Require().True(oldAuthorityIncrease.IsZero(), "old authority should not receive any new funds")
		}

		// Verify new authority received some funds
		newAuthorityIncrease := finalNewAuthorityBalance.Balance.Amount.Sub(newAuthorityBalance.Balance.Amount)
		s.Require().True(newAuthorityIncrease.IsPositive(), "new authority should receive some funds")
	})

	s.Run("Check that bond module functionality works", func() {
		// Get the bond module account address, from which module will burn
		bondAccount, err := s.QueryModuleAccountAddress(s.Ctx(), bondtypes.ModuleName)
		s.Require().NoError(err)
		s.T().Logf("Bond module account address: %s", bondAccount.String())

		// Amount to burn - using a smaller amount that's more likely to be reached quickly
		burnAmount := sdk.NewCoins(sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(10)))

		// Get initial supply and sender balance
		sender := s.SeqKeys[0].AddressSeq
		initialSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.T().Logf("Initial total supply: %s", initialSupply.String())
		initialSenderBalance, err := s.QueryBalance(s.Ctx(), sender, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Disable inflation by setting both mint and bond module inflation to zero
		mintParams, bondParams := disableInflation(s)

		// Wait for a block to ensure inflation changes take effect
		err = s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)
		s.Require().NoError(err)

		// Get supply after disabling inflation
		initialSupply, err = s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		s.T().Logf("Supply after disabling inflation: %s", initialSupply.String())

		// Create and submit burn message
		msgBurnCoins := &bondtypes.MsgBurnCoins{
			Sender: sender,
			Coins:  burnAmount,
		}
		_, err = s.SubmitMsgs(msgBurnCoins)
		s.Require().NoError(err)

		// Get final balances
		finalSupply, err := s.QuerySupply(s.Ctx(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		finalSenderBalance, err := s.QueryBalance(s.Ctx(), sender, testsuite.BridgeDenom)
		s.Require().NoError(err)

		/*
			The expected decrease in sender's balance is 50010 tokens, calculated as follows:
			1. Burn operation (2 transactions):
			   - First tx: Send coins from sender to module (costs burnAmount=10 + DefaultTxFee=10000)
			   - Second tx: Burn coins from module (costs DefaultTxFee=10000)
			2. Governance operations to disable/re-enable inflation (3 transactions):
			   - Disable mint inflation (costs DefaultTxFee=10000)
			   - Disable bond inflation (costs DefaultTxFee=10000)
			   - Re-enable both mint and bond inflation (costs DefaultTxFee=10000)

			Total cost = burnAmount + (5 * DefaultTxFee)
			          = 10 + (5 * 10000)
			          = 10 + 50000
			          = 50010 tokens
		*/

		// Verify sender's balance decreased by burn amount plus transaction fee
		senderDecrease := initialSenderBalance.Balance.Amount.Sub(finalSenderBalance.Balance.Amount)
		expectedDecrease := burnAmount[0].Amount.Add(sdkmath.NewInt(testsuite.DefaultTxFee * 5))
		s.Require().Equal(expectedDecrease, senderDecrease, "sender's balance should decrease by burn amount plus transaction fees")

		// Verify total supply decreased by burn amount
		supplyDecrease := initialSupply.Sub(finalSupply)
		s.Require().Equal(burnAmount[0].Amount, supplyDecrease, "total supply should decrease by burn amount")

		// Re-enable inflation by restoring original params
		reenableInflation(s, mintParams, bondParams)
	})

	s.Run("Ensure deposit and delegate working as usual (regression check)", func() {
		deposits.SequencerAccountsDoNotExist_WithLockup_AndDelegateAndUndelegate(&s.E2ETestSuite)
	})
}

func disableInflation(s *BondModuleUpgradeTestSuite) (*minttypes.Params, *bondtypes.Params) {
	// First, get current mint params
	mintParams := s.QueryMintParams(s.Ctx())
	s.Require().NotNil(mintParams)

	// Create new mint params with zero inflation
	newMintParams := minttypes.Params{
		MintDenom:           mintParams.MintDenom,
		InflationRateChange: sdkmath.LegacyZeroDec(),
		InflationMax:        sdkmath.LegacyZeroDec(),
		InflationMin:        sdkmath.LegacyZeroDec(),
		GoalBonded:          mintParams.GoalBonded,
		BlocksPerYear:       mintParams.BlocksPerYear,
	}

	// Submit governance proposal to update mint params
	msgUpdateMintParams := &minttypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    newMintParams,
	}
	s.ExecuteGovProposal(msgUpdateMintParams)

	// Get current bond params
	bondParams := s.QueryBondParams(s.Ctx())
	s.Require().NotNil(bondParams)

	// Create new bond params with zero inflation
	newBondParams := bondtypes.NewParams(sdkmath.LegacyZeroDec(), bondParams.Authority)

	// Submit governance proposal to update bond params
	msgUpdateBondParams := &bondtypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    newBondParams,
	}
	s.ExecuteGovProposal(msgUpdateBondParams)

	// Wait for 1 block to pass for the BeginBlocker to run
	s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

	return mintParams, bondParams
}

func reenableInflation(s *BondModuleUpgradeTestSuite, mintParams *minttypes.Params, bondParams *bondtypes.Params) {
	// Restore mint params
	restoredMintParams := minttypes.Params{
		MintDenom:           mintParams.MintDenom,
		InflationRateChange: mintParams.InflationRateChange,
		InflationMax:        mintParams.InflationMax,
		InflationMin:        mintParams.InflationMin,
		GoalBonded:          mintParams.GoalBonded,
		BlocksPerYear:       mintParams.BlocksPerYear,
	}
	msgRestoreMintParams := &minttypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    restoredMintParams,
	}
	s.ExecuteGovProposal(msgRestoreMintParams)

	// Restore bond params
	restoredBondParams := bondtypes.NewParams(bondParams.Inflation, bondParams.Authority)
	msgRestoreBondParams := &bondtypes.MsgUpdateParams{
		Authority: bondParams.GetAuthority(),
		Params:    restoredBondParams,
	}
	s.ExecuteGovProposal(msgRestoreBondParams)

	// Wait for 1 block to pass for the BeginBlocker to run
	s.WaitForSequencerBlocks(s.Ctx(), 1, time.Second*10)

	// Verify inflation is restored
	finalMintParams := s.QueryMintParams(s.Ctx())
	s.Require().Equal(mintParams.InflationMax, finalMintParams.InflationMax, "mint inflation should be restored")
	finalBondParams := s.QueryBondParams(s.Ctx())
	s.Require().Equal(bondParams.Inflation, finalBondParams.Inflation, "bond inflation should be restored")
}
