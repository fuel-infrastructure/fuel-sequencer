package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (k msgServer) SupplyDelta(goCtx context.Context, msg *types.MsgSupplyDelta) (*types.MsgSupplyDeltaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handling the message and construct full response
	_ = ctx

	nonce := k.MustGetLastEthereumNonce(ctx).AddRaw(1)
	k.SetLastEthereumNonce(ctx, nonce)

	return &types.MsgSupplyDeltaResponse{
		Nonce: nonce,
	}, nil
}
