package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// BurnCoinsFromAddress first sends the coins from the given address to the bridge module and then burns the coins.
func (k Keeper) BurnCoinsFromAddress(ctx sdk.Context, address sdk.AccAddress, amt sdk.Coins) error {
	// Send coins from address to bridge module.
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, address, types.ModuleName, amt)
	if err != nil {
		return errorsmod.Wrapf(err, "cannot send tokens from %s to %s module", address.String(), types.ModuleName)
	}

	// Burn the coins from the module account.
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
	if err != nil {
		return errorsmod.Wrapf(err, "cannot burn tokens from %s module", types.ModuleName)
	}

	return nil
}

// IsAuthorizedMessage returns true if the sdk.Msg TypeURL is present in messagesAllowed, otherwise false
func (k Keeper) IsAuthorizedMessage(messagesAllowed []string, msg sdk.Msg) bool {
	// Check that wildcard * option for allowing all message types is the only string in the array, if so, return true
	if len(messagesAllowed) == 1 && messagesAllowed[0] == types.AllowAllAuthorizeMessages {
		return true
	}

	for _, messageAllowed := range messagesAllowed {
		if messageAllowed == sdk.MsgTypeURL(msg) {
			return true
		}
	}

	return false
}
