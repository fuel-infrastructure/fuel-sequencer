package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SupplyDelta(
	goCtx context.Context, msg *types.MsgSupplyDelta,
) (resp *types.MsgSupplyDeltaResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.TryExecSpecialMessage(ctx, msg.Authority, func(ctx sdk.Context) error {
		var innerErr error
		if resp, innerErr = k.supplyDelta(ctx); innerErr != nil {
			resp = &types.MsgSupplyDeltaResponse{} // normalise
			return innerErr
		}
		return nil
	})

	return resp, err
}

func (k msgServer) supplyDelta(ctx sdk.Context) (*types.MsgSupplyDeltaResponse, error) {

	bridgeParams := k.GetParams(ctx)

	// Confirm that MsgSupplyDelta was injected at the correct height.
	blockHeight := ctx.BlockHeight()
	if !bridgeParams.IsMsgSupplyDeltaBlock(uint64(blockHeight)) {
		return nil, types.ErrUnexpectedOperation.Wrapf(
			"MsgSupplyDelta not expected at height %d", blockHeight,
		)
	}

	// Increment LastEthereumNonce and get the result so that it is added to MsgSupplyDeltaResponse.
	nonce := k.MustGetNextEthereumNonceAndIncrement(ctx)

	// Get the supply delta to be reported from state
	supplyDeltaInfo := k.MustGetSupplyDeltaInfo(ctx)
	supplyDelta := supplyDeltaInfo.ToReport

	// Reset the toReport value
	supplyDeltaInfo.ToReport = sdkmath.ZeroInt()
	k.SetSupplyDeltaInfo(ctx, supplyDeltaInfo)

	// Emit event
	err := ctx.EventManager().EmitTypedEvent(&types.EventSupplyDeltaReported{SupplyDelta: supplyDelta, Nonce: nonce})
	if err != nil {
		return nil, err
	}

	return &types.MsgSupplyDeltaResponse{
		Nonce:       nonce,
		SupplyDelta: supplyDelta,
	}, nil
}
