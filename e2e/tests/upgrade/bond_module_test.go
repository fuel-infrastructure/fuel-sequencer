package upgrades_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/bond_module"
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
	s.Run("Params are set as expected", func() {
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

		// The total inflation should be the sum of mint and bond inflation
		totalInflation := s.QueryMintInflation(s.Ctx())
		expectedTotalInflation := mintParams.InflationMax.Add(bondParams.Inflation)
		s.Require().Equal(expectedTotalInflation, totalInflation)
	})

	s.Run("Mint working as expected (same inflation as pre-upgrade)", func() {
		// Get initial balances
		feeCollectorAddr, err := s.QueryModuleAccountAddress(s.Ctx(), "fee_collector")
		s.Require().NoError(err)
		bondAuthorityAddr := s.GetGovernanceAddress() // bond authority is set to governance address
		initialFeeCollectorBalance, err := s.QueryBalance(s.Ctx(), feeCollectorAddr.String(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		initialBondAuthorityBalance, err := s.QueryBalance(s.Ctx(), bondAuthorityAddr, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Get initial supply
		initialSupply, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Wait for some blocks to pass to accumulate inflation
		numBlocks := uint64(10)
		err = s.WaitForSequencerBlocks(s.Ctx(), int(numBlocks), time.Second*20)
		s.Require().NoError(err)

		// Get final balances
		finalFeeCollectorBalance, err := s.QueryBalance(s.Ctx(), feeCollectorAddr.String(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		finalBondAuthorityBalance, err := s.QueryBalance(s.Ctx(), bondAuthorityAddr, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Get final supply
		finalSupply, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Calculate expected inflation
		bondParams := s.QueryBondParams(s.Ctx())
		mintParams := s.QueryMintParams(s.Ctx())
		totalInflation := mintParams.InflationMax.Add(bondParams.Inflation)

		// Calculate expected minted amount
		expectedMintedAmount := initialSupply.Balance.Amount.ToLegacyDec().Mul(totalInflation).Mul(sdkmath.LegacyNewDec(int64(numBlocks))).RoundInt()
		actualMintedAmount := finalSupply.Balance.Amount.Sub(initialSupply.Balance.Amount)

		// Verify total minted amount matches expected inflation
		s.Require().Equal(expectedMintedAmount, actualMintedAmount, "total minted amount should match expected inflation")

		// Calculate expected distribution between fee collector and bond authority
		var expectedFeeCollectorAmount, expectedBondAuthorityAmount sdkmath.Int
		if totalInflation.IsZero() {
			// If total inflation is zero, no tokens should be minted
			expectedFeeCollectorAmount = sdkmath.ZeroInt()
			expectedBondAuthorityAmount = sdkmath.ZeroInt()
		} else {
			mintRatio := mintParams.InflationMax.Quo(totalInflation)
			expectedFeeCollectorAmount = actualMintedAmount.ToLegacyDec().Mul(mintRatio).RoundInt()
			expectedBondAuthorityAmount = actualMintedAmount.Sub(expectedFeeCollectorAmount)
		}

		// Verify fee collector received expected amount
		feeCollectorIncrease := finalFeeCollectorBalance.Balance.Amount.Sub(initialFeeCollectorBalance.Balance.Amount)
		s.Require().Equal(expectedFeeCollectorAmount, feeCollectorIncrease, "fee collector should receive expected amount")

		// Verify bond authority received expected amount
		bondAuthorityIncrease := finalBondAuthorityBalance.Balance.Amount.Sub(initialBondAuthorityBalance.Balance.Amount)
		s.Require().Equal(expectedBondAuthorityAmount, bondAuthorityIncrease, "bond authority should receive expected amount")
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
		newInflation := sdkmath.LegacyMustNewDecFromStr("0.2")
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
		// Get initial supply and sender balance
		sender := s.SeqKeys[0].AddressSeq
		initialSupply, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		initialSenderBalance, err := s.QueryBalance(s.Ctx(), sender, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Amount to burn
		burnAmount := sdk.NewCoins(sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(100)))

		// Create and submit burn message
		msgBurnCoins := &bondtypes.MsgBurnCoins{
			Sender: sender,
			Coins:  burnAmount,
		}
		_, err = s.SubmitMsgs(msgBurnCoins)
		s.Require().NoError(err)

		// Get final balances
		finalSupply, err := s.QueryBalance(s.Ctx(), s.GetGovernanceAddress(), testsuite.BridgeDenom)
		s.Require().NoError(err)
		finalSenderBalance, err := s.QueryBalance(s.Ctx(), sender, testsuite.BridgeDenom)
		s.Require().NoError(err)

		// Verify sender's balance decreased by burn amount
		senderDecrease := initialSenderBalance.Balance.Amount.Sub(finalSenderBalance.Balance.Amount)
		s.Require().Equal(burnAmount[0].Amount, senderDecrease, "sender's balance should decrease by burn amount")

		// Verify total supply decreased by burn amount
		supplyDecrease := initialSupply.Balance.Amount.Sub(finalSupply.Balance.Amount)
		s.Require().Equal(burnAmount[0].Amount, supplyDecrease, "total supply should decrease by burn amount")
	})

	// s.Run("Ensure deposit and delegate working as usual (regression check)", func() {
	// 	// Get initial balances
	// 	sender := s.EthKeys[0]
	// 	validator := s.SeqKeys[0]
	// 	initialSenderBalance, err := s.QueryBalance(s.Ctx(), sender.AddressSeq, testsuite.BridgeDenom)
	// 	s.Require().NoError(err)
	// 	initialDelegation := s.QueryDelegation(s.Ctx(), sender.AddressSeq, validator.ValAddressSeq)

	// 	// Amount to deposit and delegate
	// 	amount := big.NewInt(1000)

	// 	// Deposit and delegate
	// 	receipt := s.DepositAndDelegateTokenToSequencer(amount, common.HexToAddress(validator.ValAddressHex))
	// 	s.Require().NotNil(receipt)

	// 	// Wait for the transaction to be processed
	// 	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, receipt.BlockNumber.Uint64())

	// 	// Verify deposit
	// 	finalSenderBalance, err := s.QueryBalance(s.Ctx(), sender.AddressSeq, testsuite.BridgeDenom)
	// 	s.Require().NoError(err)
	// 	senderIncrease := finalSenderBalance.Balance.Amount.Sub(initialSenderBalance.Balance.Amount)
	// 	s.Require().Equal(sdkmath.NewIntFromBigInt(amount), senderIncrease, "sender's balance should increase by deposit amount")

	// 	// Verify delegation
	// 	finalDelegation := s.QueryDelegation(s.Ctx(), sender.AddressSeq, validator.ValAddressSeq)
	// 	sharesIncrease := finalDelegation.Delegation.Shares.Sub(initialDelegation.Delegation.Shares)
	// 	s.Require().True(sharesIncrease.IsPositive(), "validator shares should increase")
	// })
}
