package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

func (k msgServer) PostBlob(goCtx context.Context, msg *types.MsgPostBlob) (*types.MsgPostBlobResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message and construct full response
	_ = ctx

	nonce := k.bridgeKeeper.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.bridgeKeeper.SetLastEthereumNonce(ctx, nonce)

	return &types.MsgPostBlobResponse{
		Nonce: nonce,
		From:  msg.From,
		Topic: msg.Topic,
		Order: msg.Order,
		Data:  msg.Data,
	}, nil
}
