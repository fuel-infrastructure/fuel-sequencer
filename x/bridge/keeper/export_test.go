package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateSequencerAccountFromEthereumAddress is an export of generateSequencerAccountFromEthereumAddress for testing.
func (k Keeper) GenerateSequencerAccountFromEthereumAddress(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountFromEthereumAddress(ctx, ethAddress, vestingDuration, totalCoins)
}

// AuthenticateTx is an export of authenticateTx for testing.
func (k Keeper) AuthenticateTx(ctx sdk.Context, sender string, msgs []sdk.Msg) error {
	return k.authenticateTx(ctx, sender, msgs)
}
