package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// vestingStartTimeDelay is a constant period of time during which tokens are completely locked.
var vestingStartTimeDelay = time.Hour * 24 * 365

// blankEthOwnedContinuousVestingAccount returns a blank vesting account wrapped in EthOwnedContinuousVestingAccount.
// We only need the blanked-out amounts from this struct, so we don't care about the other fields.
func blankEthOwnedContinuousVestingAccount() *types.EthOwnedContinuousVestingAccount {
	return types.NewEthOwnedContinuousVestingAccount(&vestingtypes.ContinuousVestingAccount{
		BaseVestingAccount: &vestingtypes.BaseVestingAccount{
			OriginalVesting:  nil,
			DelegatedFree:    nil,
			DelegatedVesting: nil,
		},
	})
}

// detailsFromFromExistingAcc tries to extract EthOwnedAccount and EthOwnedContinuousVestingAccount from sdk.AccountI.
// If a recognized 'EthOwned' account type is found, the account info is extracted untouched.
// In this case we also know that this is NOT the first deposit coming from Ethereum.
// Otherwise, we extract just the base account and discard any existing vesting information.
//
// Practically speaking, the last case will happen if someone pre-creates an account.
// In this case any locked tokens in this pre-created account will become liquid.
func detailsFromFromExistingAcc(acc sdk.AccountI) (
	_ *types.EthOwnedAccount,
	_ *types.EthOwnedContinuousVestingAccount,
	isFirstDepositFromEthereum bool,
) {

	// Try to parse into EthOwnedContinuousVestingAccount.
	if vacc, ok := acc.(*types.EthOwnedContinuousVestingAccount); ok {
		return vacc.ToEthOwnedAccount(), vacc, false
	}

	// Try to parse into EthOwnedAccount.
	if baseAcc, ok := acc.(*types.EthOwnedAccount); ok {
		return baseAcc, blankEthOwnedContinuousVestingAccount(), false
	}

	// Backup: reconstruct details from sdk.AccountI. This will only happen if an unexpected type of account was
	// pre-created. Thus, the vesting details are not really important and any vesting tokens will become available.
	baseAcc := authtypes.NewBaseAccount(acc.GetAddress(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
	return types.NewEthOwnedAccount(baseAcc), blankEthOwnedContinuousVestingAccount(), true
}

// depositFromEthereum generates the Sequencer address corresponding to the Ethereum address that is sending the tokens.
// Like the CreateVestingAccount function in the Cosmos SDK, we first create the account and then send tokens to it.
// Ref: https://github.com/cosmos/cosmos-sdk/blob/v0.50.4/x/auth/vesting/msg_server.go#L31
// Note: this is just a scaffold function for now and should be revised before it is used, or otherwise scrapped.
func (k Keeper) depositFromEthereum(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) error {

	address, err := k.generateSequencerAccountFromEthereumAddress(ctx, ethAddress, vestingDuration, totalCoins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, totalCoins)
	if err != nil {
		return err
	}

	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, address, totalCoins)
	if err != nil {
		return err
	}

	// TODO: consider making assertions about changes in the spendable balance to sanity check our calculations.

	return nil
}

// generateSequencerAccountFromEthereumAddress gets or creates a Sequencer account for the specified Ethereum address.
// The resultant address is a deterministic 1-1 mapping from the Ethereum address, and the account is guaranteed
// to follow the specified vestingDuration, regardless of whether the account already existed in other forms.
func (k Keeper) generateSequencerAccountFromEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {

	accAddress, err := types.GenerateSequencerAddressFromEthereumAddress(ethAddress)
	if err != nil {
		return nil, err
	}

	var acc *types.EthOwnedAccount
	var vestingAcc *types.EthOwnedContinuousVestingAccount // we're only interested in the token amounts
	var firstDepositFromEthereum bool

	// If account already exists, use it, otherwise create one.
	// We also extract the base account since we'll most likely use it.
	existingAcc := k.accountKeeper.GetAccount(ctx, accAddress)
	createNewAcc := existingAcc == nil
	if createNewAcc {
		acc = types.NewEthOwnedAccountWithAddress(accAddress)
		vestingAcc = blankEthOwnedContinuousVestingAccount()
		firstDepositFromEthereum = true
	} else {
		acc, vestingAcc, firstDepositFromEthereum = detailsFromFromExistingAcc(existingAcc)
	}

	// Calculate vesting details.
	var vestingStartTime time.Time
	var vestingEndTime time.Time
	vestingDone := true
	if vestingDuration > 0 {
		if firstDepositFromEthereum {
			// The vesting start time delay is included in the vesting duration, so
			// if the vesting duration is smaller, it cannot be considered as valid.
			// We consider vestingDuration == vestingStartTimeDelay as invalid as well.
			if vestingDuration <= vestingStartTimeDelay {
				return nil, types.ErrInvalidVestingDuration.Wrapf(
					"must be greater than vesting start time delay, got %s <= %s",
					vestingDuration, vestingStartTimeDelay,
				)
			}
			params := k.GetParams(ctx)
			vestingStartTime = params.VestingStartTime.Add(vestingStartTimeDelay)
			vestingEndTime = params.VestingStartTime.Add(vestingDuration)
		} else {
			// Reuse existing vesting schedule if the account has already been interacted with once.
			vestingStartTime = time.Unix(vestingAcc.StartTime, 0)
			vestingEndTime = time.Unix(vestingAcc.EndTime, 0)
		}
		vestingDone = ctx.BlockTime().Compare(vestingEndTime) >= 0
	}

	// If vesting done, ensure we use the base account. Otherwise, create a vesting account or reuse the existing one.
	var newAcc sdk.AccountI
	if vestingDone {
		newAcc = acc
	} else if firstDepositFromEthereum {
		// Create new vesting account.
		newVestingAcc, err := vestingtypes.NewContinuousVestingAccount(
			acc.BaseAccount, totalCoins, vestingStartTime.Unix(), vestingEndTime.Unix(),
		)
		if err != nil {
			return nil, err
		}
		newAcc = types.NewEthOwnedContinuousVestingAccount(newVestingAcc)
	} else {
		// Reuse existing vesting account.
		vestingAcc := vestingtypes.NewContinuousVestingAccountRaw(
			&vestingtypes.BaseVestingAccount{
				BaseAccount:      acc.BaseAccount,
				OriginalVesting:  vestingAcc.OriginalVesting.Add(totalCoins...),
				DelegatedFree:    vestingAcc.DelegatedFree,
				DelegatedVesting: vestingAcc.DelegatedVesting,
				EndTime:          vestingEndTime.Unix(),
			},
			vestingStartTime.Unix(),
		)
		if err = vestingAcc.Validate(); err != nil {
			return nil, err
		}
		newAcc = types.NewEthOwnedContinuousVestingAccount(vestingAcc)
	}

	// Get an account number if it's a new account.
	// This assigns a new account sequence.
	if createNewAcc {
		newAcc = k.accountKeeper.NewAccount(ctx, newAcc)
	}

	k.accountKeeper.SetAccount(ctx, newAcc)
	return accAddress, nil
}
