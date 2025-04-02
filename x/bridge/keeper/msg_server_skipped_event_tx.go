package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SkippedEventTx(goCtx context.Context, msg *types.MsgSkippedEventTx) (resp *types.MsgSkippedEventTxResponse, err error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.TryExecSpecialMessage(ctx, msg.Authority, func(ctx sdk.Context) error {
		var innerErr error
		if resp, innerErr = k.skippedEventTx(ctx, msg); innerErr != nil {
			resp = &types.MsgSkippedEventTxResponse{} // normalise
			return innerErr
		}
		return nil
	})

	return resp, err
}

func (k msgServer) skippedEventTx(ctx sdk.Context, msg *types.MsgSkippedEventTx) (*types.MsgSkippedEventTxResponse, error) {
	// Log if the reason was trimmed
	if len(msg.ReasonForSkip) > types.MaxReasonLength {
		k.Logger().Warn("reason_for_skip exceeded maximum length and was trimmed",
			"original_length", len(msg.ReasonForSkip),
			"max_length", types.MaxReasonLength,
		)
	}

	err := ctx.EventManager().EmitTypedEvent(&types.EventSkippedEventTx{
		ReasonForSkip:  msg.ReasonForSkip,
		EthBlockNumber: msg.EthBlockNumber,
		EthLogIndex:    msg.EthLogIndex,
		EthTxIndex:     msg.EthTxIndex,
		EthTxHash:      msg.EthTxHash,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgSkippedEventTxResponse{}, nil
}
