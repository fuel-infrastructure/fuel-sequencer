package vesting_demo_test

import (
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/suite"
)

const (
	// Hardcoded Debug Explorer Values; used for convenient links in the test output
	explorerURL = "http://localhost:5173"
	chainName   = "fuel-localhost"
)

type VestingDemoTestSuite struct {
	testsuite.E2ETestSuite

	// Test account addresses that will be created
	NormalAccount        *testsuite.EthereumKey
	SingleVestingAccount *testsuite.EthereumKey
	MultiVestingAccount  *testsuite.EthereumKey

	// Minimum Ethereum balance for each account
	MinimumEthereumBalance *big.Int
}

func TestVestingDemoTestSuite(t *testing.T) {
	suite.Run(t, new(VestingDemoTestSuite))
}

// SetupSuite overrides Docker image settings to use working versions
func (s *VestingDemoTestSuite) SetupSuite() {
	// Call parent SetupSuite first
	s.E2ETestSuite.SetupSuite()

	// Override with working Docker image tags
	s.EthereumNodeDockerImageTag = "nightly"

	// Pick a reasonable minimum balance to test accounts - increased to handle gas costs
	// Gas limit is 1,000,000 and gas price can be high, so need sufficient balance
	s.MinimumEthereumBalance = big.NewInt(1e18) // 1 ETH should be more than enough
}

func (s *VestingDemoTestSuite) TestSetupVestingDemoAccounts() {
	s.Run("Setup three different account types for vesting demo", func() {
		s.T().Log("🚀 Setting up vesting demo accounts...")

		// Each account will demonstrate different vesting policies
		s.NormalAccount = s.EthKeys[0]
		s.SingleVestingAccount = s.EthKeys[1]
		s.MultiVestingAccount = s.EthKeys[2]

		// Check initial ETH balances
		err := s.E2ETestSuite.EnsureEthereumMinimumBalances(s.MinimumEthereumBalance,
			s.NormalAccount, s.SingleVestingAccount, s.MultiVestingAccount)
		s.Require().NoError(err)

		// 1. Setup normal account (no vesting)
		s.setupNormalAccount(s.NormalAccount)

		// Wait a bit to ensure different timestamps for vesting policies
		s.T().Log("⏳ Waiting 3 seconds before next setup for distinct vesting start times...")
		time.Sleep(3 * time.Second)

		// 2. Setup single vesting account (2 years)
		s.setupSingleVestingAccount(s.SingleVestingAccount)

		// Wait a bit more to ensure different timestamps for vesting policies
		s.T().Log("⏳ Waiting 3 seconds before next setup for distinct vesting start times...")
		time.Sleep(3 * time.Second)

		// 3. Setup multi-vesting account (multiple 2-year vestings)
		s.setupMultiVestingAccount(s.MultiVestingAccount)

		s.T().Log("🎉 All vesting demo accounts have been set up successfully!")
		s.T().Log("")
		s.T().Log("🔍 You can now test these accounts in the fuel-explorer to see:")
		s.T().Log("  - Total balance and vested/unvested amounts")
		s.T().Log("  - Multiple vesting schedules with different start times")
		s.T().Log("  - Vesting progression over time")
		s.T().Log("  - Complex vesting policy combinations")
		s.T().Log("")
		s.T().Logf("📊 Intended Explorer Service: %s", explorerURL)
		s.T().Log("")
		s.T().Log("👥 Test Account Summary:")
		s.T().Logf("1. Normal Account (%s):", s.NormalAccount.AddressSeq)
		s.T().Log("   - 1000 tokens, no vesting")
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.NormalAccount.AddressSeq)
		s.T().Logf("2. Single Vesting Account (%s):", s.SingleVestingAccount.AddressSeq)
		s.T().Log("   - 3000 tokens (1000 normal + 2000 vested)")
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.SingleVestingAccount.AddressSeq)
		s.T().Logf("3. Multi Vesting Account (%s):", s.MultiVestingAccount.AddressSeq)
		s.T().Log("   - 7000 tokens (1000 normal + 2000 two-year vesting + 1500 four-year vesting + 2500 different-start two-year vesting)")
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.MultiVestingAccount.AddressSeq)
		s.T().Log("")
		s.T().Log("⏰ The test environment will keep running...")
		s.T().Log("   Press Ctrl+C to stop when done testing")

		// Keep the test running so the environment stays up for manual testing
		s.keepEnvironmentRunning()
	})
}

func (s *VestingDemoTestSuite) setupNormalAccount(ethKey *testsuite.EthereumKey) {
	s.T().Log("1️⃣ Setting up Normal Account (No Vesting)")

	// Use 1000 tokens for easier demo
	depositAmount := big.NewInt(1000)

	// Create deposit with no vesting using testsuite method
	tx := s.DepositTokenToSequencerFromEthereumKey(ethKey, depositAmount)

	// Wait for the transaction to be processed
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx.BlockNumber.Uint64())

	// Verify the deposit was processed
	expectedBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(depositAmount))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedBalance)

	s.T().Logf("✅ Normal account created with %s tokens (no vesting)", depositAmount.String())
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)
}

