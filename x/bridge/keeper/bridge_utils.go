package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BurnCoinsFromAddress first sends the coins from the given address to the bridge module and then burns the coins.
func (k Keeper) BurnCoinsFromAddress(ctx sdk.Context, address sdk.AccAddress, amt sdk.Coins) error {

	// Burn the coins from the module account.
	err := k.bankKeeper.BurnCoins(ctx, address, amt)
	if err != nil {
		return errorsmod.Wrapf(err, "cannot burn tokens from address %s", address.String())
	}

	return nil
}

// GetAllBlockedBech32Addresses retrieves all the blocked addresses in bech32 form.
// This list is made up of validator and module addresses from three sources:
// 1. The keeper's list of blockedAddresses, which by default mimics the bank module's block list.
// 2. The additional blocked addresses parameter of the bridge module.
// 3. The operator address of the full list of validators.
func (k Keeper) GetAllBlockedBech32Addresses(ctx sdk.Context) (map[string]bool, error) {

	blocked := make(map[string]bool)

	// Retrieve blocked addresses from keeper.
	for blockedAddr := range k.blockedAddresses {
		blocked[blockedAddr] = true
	}

	// Retrieve blocked addresses from params.
	for _, blockedAddr := range k.GetParams(ctx).AdditionalBlockedAddresses {
		blocked[blockedAddr] = true
	}

	// Attempt to retrieve all the bonded validators.
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

		blocked[sdk.AccAddress(valAddr.Bytes()).String()] = true
	}

	// Block Authority
	blocked[k.GetAuthority()] = true

	return blocked, nil
}

// IsAddressBlocked checks if an address (bech32 or hex) is a blocked address.
func (k Keeper) IsAddressBlocked(ctx sdk.Context, address string) (bool, error) {

	blockedBech32Addresses, err := k.GetAllBlockedBech32Addresses(ctx)
	if err != nil {
		return false, err
	}

	// Since blocked addresses are bech32, we should try converting the address to bech32 just in case it's hex.
	addressBz, err := k.GetAddressCodec().StringToBytes(address)
	if err != nil {
		return false, err
	}

	return blockedBech32Addresses[sdk.AccAddress(addressBz).String()], nil
}
