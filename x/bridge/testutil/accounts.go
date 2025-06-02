package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type AccountValidator func(acc sdk.AccountI) bool

// MatchesEthOwnedAcc asserts that the account matches the specified EthOwnedBaseAccount.
func MatchesEthOwnedAcc(expected *types.EthOwnedBaseAccount) AccountValidator {
	return func(acc sdk.AccountI) bool {
		bAcc, ok := acc.(*types.EthOwnedBaseAccount)
		return ok &&
			(bAcc.AccountOwner == expected.AccountOwner) &&
			(bAcc.GetAddress().Equals(expected.GetAddress())) &&
			((bAcc.PubKey == nil && expected.PubKey == nil) || (bAcc.GetPubKey().Equals(expected.GetPubKey()))) &&
			(bAcc.GetAccountNumber() == expected.GetAccountNumber()) &&
			(bAcc.GetSequence() == expected.GetSequence())
	}
}

// MatchesEthOwnedAccRaw is a wrapper for matchesEthOwnedAcc that accepts the raw values of EthOwnedBaseAccount instead
// of EthOwnedBaseAccount directly.
func MatchesEthOwnedAccRaw(ba *authtypes.BaseAccount, owner string) AccountValidator {
	return MatchesEthOwnedAcc(types.NewEthOwnedBaseAccount(ba, owner))
}

// MatchesEthOwnedContinuousVestingAcc asserts that the account matches the specified EthOwnedContinuousVestingAccount.
func MatchesEthOwnedContinuousVestingAcc(expected *types.EthOwnedContinuousVestingAccount) AccountValidator {
	return func(acc sdk.AccountI) bool {
		vAcc, ok := acc.(*types.EthOwnedContinuousVestingAccount)
		return ok &&
			(vAcc.AccountOwner == expected.AccountOwner) &&
			(vAcc.StartTime == expected.StartTime) &&
			(vAcc.OriginalVesting.Equal(expected.OriginalVesting)) &&
			(vAcc.DelegatedFree.Equal(expected.DelegatedFree)) &&
			(vAcc.DelegatedVesting.Equal(expected.DelegatedVesting)) &&
			(vAcc.EndTime == expected.EndTime) &&
			MatchesEthOwnedAcc(expected.ToEthOwnedBaseAccount())(vAcc.ToEthOwnedBaseAccount())
	}
}

// MatchesEthOwnedContinuousVestingAccRaw is a wrapper for matchesEthOwnedContinuousVestingAcc that accepts the raw
// values of EthOwnedContinuousVestingAcc instead of EthOwnedContinuousVestingAcc directly.
func MatchesEthOwnedContinuousVestingAccRaw(cva *vestingtypes.ContinuousVestingAccount, owner string) AccountValidator {
	return MatchesEthOwnedContinuousVestingAcc(types.NewEthOwnedContinuousVestingAccount(cva, owner))
}
