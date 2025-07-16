package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

type msgServer struct {
	*Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// PostBlobMetadata handles MsgBlobMetadataTx
func (ms msgServer) PostBlobMetadata(goCtx context.Context, msg *types.MsgBlobMetadataTx) (*types.MsgBlobMetadataTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Process the blob metadata using the keeper
	if err := ms.Keeper.ProcessBlobMetadata(ctx, msg); err != nil {
		return nil, err
	}

	// Emit events
	hashStr := hex.EncodeToString(msg.BlobHash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlobPosted,
			sdk.NewAttribute(types.AttributeKeyBlobHash, hashStr),
			sdk.NewAttribute(types.AttributeKeyBlobSize, fmt.Sprintf("%d", msg.GetSize_())),
			sdk.NewAttribute(types.AttributeKeyTopic, msg.Topic),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprintf("%d", msg.Nonce)),
		),
	)

	return &types.MsgBlobMetadataTxResponse{}, nil
}
