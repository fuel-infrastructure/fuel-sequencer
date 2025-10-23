package vesting_demo_test

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
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

// vestingDetails holds information about created vesting policies
type vestingDetails struct {
	TotalTokens     *big.Int
	VestingPolicies []vestingPolicy
}

type vestingPolicy struct {
	Amount         *big.Int
	DurationYears  int
	StartOffset    string // e.g., "1 year ago", "2 years from now"
	ExpectedStatus string // e.g., "50% progress", "complete", "not started"
}

func TestVestingDemoTestSuite(t *testing.T) {
	suite.Run(t, new(VestingDemoTestSuite))
}

// SetupSuite overrides Docker image settings to use working versions
func (s *VestingDemoTestSuite) SetupSuite() {
	// Enable proxy for this test suite, as it's required for the explorer service
	s.EnableProxy()

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
		err := s.EnsureEthereumMinimumBalances(s.MinimumEthereumBalance,
			s.NormalAccount, s.SingleVestingAccount, s.MultiVestingAccount)
		s.Require().NoError(err)

		// Setup accounts and collect details
		normalDetails := s.setupNormalAccount(s.NormalAccount)

		// Wait for different timestamps
		s.T().Log("⏳ Waiting for distinct vesting start times...")
		time.Sleep(3 * time.Second)

		singleDetails := s.setupSingleVestingAccount(s.SingleVestingAccount)

		// Wait for different timestamps
		s.T().Log("⏳ Waiting for distinct vesting start times...")
		time.Sleep(3 * time.Second)

		multiDetails := s.setupMultiVestingAccount(s.MultiVestingAccount)

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

		// Normal Account
		s.T().Logf("1. Normal Account (%s):", s.NormalAccount.AddressSeq)
		s.T().Logf("   - %s tokens, no vesting", normalDetails.TotalTokens.String())
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.NormalAccount.AddressSeq)

		// Single Vesting Account
		s.T().Logf("2. Single Vesting Account (%s):", s.SingleVestingAccount.AddressSeq)
		s.T().Logf("   - %s tokens total", singleDetails.TotalTokens.String())
		for _, policy := range singleDetails.VestingPolicies {
			s.T().Logf("   - %s tokens, %d-year vesting (%s) - %s",
				policy.Amount.String(), policy.DurationYears, policy.StartOffset, policy.ExpectedStatus)
		}
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.SingleVestingAccount.AddressSeq)

		// Multi Vesting Account
		s.T().Logf("3. Multi Vesting Account (%s):", s.MultiVestingAccount.AddressSeq)
		s.T().Logf("   - %s tokens total", multiDetails.TotalTokens.String())
		for _, policy := range multiDetails.VestingPolicies {
			s.T().Logf("   - %s tokens, %d-year vesting (%s) - %s",
				policy.Amount.String(), policy.DurationYears, policy.StartOffset, policy.ExpectedStatus)
		}
		s.T().Logf("   - %s/%s/account/%s", explorerURL, chainName, s.MultiVestingAccount.AddressSeq)

		s.T().Log("")
		s.T().Log("⏰ The test environment will keep running...")
		s.T().Log("   Press Ctrl+C to stop when done testing")

		// Keep the test running so the environment stays up for manual testing
		s.keepEnvironmentRunning()
	})
}

func (s *VestingDemoTestSuite) setupNormalAccount(ethKey *testsuite.EthereumKey) vestingDetails {
	s.T().Log("1️⃣ Setting up Normal Account")

	// Use 1000 tokens for easier demo
	depositAmount := big.NewInt(1000)

	// Create deposit with no vesting using testsuite method
	tx := s.DepositTokenToSequencerFromEthereumKey(ethKey, depositAmount)

	// Wait for the transaction to be processed
	s.PollForLastEthereumBlockSynced(s.Ctx(), 10, tx.BlockNumber.Uint64())

	// Verify the deposit was processed
	expectedBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(depositAmount))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedBalance)

	s.T().Logf("✅ Normal account created with %s tokens", depositAmount.String())
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)

	return vestingDetails{
		TotalTokens:     depositAmount,
		VestingPolicies: []vestingPolicy{}, // No vesting policies
	}
}

