package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/fuel-infrastructure/fuel-sequencer/app/apptesting" // ensure bech32 configs are set
	testutiltypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestEthOwnedMultiContinuousVestingAccountSetSequenceErrors(t *testing.T) {
	acc := types.NewEthOwnedMultiContinuousVestingAccount(nil, nil, "")
	require.ErrorContains(t, acc.SetSequence(1), "cannot set sequence number for eth owned multi continuous vesting account")
	require.ErrorContains(t, acc.SetSequence(2), "cannot set sequence number for eth owned multi continuous vesting account")
}

func TestEthOwnedMultiContinuousVestingAccountSetPubkeyErrors(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()

	acc := types.NewEthOwnedMultiContinuousVestingAccount(nil, nil, "")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned multi continuous vesting account")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned multi continuous vesting account")
}

// The tests here make use of vesting infos with matching vesting schedules, for simplicity.
func TestEthOwnedMultiContinuousVestingAccount_TrackDelegationAndTrackUndelegation_Simple(t *testing.T) {

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
		name                 string
		blockTime            time.Time
		vestingInfos         []*types.VestingInfo
		balanceAtDelegation  sdk.Coins
		delegationAmount     sdk.Coins
		expLockedCoinsBefore sdk.Coins
		expPanic             bool
		expPanicSpendable    sdk.Coins // spendable value that shows up in the panic message
	}{
		{
			name:      "start of vesting; cannot even delegate 1 token; panic",
			blockTime: t0, // at start of vesting
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     oneToken,        // try to delegate just one token
			expLockedCoinsBefore: originalVesting, // all are still vesting
			expPanic:             true,
			expPanicSpendable:    nil, // all are still vesting
		},
		{
			name:      "half way through vesting; delegate half original vesting; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     halfVesting,     // delegate half of the vesting
			expLockedCoinsBefore: halfVesting,     // half are still vesting
			expPanic:             false,
		},
		{
			name:      "half way through vesting with extra tokens available; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting.Add(tenTokens...), // balance has 10 extra tokens
			delegationAmount:     halfVesting.Add(tenTokens...),     // delegate half of the vesting plus 10
			expLockedCoinsBefore: halfVesting,                       // half are still vesting
			expPanic:             false,
		},
		{
			name:      "half way through vesting; delegate more than half original vesting; panic",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting,     // no balance apart from original vesting
			delegationAmount:     moreThanHalfVesting, // delegate more than half of the vesting
			expLockedCoinsBefore: halfVesting,         // half are still vesting
			expPanic:             true,
			expPanicSpendable:    halfVesting, // half are vested
		},
		{
			name:      "half way through vesting with some tokens already delegated; can delegate less than half; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:     halfVesting.Sub(tenTokens...),     // delegate half of the vesting minus 10 tokens
			expLockedCoinsBefore: halfVesting,                       // half are still vesting
			expPanic:             false,
		},
		{
			name:      "half way through vesting with some tokens already delegated; cannot delegate half original vesting; panic",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:     halfVesting,                       // delegate half of the vesting
			expLockedCoinsBefore: halfVesting,                       // half are still vesting
			expPanic:             true,
			expPanicSpendable:    halfVesting.Sub(tenTokens...), // half are vested, minus 10 missing tokens
		},
		{
			name:      "vesting done; delegate full amount; successful",
			blockTime: t1, // vesting done
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
			},
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     originalVesting, // delegate all the original vesting
			expLockedCoinsBefore: nil,             // all tokens vested
			expPanic:             false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(baseAcc, tc.vestingInfos, owner)

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

// The tests here make use of vesting infos with differing vesting schedules, for a touch of complexity.
func TestEthOwnedMultiContinuousVestingAccount_TrackDelegationAndTrackUndelegation_Complex(t *testing.T) {

	// Helper coins.
	oneToken := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1))
	tenTokens := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 10))
	originalVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	halfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	vesting1Quarter := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))
	vesting3Quarters := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 150))
	moreThanHalfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 101))

	// Helper times.
	months6 := (time.Hour * 24 * 365) / 2
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t0Plus6Months := t0.Add(months6)
	t1Plus6Months := t1.Add(months6)

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
		vestingInfos          []*types.VestingInfo
		balanceAtDelegation   sdk.Coins
		delegationAmount      sdk.Coins
		expLockedCoinsBefore  sdk.Coins
		expDelegatedFreeAfter sdk.Coins
		expPanic              bool
		expPanicSpendable     sdk.Coins // spendable value that shows up in the panic message
	}{
		{
			name:      "start of vesting; cannot even delegate 1 token; panic",
			blockTime: t0, // at start of vesting
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     oneToken,        // try to delegate just one token
			expLockedCoinsBefore: originalVesting, // all are still vesting
			expPanic:             true,
			expPanicSpendable:    nil, // all are still vesting
		},
		{
			name:      "half way through vesting; delegate half original vesting; unsuccessful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:  originalVesting,  // no balance apart from original vesting
			delegationAmount:     halfVesting,      // delegate half of the vesting
			expLockedCoinsBefore: vesting3Quarters, // three quarters are still vesting
			expPanic:             true,
			expPanicSpendable:    vesting1Quarter, // quarter is vested
		},
		{
			name:      "half way through vesting; delegate quarter original vesting; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:   originalVesting,  // no balance apart from original vesting
			delegationAmount:      vesting1Quarter,  // delegate quarter of the vesting
			expLockedCoinsBefore:  vesting3Quarters, // three quarters are still vesting
			expDelegatedFreeAfter: vesting1Quarter,  // quarter get delegated successfully
			expPanic:              false,
		},
		{
			name:      "half way through vesting with extra tokens available; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:   originalVesting.Add(tenTokens...), // balance has 10 extra tokens
			delegationAmount:      vesting1Quarter,                   // delegate quarter of the vesting + 10 tokens
			expLockedCoinsBefore:  vesting3Quarters,                  // three quarters are still vesting
			expDelegatedFreeAfter: vesting1Quarter,                   // quarter + 10 delegated successfully
			expPanic:              false,
		},
		{
			name:      "half way through vesting; delegate more than half original vesting; panic",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:  originalVesting,     // no balance apart from original vesting
			delegationAmount:     moreThanHalfVesting, // delegate more than half of the vesting
			expLockedCoinsBefore: vesting3Quarters,    // three quarters are still vesting
			expPanic:             true,
			expPanicSpendable:    vesting1Quarter, // quarter are vested
		},
		{
			name:      "half way through vesting with some tokens already delegated; can delegate less than a quarter; successful",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			delegatedFreeBefore:   tenTokens,                         // 10 tokens were delegated before
			balanceAtDelegation:   originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:      vesting1Quarter.Sub(tenTokens...), // delegate quarter of the vesting minus 10 tokens
			expLockedCoinsBefore:  vesting3Quarters,                  // three quarters are still vesting
			expDelegatedFreeAfter: vesting1Quarter,                   // (vesting1Quarter - 10) + 10 = vesting1Quarter
			expPanic:              false,
		},
		{
			name:      "half way through vesting with some tokens already delegated; cannot delegate quarter original vesting; panic",
			blockTime: t0Plus6Months,
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			delegatedFreeBefore:  tenTokens,                         // 10 tokens were delegated before
			balanceAtDelegation:  originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:     vesting1Quarter,                   // delegate quarter of the vesting
			expLockedCoinsBefore: vesting3Quarters,                  // three quarters are still vesting
			expPanic:             true,
			expPanicSpendable:    vesting1Quarter.Sub(tenTokens...), // quarter are vested, minus 10 missing tokens
		},
		{
			name:      "vesting done for schedule 1; delegate full amount; unsuccessful",
			blockTime: t1, // vesting done for schedule 1
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:  originalVesting, // no balance apart from original vesting
			delegationAmount:     originalVesting, // delegate all the original vesting
			expLockedCoinsBefore: vesting1Quarter, // quarter is still vesting
			expPanic:             true,
			expPanicSpendable:    vesting3Quarters, // three quarters are vested
		},
		{
			name:      "vesting done; delegate full amount; successful",
			blockTime: t1Plus6Months, // vesting done
			vestingInfos: []*types.VestingInfo{
				types.NewVestingInfo(halfVesting, t0.Unix(), t1.Unix()),
				types.NewVestingInfo(halfVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix()),
			},
			balanceAtDelegation:   originalVesting, // no balance apart from original vesting
			delegationAmount:      originalVesting, // delegate all the original vesting
			expLockedCoinsBefore:  nil,             // all tokens vested
			expDelegatedFreeAfter: originalVesting, // all tokens delegated successfully
			expPanic:              false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(baseAcc, tc.vestingInfos, owner)

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

func TestEthOwnedMultiContinuousVestingAccount_TrackDelegation_ZeroDelegationAmountCausesPanic(t *testing.T) {

	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	zeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 0)}
	tenTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10)}
	withZeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10), sdk.NewInt64Coin("token2", 0)}

	vAcc := types.EthOwnedMultiContinuousVestingAccount{}
	require.PanicsWithValue(t, "delegation attempt with zero amount in coins 0token1", func() {
		vAcc.TrackDelegation(t0, tenTokens, zeroTokens)
	})

	vAcc = types.EthOwnedMultiContinuousVestingAccount{}
	require.PanicsWithValue(t, "delegation attempt with zero amount in coins 10token1,0token2", func() {
		vAcc.TrackDelegation(t0, tenTokens, withZeroTokens)
	})
}

