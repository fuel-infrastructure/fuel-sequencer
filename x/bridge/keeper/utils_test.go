package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

type accountValidator func(acc sdk.AccountI) bool

// matchesBaseAcc asserts that the account is a base account and matches the supplied account.
func matchesBaseAcc(baseAcc *authtypes.BaseAccount) accountValidator {
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
func matchesContinuousVestingAccount(vestingAcc *vestingtypes.ContinuousVestingAccount) accountValidator {
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
