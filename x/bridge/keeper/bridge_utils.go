package keeper

import (
	errorsmod "cosmossdk.io/errors"
	evidencetypes "cosmossdk.io/x/evidence/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencertypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
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
func (k Keeper) GetAllBlockedAddresses(ctx sdk.Context, paramsBlockedAddresses []string) (map[string]bool, error) {
	blockedAddresses := make(map[string]bool)

	// Attemp to retrieve blocked addressed from params and set them as blocked.
	for _, paramsBlockedAddr := range paramsBlockedAddresses {
		blockedAddresses[paramsBlockedAddr] = true
	}

	// Attempt to retrieve all the validators.
	validators, err := k.stakingKeeper.GetAllValidators(ctx)
	if err != nil {
		return nil, err
	}

	// Block all the validator addresses
	for _, validator := range validators {
		blockedAddresses[validator.GetOperator()] = true
	}

	// NOTE: These need to be up to date with the list of modules registered in app/app_config.go
	modulesToBlock := []string{
		authtypes.ModuleName,
		authtypes.FeeCollectorName,
		vestingtypes.ModuleName,
		banktypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		"tx",
		genutiltypes.ModuleName,
		authz.ModuleName,
		upgradetypes.ModuleName,
		distrtypes.ModuleName,
		evidencetypes.ModuleName,
		minttypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		consensustypes.ModuleName,
		bridgetypes.ModuleName,
		sequencertypes.ModuleName,
	}

	// Block all of the above module addresses
	for _, moduleName := range modulesToBlock {
		addr := k.accountKeeper.GetModuleAddress(moduleName)
		if addr != nil {
			blockedAddresses[addr.String()] = true
		}
	}

	return blockedAddresses, nil
}