func TestEthOwnedMultiContinuousVestingAccount_TrackUndelegation_ZeroUndelegationAmountCausesPanic(t *testing.T) {

	zeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 0)}
	withZeroTokens := sdk.Coins{sdk.NewInt64Coin("token1", 10), sdk.NewInt64Coin("token2", 0)}

	vAcc := types.EthOwnedMultiContinuousVestingAccount{}
	require.PanicsWithValue(t, "undelegation attempt with zero amount in coins 0token1", func() {
		vAcc.TrackUndelegation(zeroTokens)
	})

	vAcc = types.EthOwnedMultiContinuousVestingAccount{}
	require.PanicsWithValue(t, "undelegation attempt with zero amount in coins 10token1,0token2", func() {
		vAcc.TrackUndelegation(withZeroTokens)
	})
}

func TestEthOwnedMultiContinuousVestingAccount_AddVestingCoins(t *testing.T) {

	// Helper coins.
	coinsAlreadyThere := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	coinsToAdd := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t2, _ := time.Parse(time.DateOnly, "2026-01-01")
	t3, _ := time.Parse(time.DateOnly, "2027-01-01")

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
			name: "add to EthOwnedMultiContinuousVestingAccount adds to existing vesting accounts if at least one " +
				"has a matching vesting schedule with the same start time",
			account: types.NewEthOwnedMultiContinuousVestingAccount(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
				},
				owner,
			),
			vestingStartTime: t1,
			vestingEndTime:   t2,
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsAlreadyThere.Add(coinsToAdd...), t1.Unix(), t2.Unix()), // updated
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedMultiContinuousVestingAccount adds another vesting account if start time " +
				"does not match any of the existing vesting accounts",
			account: types.NewEthOwnedMultiContinuousVestingAccount(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t3.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t3.Unix()),
				},
				owner,
			),
			vestingStartTime: t2, // mismatch
			vestingEndTime:   t3,
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t3.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t3.Unix()),
					types.NewVestingInfo(coinsToAdd, t2.Unix(), t3.Unix()),
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedMultiContinuousVestingAccount accumulates to existing schedule with same start time " +
				"regardless of end time difference",
			account: types.NewEthOwnedMultiContinuousVestingAccount(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t1.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
				},
				owner,
			),
			vestingStartTime: t0, // same start time as first existing schedule
			vestingEndTime:   t3, // different end time
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere.Add(coinsToAdd...), t0.Unix(), t1.Unix()), // accumulated to first schedule
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
				},
				owner,
			),
		},
		{
			name: "DoS prevention: multiple additions with same start time only create one schedule per unique start time",
			account: types.NewEthOwnedMultiContinuousVestingAccount(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t1.Unix()),
				},
				owner,
			),
			vestingStartTime: t0, // same start time as existing schedule
			vestingEndTime:   t2, // different end time
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere.Add(coinsToAdd...), t0.Unix(), t1.Unix()), // accumulated to existing schedule
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

