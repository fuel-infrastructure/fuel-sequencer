package types_test

import (
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

func TestEthOwnedBaseAccountSetSequenceErrors(t *testing.T) {
	baseAcc := authtypes.NewBaseAccountWithAddress(testutiltypes.TestSeqAddr1)
	acc := types.NewEthOwnedBaseAccount(baseAcc, testutiltypes.TestEthAddr1Str)
	require.ErrorContains(t, acc.SetSequence(1), "cannot set sequence number for eth owned account")
	require.ErrorContains(t, acc.SetSequence(2), "cannot set sequence number for eth owned account")
}

func TestEthOwnedBaseAccountSetPubkeyErrors(t *testing.T) {
	_, pk, _ := testdata.KeyTestPubAddr()

	acc := types.NewEthOwnedBaseAccount(&authtypes.BaseAccount{}, "")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned account")
	require.ErrorContains(t, acc.SetPubKey(pk), "cannot set public key for eth owned account")
}

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

func TestAddVestingCoins(t *testing.T) {

	// Helper coins.
	coinsAlreadyThere := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 200))
	coinsToAdd := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 100))
	coins1234 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 1234))
	coins5678 := sdk.NewCoins(sdk.NewInt64Coin(testutiltypes.TestToken, 5678))

	// Helper times.
	t0, _ := time.Parse(time.DateOnly, "2024-01-01")
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t100, _ := time.Parse(time.DateOnly, "2100-01-01")
	t101, _ := time.Parse(time.DateOnly, "2101-01-01")

	// Helper accounts and addresses.
	seqAddr1BaseAcc := &authtypes.BaseAccount{
		Address:       testutiltypes.TestSeqAddr1Str,
		AccountNumber: uint64(1),
		Sequence:      uint64(2),
	}
	owner := testutiltypes.TestEthAddr1Str
	ethOwnedBaseAcc := types.NewEthOwnedBaseAccount(seqAddr1BaseAcc, owner)

	testCases := []struct {
		name                string
		account             types.EthOwnedAccountI
		vestingStartTime    time.Time
		vestingEndTime      time.Time
		isAccountAsExpected testutil.AccountValidator
		expErrMsg           string
	}{
		{
			name:             "add to EthOwnedBaseAccount converts it to a EthOwnedContinuousVestingAccount",
			account:          ethOwnedBaseAcc,
			vestingStartTime: t0,
			vestingEndTime:   t1,
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsToAdd,
						DelegatedFree:    nil,
						DelegatedVesting: nil,
						EndTime:          t1.Unix(),
					},
					StartTime: t0.Unix(),
				},
				owner,
			),
		},
		{
			name: "add to EthOwnedContinuousVestingAccount disregards new vesting details and adds coins",
			account: types.NewEthOwnedContinuousVestingAccount(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere,
						DelegatedFree:    coins1234, // will be retained
						DelegatedVesting: coins5678, // will be retained
						EndTime:          t1.Unix(), // will be retained
					},
					StartTime: t0.Unix(), // will be retained
				},
				owner,
			),
			vestingStartTime: t100, // disregarded
			vestingEndTime:   t101, // disregarded
			isAccountAsExpected: testutil.MatchesEthOwnedContinuousVestingAccRaw(
				&vestingtypes.ContinuousVestingAccount{
					BaseVestingAccount: &vestingtypes.BaseVestingAccount{
						BaseAccount:      seqAddr1BaseAcc,
						OriginalVesting:  coinsAlreadyThere.Add(coinsToAdd...), // updated
						DelegatedFree:    coins1234,                            // unchanged
						DelegatedVesting: coins5678,                            // unchanged
						EndTime:          t1.Unix(),                            // unchanged
					},
					StartTime: t0.Unix(), // unchanged
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
