package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateSequencerAccountForEthereumAddress is an export of generateSequencerAccountForEthereumAddress for testing.
func (k Keeper) GenerateSequencerAccountForEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountForEthereumAddress(ctx, ethAddress, vestingDuration, totalCoins)
}