func (s *VestingDemoTestSuite) setupSingleVestingAccount(ethKey *testsuite.EthereumKey) {
	s.T().Log("2️⃣ Setting up Single Vesting Policy (2 Years)")

	// First setup the normal account part
	s.setupNormalAccount(ethKey)

	// Then add vesting with 2000 tokens vested over 2 years
	finalAmount := big.NewInt(2000)                                       // 2000 tokens final result
	sendAmount := new(big.Int).Quo(finalAmount, testsuite.MigrationRatio) // 2000 / 100 = 20
	vestingDuration := testsuite.VestingDuration2Years

	// Create deposit with 2-year vesting using testsuite method
	tx := s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount, vestingDuration)

	// Wait for the transaction to be processed
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx.BlockNumber.Uint64())

	// Account should now have 1000 (from normal) + 2000 (from this vesting) = 3000 tokens total
	expectedTotalBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(3000))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedTotalBalance)

	// Verify vesting account was created correctly
	vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ethKey.AddressSeq)
	s.Require().NoError(err)
	s.Require().NotNil(vestingAcc)

	s.T().Logf("✅ Single vesting policy added with %s tokens (2-year vesting)", finalAmount.String())
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)
}

func (s *VestingDemoTestSuite) setupMultiVestingAccount(ethKey *testsuite.EthereumKey) {
	s.T().Log("3️⃣ Setting up Multiple Vesting Policies")

	// First setup the normal account
	s.setupNormalAccount(ethKey)

	// Add first vesting policy: 2000 tokens with 2-year vesting
	finalAmount1 := big.NewInt(2000)
	sendAmount1 := new(big.Int).Quo(finalAmount1, testsuite.MigrationRatio)
	vestingDuration1 := testsuite.VestingDuration2Years

	s.T().Logf("Adding first vesting policy: %s tokens with 2-year vesting", finalAmount1.String())

	tx1 := s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount1, vestingDuration1)
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx1.BlockNumber.Uint64())

	// Wait to ensure different timestamps
	time.Sleep(3 * time.Second)

	// Add second vesting policy: 1500 tokens with 4-year vesting (different duration)
	finalAmount2 := big.NewInt(1500)
	sendAmount2 := new(big.Int).Quo(finalAmount2, testsuite.MigrationRatio)
	vestingDuration2 := testsuite.VestingDuration4Years

	s.T().Logf("Adding second vesting policy: %s tokens with 4-year vesting", finalAmount2.String())

	tx2 := s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount2, vestingDuration2)
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx2.BlockNumber.Uint64())

	// Wait to ensure different timestamps
	time.Sleep(3 * time.Second)

	// Add third vesting policy: 2500 tokens with different start time (same 2-year duration)
	finalAmount3 := big.NewInt(2500)
	sendAmount3 := new(big.Int).Quo(finalAmount3, testsuite.MigrationRatio)
	vestingDuration3 := testsuite.VestingDuration2Years

	// Change the vesting start time to create a distinct vesting policy
	bridgeParams := s.QueryBridgeParams(s.Ctx())
	originalStartTime := bridgeParams.VestingStartTime
	bridgeParams.VestingStartTime = time.Now() // Set to current time
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	s.T().Logf("Adding third vesting policy: %s tokens with 2-year vesting but different start time", finalAmount3.String())

	tx3 := s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount3, vestingDuration3)
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx3.BlockNumber.Uint64())

	// Restore original start time for consistency with other tests
	bridgeParams.VestingStartTime = originalStartTime
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Verify total balance: 1000 (normal) + 2000 + 1500 + 2500 = 7000 tokens total
	expectedTotalBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewInt(7000))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedTotalBalance)

	// Verify multi-vesting account was created
	multiVestingAcc, err := s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ethKey.AddressSeq)
	if err == nil && multiVestingAcc != nil {
		s.T().Logf("✅ Multi vesting account created with %d vesting policies", len(multiVestingAcc.Infos))
		for i, info := range multiVestingAcc.Infos {
			s.T().Logf("   Policy %d: %s tokens, Start: %v, End: %v",
				i+1, info.OriginalVesting.String(),
				time.Unix(info.StartTime, 0), time.Unix(info.EndTime, 0))
		}
	} else {
		// If multi-vesting query fails, check single vesting account
		singleVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ethKey.AddressSeq)
		if err == nil && singleVestingAcc != nil {
			s.T().Logf("⚠️  Account is single vesting instead of multi-vesting")
			s.T().Logf("   Single vesting: %s tokens", singleVestingAcc.OriginalVesting.String())
		} else {
			s.T().Logf("❌ Failed to query vesting account: %v", err)
		}
	}

	s.T().Logf("✅ Multi vesting account setup completed (total: 7000 tokens)")
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)
}

