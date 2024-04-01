package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// GenerateSequencerAccountFromEthereumDeposit is an export of generateSequencerAccountFromEthereumDeposit for testing.
func (k Keeper) GenerateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountFromEthereumDeposit(ctx, ethAddress, vestingDuration, totalCoins)
}

// AuthenticateTx is an export of authenticateTx for testing.
func (k Keeper) AuthenticateTx(sender string, msgs []sdk.Msg, messagesAllowed []string) error {
	return k.authenticateTx(sender, msgs, messagesAllowed)
}

// ExecuteMsg is an export of ExecuteMsg for testing.
func (k Keeper) ExecuteMsg(ctx sdk.Context, msg sdk.Msg) (*codectypes.Any, error) {
	return k.executeMsg(ctx, msg)
}

// ProcessAuthorizeEvent is an export of ProcessAuthorizeEvent for testing.
func (k Keeper) ProcessAuthorizeEvent(
	ctx sdk.Context, event *sidecartypes.AuthorizeEvent, messagesAllowed []string,
) error {
	return k.processAuthorizeEvent(ctx, event, messagesAllowed)
}

// SetRouter is a testing utility which takes the existing keeper, sets its MsgServiceRouter and returns the modified
// keeper
func SetRouter(k Keeper, router *baseapp.MsgServiceRouter) Keeper {
	k.router = router
	return k
}
