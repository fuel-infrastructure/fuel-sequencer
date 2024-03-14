package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const (
	firstAccNumber   = 7 // this is not zero due to module accounts
	firstAccSequence = 0

	token = "token"
)

type (
	accountValidator func(acc sdk.AccountI) bool
)

func (s *KeeperTestSuite) TestGetSequencerAccountFromEthereumAddress() {

	// ethAddr1Str -> seqAddr1Str
	ethAddr1Str := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	seqAddr1Str := "fuelsequencer13tch2uhman7dhjjphmx9uwx7kvg2kqfj5y56hsmljlv93pgma5vqyks99k"
	seqAddr1 := sdk.MustAccAddressFromBech32(seqAddr1Str)

	// corresponds to seqAddr1Str
	seqAddr1BaseAcc := &authtypes.BaseAccount{
		Address:       seqAddr1Str,
		AccountNumber: firstAccNumber,
		Sequence:      firstAccSequence,
	}

	// Helper times.
	blockTime, _ := time.Parse(time.DateOnly, "2024-01-01")
	oneYear := time.Hour * 24 * 365
	twoYears := oneYear * 2
	blockTimePlusOneYear := blockTime.Add(oneYear) // accounts for vesting start time delay
	someTimeWaaaayInTheFuture, _ := time.Parse(time.DateOnly, "2030-01-01")

	// Helper token amounts.
	token200 := sdk.NewCoins(sdk.NewInt64Coin(token, 200))
	token150 := sdk.NewCoins(sdk.NewInt64Coin(token, 150))
	token100 := sdk.NewCoins(sdk.NewInt64Coin(token, 100))
	token50 := sdk.NewCoins(sdk.NewInt64Coin(token, 50))

	// matchesBaseAcc asserts that the account is a base account and matches the supplied account.
	matchesBaseAcc := func(baseAcc *authtypes.BaseAccount) accountValidator {
		return func(acc sdk.AccountI) bool {
			bAcc, ok := acc.(*authtypes.BaseAccount)
			return ok &&
				(bAcc.GetAddress().Equals(baseAcc.GetAddress())) &&
				((bAcc.PubKey == nil && baseAcc.PubKey == nil) || (bAcc.GetPubKey().Equals(baseAcc.GetPubKey()))) &&
				(bAcc.GetAccountNumber() == baseAcc.GetAccountNumber()) &&
				(bAcc.GetSequence() == baseAcc.GetSequence())
		}
	}

	// matchesContinuousVestingAccount asserts that the account is a vesting account and matches the supplied account.
	matchesContinuousVestingAccount := func(vestingAcc *vestingtypes.ContinuousVestingAccount) accountValidator {
		return func(acc sdk.AccountI) bool {
			vAcc, ok := acc.(*vestingtypes.ContinuousVestingAccount)
			return ok &&
				(vAcc.StartTime == vestingAcc.StartTime) &&
				(vAcc.OriginalVesting.Equal(vestingAcc.OriginalVesting)) &&
				(vAcc.DelegatedFree.Equal(vestingAcc.DelegatedFree)) &&
				(vAcc.DelegatedVesting.Equal(vestingAcc.DelegatedVesting)) &&
				(vAcc.EndTime == vestingAcc.EndTime) &&
				matchesBaseAcc(vestingAcc.BaseAccount)(vAcc.BaseAccount)
		}
	}

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
		isAccountAsExpected  func(sdk.AccountI) bool
		expectSpendableCoins sdk.Coins
		expectErrMsg         string
	}{
		//
		// --------- Test cases with basic invalid values
		//
		{
			name:      "invalid eth address => err",
			blockTime: blockTime,
			args: fnArgs{
				ethAddress: "invalid_eth_address",
			},
			expectErrMsg: "invalid Ethereum address format (invalid_eth_address)",
		},
		{
			name:             "vesting duration < vesting start time delay (1 year) => err",
			blockTime:        blockTime,
			vestingStartTime: blockTime,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: oneYear - 1,
				totalCoins:      token100,
			},
			expectErrMsg: "must be greater than vesting start time delay, got 8759h59m59.999999999s <= 8760h0m0s",
		},
		{
			name:             "vesting duration == vesting start time delay (1 year) => err",
			blockTime:        blockTime,
			vestingStartTime: blockTime,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: oneYear,
				totalCoins:      token100,
			},
			expectErrMsg: "must be greater than vesting start time delay, got 8760h0m0s <= 8760h0m0s",
		},
		//
		// --------- Test cases with no precreated account
		//
		{
			name:        "account with no vesting and no coins => base account",
			blockTime:   blockTime,
			fundAccount: nil,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      nil,
			},
			isAccountAsExpected: matchesBaseAcc(
				authtypes.NewBaseAccount(seqAddr1, nil, firstAccNumber, firstAccSequence),
			),
			expectSpendableCoins: nil,
		},
		{
			name:        "account with no vesting and some coins => base account",
			blockTime:   blockTime,
			fundAccount: token100,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  matchesBaseAcc(seqAddr1BaseAcc),
			expectSpendableCoins: token100,
		},
		{
			// blockTime:              2024
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is before the actual start time, so we expect no tokens to be available.
			name:             "acc with vesting starting in the future => vesting account",
			blockTime:        blockTime,
			vestingStartTime: blockTime,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting duration
				},
			}),
			expectSpendableCoins: nil,
		},
		{
			// blockTime:              2025 + 0.5 year
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens are available.
			name:             "acc with vesting half-way => vesting account",
			blockTime:        blockTimePlusOneYear.Add(oneYear / 2), // half-way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			}),
			expectSpendableCoins: token50, // half of the vesting tokens are available
		},
		{
			// blockTime:              2025 + 1 year
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is at the end of the actual end time, meaning all the tokens are expected to be available.
			name:             "acc with vesting ended => base account",
			blockTime:        blockTimePlusOneYear.Add(oneYear), // all the way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected:  matchesBaseAcc(seqAddr1BaseAcc),
			expectSpendableCoins: token100, // all tokens available
		},
		{
			// blockTime:              2025 + 1 year - 1 ns
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is ALMOST at actual end time, meaning all the tokens are pretty much all available.
			name:             "edge check :: acc with vesting ALMOST ended => vesting account",
			blockTime:        blockTimePlusOneYear.Add(oneYear - 1), // all the way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token100,
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			}),
			expectSpendableCoins: token100, // all tokens available, due to rounding
		},
		//
		// --------- Test cases with overriding of a precreated account
		//
		{
			name:             "account with no vesting and some coins overrides existing base account => base account",
			precreateAccount: seqAddr1BaseAcc,
			blockTime:        blockTime,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  matchesBaseAcc(seqAddr1BaseAcc),
			expectSpendableCoins: token200, // all the 200 tokens are available
		},
		{
			// blockTime:              2024
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is before the actual start time, so we expect no tokens to be available in precreated account.
			// But we're creating an account with no vesting duration, so this gets overwritten with a base account.
			name: "account with no vesting and some coins overrides existing vesting account => base account",
			precreateAccount: &vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			},
			blockTime:        blockTime,
			vestingStartTime: blockTime,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: 0,
				totalCoins:      token100,
			},
			isAccountAsExpected:  matchesBaseAcc(seqAddr1BaseAcc),
			expectSpendableCoins: token200, // all the 200 tokens are available
		},
		{
			// blockTime:              2025 + 0.5 year
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens are available.
			name:             "acc with vesting half-way overrides existing base account => vesting account",
			precreateAccount: seqAddr1BaseAcc,
			blockTime:        blockTimePlusOneYear.Add(oneYear / 2), // half-way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,                                 // only 100 are vesting
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			}),
			expectSpendableCoins: token150, // precreated account's 100 plus half of newly vested tokens
		},
		{
			// blockTime:              2025 + 0.5 year
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens are available.
			name: "acc with vesting half-way builds on top of existing vesting account (ContinuousVestingAccount), " +
				"resulting in updated OriginalVesting and maintained delegation values => vesting account",
			precreateAccount: &vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // accounts for 1 year delay in vesting start time
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:      seqAddr1BaseAcc,
					OriginalVesting:  token100,
					DelegatedFree:    token50,                                  // these should be untouched
					DelegatedVesting: token50,                                  // these should be untouched
					EndTime:          blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			},
			blockTime:        blockTimePlusOneYear.Add(oneYear / 2), // half-way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token200, // fund with 200 due to precreated account
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // precreated account's vesting start time is disregarded
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:      seqAddr1BaseAcc,
					OriginalVesting:  token200,                                 // 100 + 100 are all vesting
					DelegatedFree:    token50,                                  // these were untouched
					DelegatedVesting: token50,                                  // these were untouched
					EndTime:          blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			}),
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
		{
			// blockTime:              2025 + 0.5 year
			// vestingStartTime:       2024
			// actualVestingStartTime: 2025
			// actualVestingEndTime:   2025 + 1 year
			//
			// Block time is half-way between actual start and end time, meaning half of the tokens are available.
			name: "acc with vesting half-way builds on top of existing vesting account (DelayedVestingAccount), " +
				"disregarding any existing vesting schedule => vesting account",
			precreateAccount: &vestingtypes.DelayedVestingAccount{ // Use unexpected vesting account type
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token50,                          // only 50 are vesting initially
					EndTime:         someTimeWaaaayInTheFuture.Unix(), // all tokens are locked
				},
			},
			blockTime:        blockTimePlusOneYear.Add(oneYear / 2), // half-way through vesting duration
			vestingStartTime: blockTime,
			fundAccount:      token150, // fund with 150 due to precreated account
			args: fnArgs{
				ethAddress:      ethAddr1Str,
				vestingDuration: twoYears,
				totalCoins:      token100,
			},
			isAccountAsExpected: matchesContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
				StartTime: blockTimePlusOneYear.Unix(), // precreated account's vesting start time is disregarded
				BaseVestingAccount: &vestingtypes.BaseVestingAccount{
					BaseAccount:     seqAddr1BaseAcc,
					OriginalVesting: token100,                                 // 100 are vesting; the original 50 are liquid
					EndTime:         blockTimePlusOneYear.Add(oneYear).Unix(), // 1 year lock + 1 year vesting
				},
			}),
			expectSpendableCoins: token100, // all of precreated account's 50 plus half of the 100 newly vested tokens
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
			accAddress, err := s.App.BridgeKeeper.GenerateSequencerAccountFromEthereumAddress(
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
