package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SkipEventTx(goCtx context.Context, msg *types.MsgSkippedEventTx) (resp *types.MsgSkippedEventTxResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.TryExecSpecialMessage(ctx, msg.Authority, func(ctx sdk.Context) error {
		var innerErr error
		if resp, innerErr = k.skipEventTx(ctx, msg); innerErr != nil {
			resp = &types.MsgSkippedEventTxResponse{} // normalise
			return innerErr
		}
		return nil
	})

	return resp, err
}

func (k msgServer) skipEventTx(ctx sdk.Context, msg *types.MsgSkippedEventTx) (*types.MsgSkippedEventTxResponse, error) {

	return nil, nil
}
