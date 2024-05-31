package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateSequencerAccountFromEthereumDeposit is an export of generateSequencerAccountFromEthereumDeposit for testing.
func (k Keeper) GenerateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountFromEthereumDeposit(ctx, ethAddress, vestingDuration, totalCoins)
}

// SetRouter is a testing utility which takes the existing keeper, sets its MsgServiceRouter and returns the modified
// keeper
func SetRouter(k Keeper, router *baseapp.MsgServiceRouter) Keeper {
	k.router = router
	return k
}
