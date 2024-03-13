package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSequencerAddressForEthereumAddress is an export of getSequencerAddressForEthereumAddress for testing.
func (k Keeper) GetSequencerAddressForEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.getSequencerAddressForEthereumAddress(ctx, ethAddress, vestingDuration, totalCoins)
}
