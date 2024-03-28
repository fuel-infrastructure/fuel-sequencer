package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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

// ExecuteMsg is an export of ExecuteMsg for testing.
func (k Keeper) ExecuteMsg(ctx sdk.Context, msg sdk.Msg) (*codectypes.Any, error) {
	return k.executeMsg(ctx, msg)
}

// SetRouter is a testing utility which takes the existing keeper, sets its MsgServiceRouter and returns the modified
// keeper
func SetRouter(k Keeper, router *baseapp.MsgServiceRouter) Keeper {
	k.router = router
	return k
}
