package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type accountValidator func(acc sdk.AccountI) bool

// matchesEthOwnedAcc asserts that the account matches the specified EthOwnedBaseAccount.
func matchesEthOwnedAcc(expected *types.EthOwnedBaseAccount) accountValidator {
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

// matchesEthOwnedAccRaw is a wrapper for matchesEthOwnedAcc that accepts the raw values of EthOwnedBaseAccount instead
// of EthOwnedBaseAccount directly.
func matchesEthOwnedAccRaw(ba *authtypes.BaseAccount, owner string) accountValidator {
	return matchesEthOwnedAcc(types.NewEthOwnedBaseAccount(ba, owner))
}

// matchesEthOwnedContinuousVestingAcc asserts that the account matches the specified EthOwnedContinuousVestingAccount.
func matchesEthOwnedContinuousVestingAcc(expected *types.EthOwnedContinuousVestingAccount) accountValidator {
	return func(acc sdk.AccountI) bool {
		vAcc, ok := acc.(*types.EthOwnedContinuousVestingAccount)
		return ok &&
			(vAcc.AccountOwner == expected.AccountOwner) &&
			(vAcc.StartTime == expected.StartTime) &&
			(vAcc.OriginalVesting.Equal(expected.OriginalVesting)) &&
			(vAcc.DelegatedFree.Equal(expected.DelegatedFree)) &&
			(vAcc.DelegatedVesting.Equal(expected.DelegatedVesting)) &&
			(vAcc.EndTime == expected.EndTime) &&
			matchesEthOwnedAcc(expected.ToEthOwnedBaseAccount())(vAcc.ToEthOwnedBaseAccount())
	}
}

// matchesEthOwnedContinuousVestingAccRaw is a wrapper for matchesEthOwnedContinuousVestingAcc that accepts the raw
// values of EthOwnedContinuousVestingAcc instead of EthOwnedContinuousVestingAcc directly.
func matchesEthOwnedContinuousVestingAccRaw(cva *vestingtypes.ContinuousVestingAccount, owner string) accountValidator {
	return matchesEthOwnedContinuousVestingAcc(types.NewEthOwnedContinuousVestingAccount(cva, owner))
}
