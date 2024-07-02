package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// BurnCoinsFromAddress first sends the coins from the given address to the bridge module and then burns the coins.
func (k Keeper) BurnCoinsFromAddress(ctx sdk.Context, address sdk.AccAddress, amt sdk.Coins) error {
	// Send coins from address to bridge module.
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, address, bridgetypes.ModuleName, amt)
	if err != nil {
		return errorsmod.Wrapf(err, "cannot send tokens from %s to %s module", address.String(), bridgetypes.ModuleName)
	}

	// Burn the coins from the module account.
	err = k.bankKeeper.BurnCoins(ctx, bridgetypes.ModuleName, amt)
	if err != nil {
		return errorsmod.Wrapf(err, "cannot burn tokens from %s module", bridgetypes.ModuleName)
	}

	return nil
}

// GetAllBlockedAddresses retrieves all the blocked addresses made up of validator and module addresses.
func (k Keeper) GetAllBlockedAddresses(
	ctx sdk.Context,
	paramsBlockedAddresses []string,
) (map[string]bool, error) {

	// Attempt to retrieve blocked addressed from params and set them as blocked.
	for _, paramsBlockedAddr := range paramsBlockedAddresses {
		k.blockedAddresses[paramsBlockedAddr] = true
	}

	// Attempt to retrieve all the validators.
	validators, err := k.stakingKeeper.GetLastValidators(ctx)
	if err != nil {
		return nil, err
	}

	// Block all the validator addresses.
	for _, validator := range validators {
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return nil, err
		}

		k.blockedAddresses[sdk.AccAddress(valAddr.Bytes()).String()] = true
	}

	// Block Authority
	k.blockedAddresses[k.GetAuthority()] = true

	return k.blockedAddresses, nil
}
