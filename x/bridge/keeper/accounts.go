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

// baseAccFromAcc extracts the base account from a sdk.AccountI.
// The account must be a base account or a ContinuousVestingAccount.
func baseAccFromAcc(acc sdk.AccountI) (*authtypes.BaseAccount, error) {
	if vacc, ok := acc.(*vestingtypes.ContinuousVestingAccount); ok {
		return vacc.BaseVestingAccount.BaseAccount, nil
	}

	if baseAcc, ok := acc.(*authtypes.BaseAccount); ok {
		return baseAcc, nil
	}

	return nil, types.ErrUnexpectedAccountType.Wrapf("could not extract base acc from %s", acc.GetAddress().String())
}

// getSequencerAddressForEthereumAddress gets or creates a Sequencer account for the specified Ethereum address.
// The resultant address is a deterministic 1-1 mapping from the Ethereum address, and the account is guaranteed
// to follow the specified vestingDuration, regardless of whether the account already existed in other forms.
func (k Keeper) getSequencerAddressForEthereumAddress(
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
		vestingDone = ctx.BlockTime().After(vestingEndTime)
	}

	var baseAcc *authtypes.BaseAccount
	var newAcc bool

	// If account already exists, use it, otherwise create one.
	// We also extract the base account since we'll most likely use it.
	acc := k.accountKeeper.GetAccount(ctx, accAddress)
	if acc == nil {
		newAcc = true
		baseAcc = authtypes.NewBaseAccountWithAddress(accAddress)
		acc = sdk.AccountI(baseAcc)
	} else {
		newAcc = false
		baseAcc, err = baseAccFromAcc(acc)
		if err != nil {
			return nil, err
		}
	}

	// If vesting done, ensure we use the base account. Otherwise, create a vesting account.
	if vestingDone {
		acc = baseAcc
	} else {
		// Calculate the new total balance for the vesting account.
		existingBalance := k.bankKeeper.GetAllBalances(ctx, accAddress)
		newBalance := existingBalance.Add(totalCoins...)

		acc, err = vestingtypes.NewContinuousVestingAccount(
			baseAcc, newBalance, vestingStartTime.Unix(), vestingEndTime.Unix(),
		)
		if err != nil {
			return nil, err
		}
	}

	// Get an account number if it's a new account.
	if newAcc {
		k.accountKeeper.NewAccount(ctx, acc)
	}

	k.accountKeeper.SetAccount(ctx, acc)
	return accAddress, nil
}
