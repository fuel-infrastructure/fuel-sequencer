package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// normaliseExistingAccount tries to extract EthOwnedAccountI from sdk.AccountI. If an 'EthOwned' account is found,
// the account info is extracted untouched. Otherwise, we use the base account and wrap it as a EthOwnedBaseAccount.
// If the existing account was a non-'EthOwned' vesting account, any vesting details are discarded.
func normaliseExistingAccount(acc sdk.AccountI, ethAddress string) types.EthOwnedAccountI {

	// Try to parse into EthOwnedContinuousVestingAccount.
	if vAcc, ok := acc.(*types.EthOwnedContinuousVestingAccount); ok {
		return vAcc
	}

	// Try to parse into EthOwnedBaseAccount.
	if baseAcc, ok := acc.(*types.EthOwnedBaseAccount); ok {
		return baseAcc
	}

	// Backup: reconstruct details from sdk.AccountI. This will only happen if an unexpected type of account was
	// pre-created. Thus, the vesting details are not really important and any vesting tokens will become available.
	baseAcc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	return types.NewEthOwnedBaseAccount(baseAcc, ethAddress)
}

// generateSequencerAccountFromEthereumDeposit gets or creates a Sequencer account for the specified Ethereum address.
// The resultant address is a deterministic 1-1 mapping from the Ethereum address, and the account is guaranteed
// to follow the specified vestingDuration, unless an EthOwnedContinuousVestingAccount exists already, in which
// case the newly specified vestingDuration will be ignored and new coins will assume the existing vesting schedule.
func (k Keeper) generateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, coins sdk.Coins,
) (sdk.AccAddress, error) {

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddress(ethAddress)
	if err != nil {
		return nil, err
	}

	// Get existing account, if any.
	existingAcc := k.accountKeeper.GetAccount(ctx, accAddress)
	createNewAcc := existingAcc == nil

	// Normalise any existing account into EthOwnedAccountI or create a new one.
	var acc types.EthOwnedAccountI
	if createNewAcc {
		acc = types.NewEthOwnedBaseAccountWithAddress(accAddress, ethAddress)
	} else {
		acc = normaliseExistingAccount(existingAcc, ethAddress)
	}

	// Update account with vesting details, if any.
	if vestingDuration > 0 {
		startTime, endTime, err := k.GetParams(ctx).VestingTimesFromVestingDuration(vestingDuration)
		if err != nil {
			return nil, err
		}

		acc, err = acc.AddVestingCoins(coins, startTime, endTime)
		if err != nil {
			return nil, err
		}
	}

	// Get an account number and account sequence if it's a new account.
	accI := sdk.AccountI(acc)
	if createNewAcc {
		accI = k.accountKeeper.NewAccount(ctx, accI)
	}

	k.accountKeeper.SetAccount(ctx, accI)
	return accAddress, nil
}