func TestEthOwnedMultiContinuousVestingAccount_GetVestedCoins(t *testing.T) {

	// Helper accounts and addresses.
	baseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}
	owner := testutiltypes.TestEthAddr1Str

	// Helper times
	months6 := (time.Hour * 24 * 365) / 2
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t2, _ := time.Parse(time.DateOnly, "2026-01-01")
	t0Plus6Months := t0.Add(months6)
	blockTime := t0Plus6Months

	amount := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	halfAmount := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))
	threeQuarters := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 75))

	// Start off with 1 vesting info with 50 vested
	vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(
		baseAcc,
		[]*types.VestingInfo{
			types.NewVestingInfo(amount, t0.Unix(), t1.Unix()),
		},
		owner,
	)
	require.True(t, vestingAcc.GetVestedCoins(blockTime).Equal(halfAmount)) // 50 vested

	// Add another vesting info with 25 more vested
	newVestingInfo := types.NewVestingInfo(amount, t0.Unix(), t2.Unix())
	vestingAcc.Infos = append(vestingAcc.Infos, newVestingInfo)
	require.True(t, vestingAcc.GetVestedCoins(blockTime).Equal(threeQuarters)) // 75 vested

	// Add another vesting info with none vested
	newVestingInfo = types.NewVestingInfo(amount, t0Plus6Months.Unix(), t2.Unix())
	vestingAcc.Infos = append(vestingAcc.Infos, newVestingInfo)
	require.True(t, vestingAcc.GetVestedCoins(blockTime).Equal(threeQuarters)) // still 75 vested
}

