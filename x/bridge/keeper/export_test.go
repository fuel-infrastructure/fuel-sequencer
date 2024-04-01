package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// GenerateSequencerAccountFromEthereumDeposit is an export of generateSequencerAccountFromEthereumDeposit for testing.
func (k Keeper) GenerateSequencerAccountFromEthereumDeposit(
	ctx sdk.Context, ethAddress string, vestingDuration time.Duration, totalCoins sdk.Coins,
) (sdk.AccAddress, error) {
	return k.generateSequencerAccountFromEthereumDeposit(ctx, ethAddress, vestingDuration, totalCoins)
}

// AuthenticateTx is an export of authenticateTx for testing.
func (k Keeper) AuthenticateTx(sender string, msgs []sdk.Msg, bridgeParams *types.Params) error {
	return k.authenticateTx(sender, msgs, bridgeParams)
}

// ExecuteMsg is an export of ExecuteMsg for testing.
func (k Keeper) ExecuteMsg(ctx sdk.Context, msg sdk.Msg) error {
	return k.executeMsg(ctx, msg)
}

// ProcessAuthorizeEvent is an export of ProcessAuthorizeEvent for testing.
func (k Keeper) ProcessAuthorizeEvent(
	ctx sdk.Context, event *sidecartypes.AuthorizeEvent, bridgeParams *types.Params,
) error {
	return k.processAuthorizeEvent(ctx, event, bridgeParams)
}

func (k Keeper) ProcessSendToSequencerEvent(
	ctx sdk.Context,
	sendEvent *sidecartypes.SendToSequencerEvent,
	params types.Params,
	supplyDeltaInfo *types.SupplyDeltaInfo,
) {
	k.processSendToSequencerEvent(ctx, sendEvent, params, supplyDeltaInfo)
}

// SetRouter is a testing utility which takes the existing keeper, sets its MsgServiceRouter and returns the modified
// keeper
func SetRouter(k Keeper, router *baseapp.MsgServiceRouter) Keeper {
	k.router = router
	return k
}