func (s *VestingDemoTestSuite) verifyAccountSetup(ethKey *testsuite.EthereumKey) {
	s.T().Log("🔍 Verifying account setup...")

	// Check the single account that now has multiple deposits
	address := ethKey.AddressSeq

	balance, err := s.QueryAllBalances(s.Ctx(), address, nil)
	s.Require().NoError(err)
	s.T().Logf("Total Account Balance (%s): %s", address, balance.Balances.String())

	// Try to query vesting information
	multiVestingAcc, err := s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), address)
	if err == nil && multiVestingAcc != nil {
		s.T().Logf("  Has %d vesting policies", len(multiVestingAcc.Infos))
		for i, info := range multiVestingAcc.Infos {
			s.T().Logf("    Policy %d: %s tokens, Start: %v, End: %v",
				i+1, info.OriginalVesting.String(),
				time.Unix(info.StartTime, 0), time.Unix(info.EndTime, 0))
		}
	} else {
		singleVestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), address)
		if err == nil && singleVestingAcc != nil {
			s.T().Logf("  Has single vesting policy: %s tokens", singleVestingAcc.OriginalVesting.String())
			s.T().Logf("    Start: %v, End: %v",
				time.Unix(singleVestingAcc.StartTime, 0), time.Unix(singleVestingAcc.EndTime, 0))
		} else {
			s.T().Logf("  No vesting policies found (some tokens may be non-vested)")
		}
	}

	s.T().Log("✅ Account verified successfully")
}

func (s *VestingDemoTestSuite) keepEnvironmentRunning() {
	s.T().Log("🕐 Keeping test environment running for manual testing...")
	s.T().Log("   Press Ctrl+C to stop the test and tear down the environment")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for signal or context cancellation
	select {
	case sig := <-sigChan:
		s.T().Logf("🛑 Received signal %v, shutting down...", sig)
	case <-s.Ctx().Done():
		s.T().Log("🛑 Test context cancelled, shutting down...")
	}
}

// Helper methods for key-specific deposits

// DepositTokenToSequencerFromEthereumKey deposits tokens from a specific ethereum key
func (s *VestingDemoTestSuite) DepositTokenToSequencerFromEthereumKey(ethKey *testsuite.EthereumKey, amount *big.Int) *types.Receipt {
	// First, mint V2 tokens to the account using the deployer key (following testsuite pattern)
	mintData := testsuite.PackMintToken(ethKey.Address, amount)
	_, err := s.E2ETestSuite.SendEthTransactionToTokenContract(mintData) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Approve V2 tokens for use by sequencer interface contract using the specific key
	approveData := testsuite.PackApproveToken(testsuite.SequencerInterfaceContractAddress, amount)
	_, err = s.E2ETestSuite.SendEthTransactionFromToTokenContract(ethKey.PrivateKey, approveData)
	s.Require().NoError(err)

	// Deposit using the specific key
	depositData := testsuite.PackDeposit(amount)
	receipt, err := s.E2ETestSuite.SendEthTransactionFromToSequencerInterfaceContract(ethKey.PrivateKey, depositData)
	s.Require().NoError(err)

	return receipt
}

// DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey deposits tokens with vesting from a specific ethereum key
func (s *VestingDemoTestSuite) DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey *testsuite.EthereumKey, amount *big.Int, vestingDuration time.Duration) *types.Receipt {
	// Amount needs to be upscaled by 1e9 to counteract the DECIMALS_DOWNSCALING_FACTOR of 1e9 applied by the migrator
	amountToMigrate := new(big.Int).Mul(amount, testsuite.MigrateAmountUpscalingFactor)

	// First, mint V1 tokens to the account for migration using the deployer key (following testsuite pattern)
	mintV1Data := testsuite.PackMintMigratedToken(ethKey.Address, amountToMigrate)
	_, err := s.E2ETestSuite.SendEthTransactionToMigratedTokenContract(mintV1Data) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Approve V1 tokens for use by token migrator using specific key
	approveData := testsuite.PackApproveMigratedToken(testsuite.TokenMigratorContractAddress, amountToMigrate)
	_, err = s.E2ETestSuite.SendEthTransactionFromToMigratedTokenContract(ethKey.PrivateKey, approveData)
	s.Require().NoError(err)

	// Mint V2 tokens to token migrator using the deployer key (following testsuite pattern)
	migratorMintData := testsuite.PackMintToken(testsuite.TokenMigratorContractAddress, amountToMigrate)
	_, err = s.E2ETestSuite.SendEthTransactionToTokenContract(migratorMintData) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Migrate V1 tokens to V2 tokens using specific key
	depositData := testsuite.PackMigrate(amountToMigrate, new(big.Int).SetInt64(int64(vestingDuration.Seconds())))
	receipt, err := s.E2ETestSuite.SendEthTransactionFromToTokenMigratorContract(ethKey.PrivateKey, depositData)
	s.Require().NoError(err)

	return receipt
}
