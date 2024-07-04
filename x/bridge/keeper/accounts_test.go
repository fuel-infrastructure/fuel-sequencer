package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	minttypes "github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"
	"github.com/stretchr/testify/require"
)

// Ref: https://etherscan.io/address/0x0000000000000000000000000000000000000000
func TestNullAddressIsAsExpected(t *testing.T) {
	require.Equal(t, keeper.NullEthereumAddress, "0x0000000000000000000000000000000000000000")
}

func (s *KeeperTestSuite) TestGenerateSequencerAccountFromEthereumAddress() {
	accAddress, err := s.App.BridgeKeeper.GenerateSequencerAddressFromEthereumAddress(testutiltypes.TestEthAddr1Str)
	s.Require().NoError(err)

	s.Require().Equal(testutiltypes.TestSeqAddr1Str, accAddress.String())
}

func (s *KeeperTestSuite) TestGetSequencerAccountFromEthereumAddress() {

	// The first account number depends on the number of module accounts created.
	firstAccNumber := uint64(len(s.App.AccountKeeper.GetModulePermissions()))

	seqAddr1BaseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: firstAccNumber,
		Sequence:      testutiltypes.FirstAccountSequence,
	}

	// Helper durations.
	years1 := time.Hour * 24 * 365
	years2 := years1 * 2
	years100 := years1 * 100

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t0Plus1Year := t0.Add(years1)  // accounts for vesting start time delay
	t0Plus2Years := t0.Add(years2) // used for 2-year vesting duration
	someTimeWaaaayInTheFuture, _ := time.Parse(time.DateOnly, "2030-01-01")

	// Helper token amounts.
	token200 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	token150 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 150))
	token100 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	token50 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))

	type fnArgs struct {
		ethAddress      string
		vestingDuration time.Duration
		totalCoins      sdk.Coins
	}
	testsCases := []struct {
		name                 string
		precreateAccount     sdk.AccountI
		blockTime            time.Time
		vestingStartTime     time.Time
		fundAccount          sdk.Coins
		args                 fnArgs
		isAccountAsExpected  testutil.AccountValidator
		expectSpendableCoins sdk.Coins
		expectErrMsg         string
	}{
		//
		// --------- Test cases with basic invalid values
		//
		{
			name: "invalid eth address => err",
			args: fnArgs{
				ethAddress: "invalid_eth_address",
			},
			expectErrMsg: "invalid Ethereum address format (invalid_eth_address)",
		},
		{
			name:             "vesting duration < vesting start time delay (1 year) => err",
			blockTime:        t0,
			vestingStartTime: t0,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years1 - 1,
				totalCoins:      token100,
			},
			expectErrMsg: "must be greater than vesting start time delay, got 8759h59m59.999999999s <= 8760h0m0s",
		},
		{
			name:             "vesting duration == vesting start time delay (1 year) => err",
			blockTime:        t0,
			vestingStartTime: t0,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years1,
				totalCoins:      token100,
			},
			expectErrMsg: "must be greater than vesting start time delay, got 8760h0m0s <= 8760h0m0s",
		},
		//
		// --------- Test cases with no precreated account
		//
		{
			name:      "deposit with no vesting and no coins => EthOwnedBaseAccount",
			blockTime: t0,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      nil,
			},
			isAccountAsExpected:  testutil.MatchesEthOwnedAccRaw(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			expectSpendableCoins: nil,
		},
		{
			name:        "deposit with no vesting and some coins => EthOwnedBaseAccount",
			blockTime:   t0,
			fundAccount: token100,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  testutil.MatchesEthOwnedAccRaw(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			expectSpendableCoins: token100,
		},
		{
			// blockTime:              t0
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is before the actual start time, so we expect no tokens to be available.
			name:             "deposit with vesting starting in the future => EthOwnedContinuousVestingAccount",
			blockTime:        t0,
			vestingStartTime: t0,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years2,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(),
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:     seqAddr1BaseAcc,
						OriginalVesting: token100,
						EndTime:         t0Plus2Years.Unix(),
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: nil, // none of the vesting tokens are available
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens are available.
			name:             "deposit with vesting half-way => EthOwnedContinuousVestingAccount",
			blockTime:        t0Plus1Year.Add(years1 / 2),
			vestingStartTime: t0,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years2,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(),
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:     seqAddr1BaseAcc,
						OriginalVesting: token100,
						EndTime:         t0Plus2Years.Unix(),
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token50, // half of the 100 vesting tokens are available
		},
		//
		// --------- Test cases with overriding of a precreated account
		//
		{
			name:             "deposit with no vesting overrides BaseAccount => EthOwnedBaseAccount",
			precreateAccount: seqAddr1BaseAcc,
			blockTime:        t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  testutil.MatchesEthOwnedAccRaw(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			expectSpendableCoins: token200, // all the 200 tokens are available
		},
		{
			// blockTime:              t0
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is before the actual start time, so we expect no tokens to be available in precreated account.
			// But the vesting account will get overridden by an EthOwnedBaseAccount and all tokens become available.
			name: "deposit with no vesting overrides ContinuousVestingAccount => EthOwnedBaseAccount",
			precreateAccount: &vestingtypes.ContinuousVestingAccount{
				StartTime: t0Plus1Year.Unix(),
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,
					EndTime:         t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
				},
			},
			blockTime:        t0,
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  testutil.MatchesEthOwnedAccRaw(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			expectSpendableCoins: token200, // all the 200 tokens are available
		},
		{
			name:             "deposit with no vesting builds on EthOwnedBaseAccount => EthOwnedBaseAccount",
			precreateAccount: types.NewEthOwnedBaseAccount(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			blockTime:        t0,
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  testutil.MatchesEthOwnedAccRaw(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			expectSpendableCoins: token200, // precreated account's 100 plus newly deposited 100
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens will be available.
			// The newly deposited tokens are available and the existing ones are not affected.
			name: "deposit with no vesting builds on existing EthOwnedContinuousVestingAccount but does not affect " +
				"the vesting details => EthOwnedContinuousVestingAccount",
			precreateAccount: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(), // this should be untouched
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  token100,            // this should be untouched
						DelegatedFree:    token50,             // this should be untouched
						DelegatedVesting: token50,             // this should be untouched
						EndTime:          t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			blockTime:        t0Plus1Year.Add(years1 / 2), // half-way through vesting duration
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(), // this was untouched
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  token100,            // this was untouched
						DelegatedFree:    token50,             // this was untouched
						DelegatedVesting: token50,             // this was untouched
						EndTime:          t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token200, // balance - vesting + delegatedVesting = 200 - 50 + 50 = 200
			//
			// If this value is confusing, and you expected the spendable tokens to be 150, look at it this way:
			// - We explicitly funded the account with 200 tokens.
			// - At the same time we're saying that it has 50 tokens that are vesting and delegated (DelegatedVesting).
			// - We're also saying that it has 50 tokens that are vested and delegated (DelegatedFree).
			//
			// This means that in reality we implicitly funded the account with 300 tokens, not 200.
			//
			// Out of the 300 tokens:
			//
			// - Point of view 1:
			//   - 200 are in the balance
			//   - 100 are staked
			// - Point of view 2:
			//   - 50 are vesting (of which 50 staked)
			//   - 50 are vested (of which 50 staked)
			//   - 200 are available [apart from the vesting information]
			//
			// The 200 comes from the 200 that are available.
			//
			// The definition of LockedCoins: "vesting coins that are not delegated"
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/x/bank/types/vesting.go#L11-L12
			// The definition of SpendableCoins: "total balance minus locked coins"
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/x/bank/types/vesting.go#L14-L16
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens will be available.
			// However, we override the funded BaseAccount, so only newly deposited tokens will be vesting.
			name:             "deposit with vesting half-way overrides BaseAccount => EthOwnedContinuousVestingAccount",
			precreateAccount: seqAddr1BaseAcc,
			blockTime:        t0Plus1Year.Add(years1 / 2), // half-way through vesting duration
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years2,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(),
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:     seqAddr1BaseAcc,
						OriginalVesting: token100,            // only 100 are vesting
						EndTime:         t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token150, // precreated account's 100 plus half of newly vested tokens
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens will be available.
			// However, we override the funded ContinuousVestingAccount, so only newly deposited tokens will be vesting.
			name: "deposit with vesting half-way overrides ContinuousVestingAccount => EthOwnedContinuousVestingAccount",
			precreateAccount: &vestingtypes.ContinuousVestingAccount{
				StartTime: t0Plus1Year.Unix(),
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:      seqAddr1BaseAcc,
					OriginalVesting:  token100,                         // will be overwritten
					DelegatedFree:    token50,                          // will be overwritten
					DelegatedVesting: token50,                          // will be overwritten
					EndTime:          someTimeWaaaayInTheFuture.Unix(), // will be overwritten
				},
			},
			blockTime:        t0Plus1Year.Add(years1 / 2), // half-way through vesting duration
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years2,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(), // precreated account's vesting start time is disregarded
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  token100, // of which half are vested
						DelegatedFree:    nil,
						DelegatedVesting: nil,
						EndTime:          t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token150, // precreated account's 100 plus half of the vested tokens
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens will be available.
			// However, we override the funded EthOwnedBaseAccount, so only newly deposited tokens will be vesting.
			name:             "deposit with vesting half-way overrides EthOwnedBaseAccount => EthOwnedContinuousVestingAccount",
			precreateAccount: types.NewEthOwnedBaseAccount(seqAddr1BaseAcc, testutiltypes.TestEthAddr1Str),
			blockTime:        t0Plus1Year.Add(years1 / 2), // half-way through vesting duration
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years2,
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(),
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:     seqAddr1BaseAcc,
						OriginalVesting: token100,            // of which half are vested
						EndTime:         t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token150, // precreated account's 100 plus half of the vested tokens
		},
		{
			// blockTime:              t0 + 1.5 years
			// vestingStartTime:       t0
			// actualVestingStartTime: t0 + 1 year
			// actualVestingEndTime:   t0 + 2 years
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens will be available.
			// The specified vesting duration is ignored, in favor of the existing EthOwnedContinuousVestingAccount's.
			name: "deposit with vesting half-way builds on existing EthOwnedContinuousVestingAccount => " +
				"updated OriginalVesting and retained vesting values => EthOwnedContinuousVestingAccount",
			precreateAccount: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(), // this should be untouched
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  token100,
						DelegatedFree:    token50,             // this should be untouched
						DelegatedVesting: token50,             // this should be untouched
						EndTime:          t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			blockTime:        t0Plus1Year.Add(years1 / 2), // half-way through vesting duration
			vestingStartTime: t0,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      testutiltypes.TestEthAddr1Str,
				vestingDuration: years100, // NB: this gets ignored if account is EthOwnedContinuousVestingAccount
				totalCoins:      token100,
			},
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					StartTime: t0Plus1Year.Unix(), // this was untouched
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  token200,            // of which half are vested
						DelegatedFree:    token50,             // this was untouched
						DelegatedVesting: token50,             // this was untouched
						EndTime:          t0Plus2Years.Unix(), // 1 year lock + 1 year vesting
					},
				},
				testutiltypes.TestEthAddr1Str,
			),
			expectSpendableCoins: token150, // balance - vesting + delegatedVesting = 200 - 100 + 50 = 150
			//
			// If this value is confusing, and you expected the spendable tokens to be 100, look at it this way:
			// - We explicitly funded the account with 200 tokens.
			// - At the same time we're saying that it has 50 tokens that are vesting and delegated (DelegatedVesting).
			// - We're also saying that it has 50 tokens that are vested and delegated (DelegatedFree).
			//
			// This means that in reality we implicitly funded the account with 300 tokens, not 200.
			//
			// Out of the 300 tokens:
			//
			// - Point of view 1:
			//   - 200 are in the balance
			//   - 100 are staked
			// - Point of view 2:
			//   - 100 are vesting (of which 50 staked)
			//   - 100 are vested (of which 50 staked)
			//   - 100 are available [apart from the vesting information]
			//
			// The 150 comes from the 100 that are available and the 50 which are vested but not staked.
			//
			// The definition of LockedCoins: "vesting coins that are not delegated"
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/x/bank/types/vesting.go#L11-L12
			// The definition of SpendableCoins: "total balance minus locked coins"
			// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/x/bank/types/vesting.go#L14-L16
		},
	}

	for _, tc := range testsCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			ctx := s.Ctx().WithBlockTime(tc.blockTime)

			// Set vesting start time
			params := s.App.BridgeKeeper.GetParams(ctx)
			params.VestingStartTime = tc.vestingStartTime
			err := s.App.BridgeKeeper.SetParams(ctx, params)
			s.Require().NoError(err)

			// Precreate account
			if tc.precreateAccount != nil {
				s.App.AccountKeeper.NewAccount(ctx, tc.precreateAccount)
				s.App.AccountKeeper.SetAccount(ctx, tc.precreateAccount)

				// Confirm creation
				addr := tc.precreateAccount.GetAddress()
				s.Require().True(s.App.AccountKeeper.GetAccount(ctx, addr).GetAddress().Equals(addr))
			}

			// Get sequencer account
			accAddress, err := s.App.BridgeKeeper.GenerateSequencerAccountFromEthereumDeposit(
				ctx, tc.args.ethAddress, tc.args.vestingDuration, tc.args.totalCoins,
			)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			// Simulate minting of tokens to account
			if !tc.fundAccount.Empty() {
				err = s.App.MintKeeper.MintCoins(ctx, tc.fundAccount)
				s.Require().NoError(err)
				err = s.App.BankKeeper.SendCoinsFromModuleToAccount(
					ctx, minttypes.ModuleName, accAddress, tc.fundAccount,
				)
				s.Require().NoError(err)
			}

			account := s.App.AccountKeeper.GetAccount(ctx, accAddress)
			s.Require().True(tc.isAccountAsExpected(account))

			spendableCoins := s.App.BankKeeper.SpendableCoins(ctx, accAddress)
			s.Require().True(
				tc.expectSpendableCoins.Equal(spendableCoins),
				fmt.Sprintf("%s != %s", tc.expectSpendableCoins, spendableCoins),
			)
		})
	}
}