func (s *VestingDemoTestSuite) setupSingleVestingAccount(ethKey *testsuite.EthereumKey) vestingDetails {
	s.T().Log("2️⃣ Setting up Single Vesting Policy")

	// First setup the normal account part
	normalDetails := s.setupNormalAccount(ethKey)

	// Save original bridge params
	bridgeParams := s.QueryBridgeParams(s.Ctx())
	originalStartTime := bridgeParams.VestingStartTime

	// Set vesting start time for desired progress state
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	bridgeParams.VestingStartTime = oneYearAgo

	// Submit and wait for governance proposal to pass
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after governance proposal
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	// Add vesting tokens
	vestingAmount := big.NewInt(2000)
	sendAmount := new(big.Int).Quo(vestingAmount, testsuite.MigrationRatio)
	vestingDuration := testsuite.VestingDuration2Years

	// Create deposit with vesting using testsuite method
	s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount, vestingDuration)

	// Wait for the latest block to be synced after deposit
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	// Restore original start time
	bridgeParams.VestingStartTime = originalStartTime
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after restoring original params
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	// Calculate total balance
	totalBalance := new(big.Int).Add(normalDetails.TotalTokens, vestingAmount)
	expectedTotalBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(totalBalance))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedTotalBalance)

	// Verify vesting account was created correctly
	vestingAcc, err := s.QueryEthOwnedContinuousVestingAccount(s.Ctx(), ethKey.AddressSeq)
	s.Require().NoError(err)
	s.Require().NotNil(vestingAcc)

	s.T().Logf("✅ Single vesting policy added with %s tokens", vestingAmount.String())
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)

	return vestingDetails{
		TotalTokens: totalBalance,
		VestingPolicies: []vestingPolicy{
			{
				Amount:         vestingAmount,
				DurationYears:  2,
				StartOffset:    "started 1 year ago",
				ExpectedStatus: "~50% progress",
			},
		},
	}
}

