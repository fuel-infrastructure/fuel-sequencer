package keeper

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// NullEthereumAddress is the null Ethereum address (0x0000000000000000000000000000000000000000).
// If this is specified as a deposit Recipient, the Recipient is considered to be owned by the Sender.
var NullEthereumAddress = new(common.Address).String()

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

// isRecipientOwnedByDepositor determines whether the DepositEvent.Depositor owns DepositEvent.Recipient on the
// Sequencer. An account is owned by the Depositor iff Recipient is a null address, or Recipient is equivalent to
// Depositor (both Ethereum addresses), or Recipient is equivalent to the mapping of Depositor as a Sequencer address.
func isRecipientOwnedByDepositor(depositor, recipient, depositorSeq string, seqMappingErr error) bool {
	return recipient == NullEthereumAddress || recipient == depositor || (seqMappingErr == nil && recipient == depositorSeq)
}

// GenerateSequencerAddressFromEthereumAddress uses the App address codec to generate a Sequencer address from an
// Ethereum one. It uses the underlying StringToBytes which trims the 0x prefix from an Ethereum address, if any, and
// decodes the Ethereum address into bytes.
func (k Keeper) GenerateSequencerAddressFromEthereumAddress(ethAddress string) (sdk.AccAddress, error) {
	// TODO: We might want to verify checksum of address
	if !common.IsHexAddress(ethAddress) {
		return nil, errorsmod.Wrapf(types.ErrInvalidEthAddress, "invalid Ethereum address format (%s)", ethAddress)
	}

	return k.GetAddressCodec().StringToBytes(ethAddress)
}

// GenerateEthereumAddressFromSequencerAddress tries to parse the specified address into a Sequencer AccAddress or
// ValAddress and then parses the resultant bytes into an Ethereum address.
func (k Keeper) GenerateEthereumAddressFromSequencerAddress(seqAddress string) (address common.Address, err error) {

	// Try to parse as account address
	accAddress, err1 := sdk.AccAddressFromBech32(seqAddress)
	if err1 != nil {
		// Try to parse as validator address
		valAddress, err2 := sdk.ValAddressFromBech32(seqAddress)
		if err2 != nil {
			return common.Address{}, fmt.Errorf("could not parse address into a Sequencer address: %s", err1)
		}
		address = common.BytesToAddress(valAddress)
	} else {
		address = common.BytesToAddress(accAddress)
	}
	return address, nil
}

// generateSequencerAccountFromEthereumDeposit gets or creates a Sequencer account for the specified Ethereum address.
// The resultant address is a deterministic 1-1 mapping from the Ethereum address, and the account is guaranteed
// to follow the specified vestingDuration, unless an EthOwnedContinuousVestingAccount exists already, in which
// case the newly specified vestingDuration will be ignored and new coins will assume the existing vesting schedule.
func (k Keeper) generateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, coins sdk.Coins,
) (sdk.AccAddress, error) {

	accAddress, err := k.GenerateSequencerAddressFromEthereumAddress(ethAddress)
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
