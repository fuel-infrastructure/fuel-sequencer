package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // ensure bech32 configs are set
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestEthOwnedContinuousVestingAccountSetSequenceErrors(t *testing.T) {
	acc := types.NewEthOwnedContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{}, "")
	require.ErrorContains(t, acc.SetSequence(1), "cannot set sequence number for eth owned continuous vesting account")
	require.ErrorContains(t, acc.SetSequence(2), "cannot set sequence number for eth owned continuous vesting account")
}

func TestEthOwnedContinuousVestingAccountSetPubkeyErrors(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()

	acc := types.NewEthOwnedContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{}, "")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned continuous vesting account")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned continuous vesting account")
}

func TestEthOwnedContinuousVestingAccount_TrackDelegationAndTrackUndelegation(t *testing.T) {

	// Helper coins.
	oneToken := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1))
	tenTokens := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 10))
	originalVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	halfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	moreThanHalfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 101))

	// Helper times.
	months6 := (time.Hour * 24 * 365) / 2
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t0Plus6Months := t0.Add(months6)

	// Helper accounts and addresses.
	owner := testutiltypes.TestSeqAddr1Str
	baseAcc := &authtypes.BaseAccount{
		Address:       owner,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}

	testCases := []struct {
		name                  string
		delegatedFreeBefore   sdk.Coins
		blockTime             time.Time
		balanceAtDelegation   sdk.Coins
		delegationAmount      sdk.Coins
		expLockedCoinsBefore  sdk.Coins
		expDelegatedFreeAfter sdk.Coins
		expPanic              bool
		expPanicSpendable     sdk.Coins // spendable value that shows up in the panic message
	}{
		{
			name:                 "start of vesting; cannot even delegate 1 token; panic",
			blockTime:            t0,              // at start of vesting
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     oneToken,        // try to delegate just one token
			expLockedCoinsBefore: originalVesting, // all are still vesting
			expPanic:             true,
			expPanicSpendable:    nil, // all are still vesting
		},
		{
			name:                  "half way through vesting; delegate half original vesting; successful",
			blockTime:             t0Plus6Months,
			balanceAtDelegation:   originalVesting, // no balance apart from original vesting
			delegationAmount:      halfVesting,     // delegate half of the vesting
			expLockedCoinsBefore:  halfVesting,     // half are still vesting
			expDelegatedFreeAfter: halfVesting,     // half get delegated successfully
			expPanic:              false,
		},
		{
			name:                  "half way through vesting with extra tokens available; successful",
			blockTime:             t0Plus6Months,
			balanceAtDelegation:   originalVesting.Add(tenTokens...), // balance has 10 extra tokens
			delegationAmount:      halfVesting.Add(tenTokens...),     // delegate half of the vesting plus 10
			expLockedCoinsBefore:  halfVesting,                       // half are still vesting
			expDelegatedFreeAfter: halfVesting.Add(tenTokens...),     // half plus 10 get delegated successfully
			expPanic:              false,
		},
		{
			name:                 "half way through vesting; delegate more than half original vesting; panic",
			blockTime:            t0Plus6Months,
			balanceAtDelegation:  originalVesting,     // no balance apart from original vesting
			delegationAmount:     moreThanHalfVesting, // delegate more than half of the vesting
			expLockedCoinsBefore: halfVesting,         // half are still vesting
			expPanic:             true,
			expPanicSpendable:    halfVesting, // half are vested
		},
		{
			name:                  "half way through vesting with some tokens already delegated; can delegate less than half; successful",
			blockTime:             t0Plus6Months,
			delegatedFreeBefore:   tenTokens,                         // 10 tokens were delegated before
			balanceAtDelegation:   originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:      halfVesting.Sub(tenTokens...),     // delegate half of the vesting minus 10 tokens
			expLockedCoinsBefore:  halfVesting,                       // half are still vesting
			expDelegatedFreeAfter: halfVesting,                       // (halfVesting - 10) + 10 = halfVesting
			expPanic:              false,
		},
		{
			name:                  "half way through vesting with some tokens already delegated; cannot delegate half original vesting; panic",
			blockTime:             t0Plus6Months,
			delegatedFreeBefore:   tenTokens,                         // 10 tokens were delegated before
			balanceAtDelegation:   originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:      halfVesting,                       // delegate half of the vesting
			expLockedCoinsBefore:  halfVesting,                       // half are still vesting
			expDelegatedFreeAfter: halfVesting.Add(tenTokens...),     // half get delegated successfully
			expPanic:              true,
			expPanicSpendable:     halfVesting.Sub(tenTokens...), // half are vested, minus 10 missing tokens
		},
		{
			name:                  "vesting done; delegate full amount; successful",
			blockTime:             t1,              // vesting done
			balanceAtDelegation:   originalVesting, // no balance apart from original vesting
			delegationAmount:      originalVesting, // delegate all the original vesting
			expLockedCoinsBefore:  nil,             // all tokens vested
			expDelegatedFreeAfter: originalVesting, // all tokens delegated successfully
			expPanic:              false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			vestingAcc := types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      baseAcc,
						OriginalVesting:  originalVesting,
						DelegatedFree:    tc.delegatedFreeBefore,
						DelegatedVesting: nil, // we expect this to never get set
						EndTime:          t1.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			)

			// Sanity check locked and vesting coins, which are always expected to be equal
			require.True(t, vestingAcc.LockedCoins(tc.blockTime).Equal(tc.expLockedCoinsBefore))
			require.True(t, vestingAcc.GetVestingCoins(tc.blockTime).Equal(tc.expLockedCoinsBefore))

			// Delegation
			if tc.expPanic {
				panicValue := fmt.Sprintf("cannot delegate locked coins; max spendable is %s", tc.expPanicSpendable)
				require.PanicsWithValue(t, panicValue, func() {
					vestingAcc.TrackDelegation(tc.blockTime, tc.balanceAtDelegation, tc.delegationAmount)
				})
				return // test is over
			} else {
				vestingAcc.TrackDelegation(tc.blockTime, tc.balanceAtDelegation, tc.delegationAmount)
			}

			// Undelegate the delegated balance to get to the original values
			vestingAcc.TrackUndelegation(tc.delegationAmount)
		})
	}
}

