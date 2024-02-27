package keeper

import (
	"errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k Keeper) createCrosschainAccount(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	// TODO: what if we receive multiple deposit requests to the same account with a different vesting duration?
	// TODO: can we safely overwrite an account if it was pre-created by someone by means of transferring tokens to it?

	accAddress, err := types.GenerateCrosschainAccountAddress(ethAddress)
	if err != nil {
		return nil, err
	}

	var vestingEndTime time.Time
	var vestingDone bool
	if vestingDuration > 0 {
		vestingEndTime = k.GetParams(ctx).VestingStartTime.Add(vestingDuration)
		vestingDone = ctx.BlockTime().After(vestingEndTime)
	}

	// If account already exists, use it.
	acc := k.accountKeeper.GetAccount(ctx, accAddress)
	if acc != nil {

		// Vesting done, so we don't need to confirm that the account is a vesting account.
		if vestingDone {
			return accAddress, nil
		}

		// If account is a vesting account, we can confirm that it was set up correctly
		vacc, ok := acc.(banktypes.VestingAccount)
		if ok {

		}
	}

	// Assume account will be a base account.
	baseAcc := authtypes.NewBaseAccountWithAddress(accAddress)
	acc := sdk.AccountI(baseAcc)

	// Create vesting account if vesting not done.
	if !vestingDone {
		acc, err = vestingtypes.NewDelayedVestingAccount(baseAcc, totalCoins, vestingEndTime.Unix())
		if err != nil {
			return nil, errors.New("some error") // TODO: some error
		}
	}

	k.accountKeeper.NewAccount(ctx, acc)
	k.accountKeeper.SetAccount(ctx, acc)

	return accAddress, nil
}
