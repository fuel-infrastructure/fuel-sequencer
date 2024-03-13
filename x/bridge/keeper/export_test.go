package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetSequencerAccountForEthereumAddress is an export of getSequencerAccountForEthereumAddress for testing.
func (k Keeper) GetSequencerAccountForEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.getSequencerAccountForEthereumAddress(ctx, ethAddress, vestingDuration, totalCoins)
}