func TestEthOwnedContinuousVestingAccount_TrackDelegation_ZeroDelegationAmountCausesPanic(t *testing.T) {

	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	zeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 0)}
	tenTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10)}
	withZeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10), sdk.NewInt64Coin("token2", 0)}

	vAcc := types.EthOwnedContinuousVestingAccount{}
	require.PanicsWithValue(t, "delegation attempt with zero amount in coins 0token1", func() {
		vAcc.TrackDelegation(t0, tenTokens, zeroTokens)
	})

	vAcc = types.EthOwnedContinuousVestingAccount{}
	require.PanicsWithValue(t, "delegation attempt with zero amount in coins 10token1,0token2", func() {
		vAcc.TrackDelegation(t0, tenTokens, withZeroTokens)
	})
}

func TestEthOwnedContinuousVestingAccount_TrackUndelegation_ZeroUndelegationAmountCausesPanic(t *testing.T) {

	zeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 0)}
	withZeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10), sdk.NewInt64Coin("token2", 0)}

	vAcc := types.EthOwnedContinuousVestingAccount{}
	require.PanicsWithValue(t, "undelegation attempt with zero amount in coins 0token1", func() {
		vAcc.TrackUndelegation(zeroTokens)
	})

	vAcc = types.EthOwnedContinuousVestingAccount{}
	require.PanicsWithValue(t, "undelegation attempt with zero amount in coins 10token1,0token2", func() {
		vAcc.TrackUndelegation(withZeroTokens)
	})
}

func TestEthOwnedContinuousVestingAccount_AddVestingCoins(t *testing.T) {

	// Helper coins.
	coinsAlreadyThere := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	coinsToAdd := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	coins1234 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1234))
	coins5678 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 5678))

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t2, _ := time.Parse(time.DateOnly, "2026-01-01")

	// Helper accounts and addresses.
	seqAddr1BaseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}
	owner := testutiltypes.TestEthAddr1Str

	testCases := []struct {
		name                string
		account             types.EthOwnedAccountI
		vestingStartTime    time.Time
		vestingEndTime      time.Time
		isAccountAsExpected testutil.AccountValidator
		expErrMsg           string
	}{
		{
			name: "add to EthOwnedContinuousVestingAccount just adds coins if start and end time match",
			account: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere,
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
			vestingStartTime: t0, // same start time
			vestingEndTime:   t2, // same end time
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere.Add(coinsToAdd...), // updated
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedContinuousVestingAccount just adds coins even if only start time matches",
			account: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere,
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
			vestingStartTime: t0, // same start time
			vestingEndTime:   t1, // different end time
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere.Add(coinsToAdd...), // updated
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedContinuousVestingAccount converts to multi account if start time is different",
			account: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere,
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
			vestingStartTime: t1, // different start time
			vestingEndTime:   t2, // same end time
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsToAdd, t1.Unix(), t2.Unix()), // new vesting info
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedContinuousVestingAccount retains single account if just end time is different",
			account: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere,
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
			vestingStartTime: t0, // same start time
			vestingEndTime:   t1, // different end time
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere.Add(coinsToAdd...), // updated
						DelegatedFree:    coins1234,
						DelegatedVesting: coins5678,
						EndTime:          t2.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			acc, err := tc.account.AddVestingCoins(coinsToAdd, tc.vestingStartTime, tc.vestingEndTime)
			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.True(t, tc.isAccountAsExpected(acc))
		})
	}
}
