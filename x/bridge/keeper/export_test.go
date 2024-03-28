package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateSequencerAccountFromEthereumDeposit is an export of generateSequencerAccountFromEthereumDeposit for testing.
func (k Keeper) GenerateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountFromEthereumDeposit(ctx, ethAddress, vestingDuration, totalCoins)
}
