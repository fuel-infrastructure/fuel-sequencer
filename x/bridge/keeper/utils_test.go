package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type accountValidator func(acc sdk.AccountI) bool

// matchesEthOwnedAcc asserts that the account is an EthOwnedAccount and matches the supplied account.
func matchesEthOwnedAcc(baseAcc *types.EthOwnedAccount) accountValidator {
	return func(acc sdk.AccountI) bool {
		bAcc, ok := acc.(*types.EthOwnedAccount)
		return ok &&
			(bAcc.GetAddress().Equals(baseAcc.GetAddress())) &&
			((bAcc.PubKey == nil && baseAcc.PubKey == nil) || (bAcc.GetPubKey().Equals(baseAcc.GetPubKey()))) &&
			(bAcc.GetAccountNumber() == baseAcc.GetAccountNumber()) &&
			(bAcc.GetSequence() == baseAcc.GetSequence())
	}
}

// matchesEthOwnedContinuousVestingAcc asserts that the account is an EthOwnedContinuousVestingAccount and matches the supplied account.
func matchesEthOwnedContinuousVestingAcc(vestingAcc *types.EthOwnedContinuousVestingAccount) accountValidator {
	return func(acc sdk.AccountI) bool {
		vAcc, ok := acc.(*types.EthOwnedContinuousVestingAccount)
		return ok &&
			(vAcc.StartTime == vestingAcc.StartTime) &&
			(vAcc.OriginalVesting.Equal(vestingAcc.OriginalVesting)) &&
			(vAcc.DelegatedFree.Equal(vestingAcc.DelegatedFree)) &&
			(vAcc.DelegatedVesting.Equal(vestingAcc.DelegatedVesting)) &&
			(vAcc.EndTime == vestingAcc.EndTime) &&
			matchesEthOwnedAcc(vestingAcc.ToEthOwnedAccount())(vAcc.ToEthOwnedAccount())
	}
}
