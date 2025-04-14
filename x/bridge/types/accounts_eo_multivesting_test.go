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

func TestEthOwnedMultiContinuousVestingAccount_TrackDelegationAndTrackUndelegation_(t *testing.T) {

	// Helper coins.
	oneToken := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1))
	//tenTokens := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 10))
	originalVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	doubleVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 400)) // 2x originalVesting
	//halfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	//moreThanHalfVesting := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 101))

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

	info1 := types.NewVestingInfo(originalVesting, t0.Unix(), t1.Unix())
	info2 := types.NewVestingInfo(originalVesting, t0Plus6Months.Unix(), t1Plus6Months.Unix())
	vAcc := types.NewEthOwnedMultiContinuousVestingAccount(
		baseAcc,
		[]*types.VestingInfo{info1, info2},
		owner,
	)

	require.Panics(t, func() {
		vAcc.TrackDelegation(t0, doubleVesting, oneToken)
	})
}

func TestEthOwnedMultiContinuousVestingAccount_TrackDelegationAndTrackUndelegation(t *testing.T) {

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
			delegatedFreeBefore:   nil,                               // TODO: tenTokens,                         // 10 tokens were delegated before
			balanceAtDelegation:   originalVesting.Sub(tenTokens...), // balance is missing 10 tokens
			delegationAmount:      halfVesting.Sub(tenTokens...),     // delegate half of the vesting minus 10 tokens
			expLockedCoinsBefore:  halfVesting,                       // half are still vesting
			expDelegatedFreeAfter: halfVesting,                       // (halfVesting - 10) + 10 = halfVesting
			expPanic:              false,
		},
		{
			name:                  "half way through vesting with some tokens already delegated; cannot delegate half original vesting; panic",
			blockTime:             t0Plus6Months,
			delegatedFreeBefore:   nil,                               // TODO: tenTokens,                         // 10 tokens were delegated before
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

			vestingAcc := types.NewEthOwnedMultiContinuousVestingAccount(
				baseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(originalVesting, t0.Unix(), t1.Unix()),
					types.NewVestingInfo(originalVesting, t0.Unix(), t1.Unix()),
				},
				owner,
			)
			vestingAcc.DelegatedFree = tc.delegatedFreeBefore

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

			// Check delegation fields after delegation
			require.True(t, vestingAcc.GetDelegatedFree().Equal(tc.expDelegatedFreeAfter))
			require.True(t, vestingAcc.GetDelegatedVesting().IsZero()) // we expect this to never get set

			// Undelegate the delegated balance to get to the original values
			vestingAcc.TrackUndelegation(tc.delegationAmount)

			// Check delegation fields after undelegation
			require.True(t, vestingAcc.GetDelegatedFree().Equal(tc.delegatedFreeBefore)) // back to original DelegatedFree
			require.True(t, vestingAcc.GetDelegatedVesting().IsZero())                   // we expect this to never get set
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
	coins1234 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1234))

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
			name: "add to EthOwnedMultiContinuousVestingAccount adds to existing vesting accounts if at least one " +
				"has a matching vesting schedule",
			account: types.NewEthOwnedMultiContinuousVestingAccountWithDelegation(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
				},
				coins1234, // DelegatedFree - will be retained
				nil,       // DelegatedVesting - will be retained
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
				coins1234,
				nil,
				owner,
			),
		},
		{
			name: "add to EthOwnedMultiContinuousVestingAccount adds another vesting account if vesting schedule " +
				"does not match any of the vesting accounts",
			account: types.NewEthOwnedMultiContinuousVestingAccountWithDelegation(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
				},
				coins1234,
				nil,
				owner,
			),
			vestingStartTime: t0,
			vestingEndTime:   t1,
			isAccountAsExpected: testutil.MatchesEthOwnedMultiContinuousVestingAccRaw(
				seqAddr1BaseAcc,
				[]*types.VestingInfo{
					types.NewVestingInfo(coinsAlreadyThere, t0.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsAlreadyThere, t1.Unix(), t2.Unix()),
					types.NewVestingInfo(coinsToAdd, t0.Unix(), t1.Unix()),
				},
				coins1234,
				nil,
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
