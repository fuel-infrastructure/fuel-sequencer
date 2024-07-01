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

// GetAllBlockedAddresses retrieves all the blocked addresses in bech32 form, made up of validator and module addresses.
func (k Keeper) GetAllBlockedAddresses(
	ctx sdk.Context,
	paramsBlockedAddresses []string,
) (map[string]bool, error) {

	allBlockedAddresses := make(map[string]bool)

	// Retrieve blocked addresses from keeper.
	for blockedAddr := range k.blockedAddresses {
		allBlockedAddresses[blockedAddr] = true
	}

	// Retrieve blocked addresses from params.
	for _, blockedAddr := range paramsBlockedAddresses {
		allBlockedAddresses[blockedAddr] = true
	}

	// Attempt to retrieve all the validators.
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	// Block all the validator addresses.
	for _, validator := range validators {
		valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
		if err != nil {
			return nil, err
		}

		allBlockedAddresses[sdk.AccAddress(valAddr.Bytes()).String()] = true
	}

	// Block Authority
	allBlockedAddresses[k.GetAuthority()] = true

	return allBlockedAddresses, nil
}

// IsAddressBlocked checks if an address (bech32 or hex) is a blocked address.
func (k Keeper) IsAddressBlocked(ctx sdk.Context, address string, paramsBlockedAddresses []string) (bool, error) {

	blockedAddresses, err := k.GetAllBlockedAddresses(ctx, paramsBlockedAddresses)
	if err != nil {
		return false, err
	}

	// Since blocked addresses are bech32, we should try converting the address to bech32 just in case it's hex.
	addressBz, err := k.GetAddressCodec().StringToBytes(address)
	if err != nil {
		return false, err
	}

	return blockedAddresses[sdk.AccAddress(addressBz).String()], nil
}