func (s *VestingDemoTestSuite) setupMultiVestingAccount(ethKey *testsuite.EthereumKey) vestingDetails {
	s.T().Log("3️⃣ Setting up Multiple Vesting Policies")

	// First setup the normal account
	normalDetails := s.setupNormalAccount(ethKey)

	// Save original bridge params
	bridgeParams := s.QueryBridgeParams(s.Ctx())
	originalStartTime := bridgeParams.VestingStartTime

	var policies []vestingPolicy
	totalVesting := big.NewInt(0)

	// First vesting policy: Complete vesting
	amount1 := big.NewInt(2000)
	sendAmount1 := new(big.Int).Quo(amount1, testsuite.MigrationRatio)

	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	bridgeParams.VestingStartTime = fiveYearsAgo
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after governance proposal
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	s.T().Logf("Adding first vesting policy: %s tokens", amount1.String())
	s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount1, testsuite.VestingDuration4Years)

	// Wait for the latest block to be synced after deposit
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	policies = append(policies, vestingPolicy{
		Amount:         amount1,
		DurationYears:  4,
		StartOffset:    "started 5 years ago",
		ExpectedStatus: "complete",
	})
	totalVesting.Add(totalVesting, amount1)

	// Second vesting policy: Active vesting
	amount2 := big.NewInt(1500)
	sendAmount2 := new(big.Int).Quo(amount2, testsuite.MigrationRatio)

	twoYearsAgo := time.Now().AddDate(-2, 0, 0)
	bridgeParams.VestingStartTime = twoYearsAgo
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after governance proposal
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	s.T().Logf("Adding second vesting policy: %s tokens", amount2.String())
	s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount2, testsuite.VestingDuration4Years)

	// Wait for the latest block to be synced after deposit
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	policies = append(policies, vestingPolicy{
		Amount:         amount2,
		DurationYears:  4,
		StartOffset:    "started 2 years ago",
		ExpectedStatus: "~50% progress",
	})
	totalVesting.Add(totalVesting, amount2)

	// Third vesting policy: Future vesting
	amount3 := big.NewInt(2500)
	sendAmount3 := new(big.Int).Quo(amount3, testsuite.MigrationRatio)

	oneYearFromNow := time.Now().AddDate(1, 0, 0)
	bridgeParams.VestingStartTime = oneYearFromNow
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after governance proposal
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	s.T().Logf("Adding third vesting policy: %s tokens", amount3.String())
	s.DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey, sendAmount3, testsuite.VestingDuration4Years)

	// Wait for the latest block to be synced after deposit
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	policies = append(policies, vestingPolicy{
		Amount:         amount3,
		DurationYears:  4,
		StartOffset:    "starts 1 year from now",
		ExpectedStatus: "not started",
	})
	totalVesting.Add(totalVesting, amount3)

	// Restore original start time for consistency with other tests
	bridgeParams.VestingStartTime = originalStartTime
	s.ExecuteGovProposal(&bridgetypes.MsgUpdateParams{
		Authority: s.GetGovernanceAddress(),
		Params:    *bridgeParams,
	})

	// Wait for the latest block to be synced after restoring original params
	s.PollForLatestEthereumBlockSynced(s.Ctx(), 10)

	// Calculate total balance
	totalBalance := new(big.Int).Add(normalDetails.TotalTokens, totalVesting)
	expectedTotalBalance := sdk.NewCoin(testsuite.BridgeDenom, sdkmath.NewIntFromBigInt(totalBalance))
	s.PollForBalance(s.Ctx(), 10, ethKey.AddressSeq, expectedTotalBalance)

	// Verify multi-vesting account was created
	multiVestingAcc, err := s.QueryEthOwnedMultiContinuousVestingAccount(s.Ctx(), ethKey.AddressSeq)
	if err == nil && multiVestingAcc != nil {
		s.T().Logf("✅ Multi vesting account created with %d vesting policies", len(multiVestingAcc.Infos))
		for i, info := range multiVestingAcc.Infos {
			startTime := time.Unix(info.StartTime, 0)
			endTime := time.Unix(info.EndTime, 0)
			now := time.Now()

			var status string
			if now.Before(startTime) {
				status = "Not Started"
			} else if now.After(endTime) {
				status = "Complete"
			} else {
				progress := float64(now.Unix()-info.StartTime) / float64(info.EndTime-info.StartTime) * 100
				status = fmt.Sprintf("%.1f%% Progress", progress)
			}

			s.T().Logf("   Policy %d: %s tokens, %s, Start: %v, End: %v",
				i+1, info.OriginalVesting.String(), status,
				startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
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

	s.T().Logf("✅ Multi vesting account setup completed (total: %s tokens)", totalBalance.String())
	s.T().Logf("   Account: %s", ethKey.AddressSeq)

	s.verifyAccountSetup(ethKey)

	return vestingDetails{
		TotalTokens:     totalBalance,
		VestingPolicies: policies,
	}
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
	_, err := s.SendEthTransactionToTokenContract(mintData) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Approve V2 tokens for use by sequencer interface contract using the specific key
	approveData := testsuite.PackApproveToken(testsuite.SequencerInterfaceContractAddress, amount)
	_, err = s.SendEthTransactionFromToTokenContract(ethKey.PrivateKey, approveData)
	s.Require().NoError(err)

	// Deposit using the specific key
	depositData := testsuite.PackDeposit(amount)
	receipt, err := s.SendEthTransactionFromToSequencerInterfaceContract(ethKey.PrivateKey, depositData)
	s.Require().NoError(err)

	return receipt
}

// DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey deposits tokens with vesting from a specific ethereum key
func (s *VestingDemoTestSuite) DepositTokenToSequencerFromMigrationNoDelegationFromEthereumKey(ethKey *testsuite.EthereumKey, amount *big.Int, vestingDuration time.Duration) *types.Receipt {
	// Amount needs to be upscaled by 1e9 to counteract the DECIMALS_DOWNSCALING_FACTOR of 1e9 applied by the migrator
	amountToMigrate := new(big.Int).Mul(amount, testsuite.MigrateAmountUpscalingFactor)

	// First, mint V1 tokens to the account for migration using the deployer key (following testsuite pattern)
	mintV1Data := testsuite.PackMintMigratedToken(ethKey.Address, amountToMigrate)
	_, err := s.SendEthTransactionToMigratedTokenContract(mintV1Data) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Approve V1 tokens for use by token migrator using specific key
	approveData := testsuite.PackApproveMigratedToken(testsuite.TokenMigratorContractAddress, amountToMigrate)
	_, err = s.SendEthTransactionFromToMigratedTokenContract(ethKey.PrivateKey, approveData)
	s.Require().NoError(err)

	// Mint V2 tokens to token migrator using the deployer key (following testsuite pattern)
	migratorMintData := testsuite.PackMintToken(testsuite.TokenMigratorContractAddress, amountToMigrate)
	_, err = s.SendEthTransactionToTokenContract(migratorMintData) // Uses s.EthKeys[0] by default
	s.Require().NoError(err)

	// Migrate V1 tokens to V2 tokens using specific key
	depositData := testsuite.PackMigrate(amountToMigrate, new(big.Int).SetInt64(int64(vestingDuration.Seconds())))
	receipt, err := s.SendEthTransactionFromToTokenMigratorContract(ethKey.PrivateKey, depositData)
	s.Require().NoError(err)

	return receipt
}

// PollForLatestEthereumBlockSynced waits for the latest Ethereum block to be synced
func (s *VestingDemoTestSuite) PollForLatestEthereumBlockSynced(ctx context.Context, maxAttempts uint64) {
	// Get the current latest block number
	latestBlockNum := s.QueryLastEthereumBlockSynced(ctx)

	// Wait for that block to be synced
	s.PollForLastEthereumBlockSynced(ctx, maxAttempts, latestBlockNum)
}
