package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// vestingStartTimeDelay is a constant period of time in which tokens are completely locked.
var vestingStartTimeDelay = time.Hour * 24 * 365

// blankBaseVestingAccount returns a blank base vesting account. We only need
// the amounts from this struct, so we don't care about the other fields.
func blankBaseVestingAccount() *vestingtypes.BaseVestingAccount {
	return &vestingtypes.BaseVestingAccount{
		OriginalVesting:  nil,
		DelegatedFree:    nil,
		DelegatedVesting: nil,
	}
}

// detailsFromFromExistingAcc extracts the base account and base vesting account, if any, from a sdk.AccountI.
// If a recognized account type is found, the base account and base vesting account are extracted untouched.
// Otherwise, we extract just the base account and discard any existing vesting information.
//
// Practically speaking, the last case will happen if someone pre-creates an unrecognized vesting account type.
// In this case any locked tokens in this pre-created vesting account will become liquid.
func detailsFromFromExistingAcc(acc sdk.AccountI) (*authtypes.BaseAccount, *vestingtypes.BaseVestingAccount) {

	// Try to parse into ContinuousVestingAccount.
	if vacc, ok := acc.(*vestingtypes.ContinuousVestingAccount); ok {
		return vacc.BaseVestingAccount.BaseAccount, vacc.BaseVestingAccount
	}

	// Try to parse into BaseAccount.
	if baseAcc, ok := acc.(*authtypes.BaseAccount); ok {
		return baseAcc, blankBaseVestingAccount()
	}

	// Backup: reconstruct details from sdk.AccountI. This will only happen if an unexpected type of vesting account
	// was pre-created. Thus, the vesting details are not really important and any vesting tokens will become available.
	baseAcc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	return baseAcc, blankBaseVestingAccount()
}

// generateSequencerAccountForEthereumAddress gets or creates a Sequencer account for the specified Ethereum address.
// The resultant address is a deterministic 1-1 mapping from the Ethereum address, and the account is guaranteed
// to follow the specified vestingDuration, regardless of whether the account already existed in other forms.
func (k Keeper) generateSequencerAccountForEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {

	accAddress, err := types.GenerateSequencerAddressForEthereumAddress(ethAddress)
	if err != nil {
		return nil, err
	}

	// Calculate vesting details.
	var vestingStartTime time.Time
	var vestingEndTime time.Time
	vestingDone := true
	if vestingDuration > 0 {
		vestingStartTime = k.GetParams(ctx).VestingStartTime.Add(vestingStartTimeDelay)
		vestingEndTime = vestingStartTime.Add(vestingDuration)
		vestingDone = ctx.BlockTime().Compare(vestingEndTime) >= 0 // block time is at or after vesting end time
	}

	var baseAcc *authtypes.BaseAccount
	var baseVestingAcc *vestingtypes.BaseVestingAccount // we're only interested in the amounts

	// If account already exists, use it, otherwise create one.
	// We also extract the base account since we'll most likely use it.
	acc := k.accountKeeper.GetAccount(ctx, accAddress)
	createNewAcc := acc == nil
	if acc == nil {
		baseAcc = authtypes.NewBaseAccountWithAddress(accAddress)
		baseVestingAcc = blankBaseVestingAccount()
		acc = sdk.AccountI(baseAcc)
	} else {
		baseAcc, baseVestingAcc = detailsFromFromExistingAcc(acc)
	}

	// If vesting done, ensure we use the base account. Otherwise, create a vesting account.
	if vestingDone {
		acc = baseAcc
	} else {
		vestingAcc := vestingtypes.NewContinuousVestingAccountRaw(
			&vestingtypes.BaseVestingAccount{
				BaseAccount:      baseAcc,
				OriginalVesting:  baseVestingAcc.OriginalVesting.Add(totalCoins...),
				DelegatedFree:    baseVestingAcc.DelegatedFree,
				DelegatedVesting: baseVestingAcc.DelegatedVesting,
				EndTime:          vestingEndTime.Unix(),
			},
			vestingStartTime.Unix(),
		)
		if err = vestingAcc.Validate(); err != nil {
			return nil, err
		}

		acc = vestingAcc
	}

	// Get an account number if it's a new account.
	// This assigns a new account sequence.
	if createNewAcc {
		k.accountKeeper.NewAccount(ctx, acc)
	}

	k.accountKeeper.SetAccount(ctx, acc)
	return accAddress, nil
}