func TestEthOwnedMultiContinuousVestingAccount_GetVestingCoins(t *testing.T) {

	// Helper accounts and addresses.
	baseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}
	owner := testutiltypes.TestEthAddr1Str

	// Helper times
	months6 := (time.Hour * 24 * 365) / 2
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t2, _ := time.Parse(time.DateOnly, "2026-01-01")
	t0Plus6Months := t0.Add(months6)
	blockTime := t0Plus6Months

	tokens225 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 225))
	tokens125 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 125))
	tokens100 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	tokens50 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))
	tokens0 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 0))

	// Start off with 1 vesting info with 0 vesting
	vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(
		baseAcc,
		[]*types.VestingInfo{
			types.NewVestingInfo(tokens100, t0.Unix(), t0Plus6Months.Unix()),
		},
		owner,
	)
	require.True(t, vestingAcc.GetVestingCoins(blockTime).Equal(tokens0)) // 0 vesting

	// Add another vesting info with 50 vesting
	newVestingInfo := types.NewVestingInfo(tokens100, t0.Unix(), t1.Unix())
	vestingAcc.Infos = append(vestingAcc.Infos, newVestingInfo)
	require.True(t, vestingAcc.GetVestingCoins(blockTime).Equal(tokens50)) // 50

	// Add another vesting info with 75 more vesting
	newVestingInfo = types.NewVestingInfo(tokens100, t0.Unix(), t2.Unix())
	vestingAcc.Infos = append(vestingAcc.Infos, newVestingInfo)
	require.True(t, vestingAcc.GetVestingCoins(blockTime).Equal(tokens125)) // 125 vesting

	// Add another vesting info with 100 more vesting
	newVestingInfo = types.NewVestingInfo(tokens100, t0Plus6Months.Unix(), t2.Unix())
	vestingAcc.Infos = append(vestingAcc.Infos, newVestingInfo)
	require.True(t, vestingAcc.GetVestingCoins(blockTime).Equal(tokens225)) // 225 vesting
}

func TestEthOwnedMultiContinuousVestingAccount_GetOriginalVesting(t *testing.T) {

	// Helper accounts and addresses.
	baseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}
	owner := testutiltypes.TestEthAddr1Str

	amount1 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 10))
	amount2 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 20))
	amount3 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 50))
	total := amount1.Add(amount2...).Add(amount3...)

	vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(
		baseAcc,
		[]*types.VestingInfo{
			types.NewVestingInfo(amount1, 0, 0),
			types.NewVestingInfo(amount2, 0, 0),
			types.NewVestingInfo(amount3, 0, 0),
		},
		owner,
	)

	require.True(t, vestingAcc.GetOriginalVesting().Equal(total))
}
