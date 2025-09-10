package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// PostBlobMetadata handles MsgBlobMetadataTx
func (k msgServer) PostBlobMetadata(
	goCtx context.Context, msg *types.MsgBlobMetadataTx,
) (*types.MsgBlobMetadataTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Process the blob metadata using the keeper
	if err := k.ProcessBlobMetadata(ctx, msg); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlobMetadataReceived,
			sdk.NewAttribute(types.AttributeKeyBlobHash, msg.Hash),
			sdk.NewAttribute(types.AttributeKeyBlobSize, fmt.Sprintf("%d", msg.GetSize_())),
			sdk.NewAttribute(types.AttributeKeyTopic, msg.Topic),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprintf("%d", msg.Nonce)),
		),
	)

	return &types.MsgBlobMetadataTxResponse{}, nil
}

// ProcessBlobMetadata processes a blob metadata transaction
func (k Keeper) ProcessBlobMetadata(ctx sdk.Context, msg *types.MsgBlobMetadataTx) error {
	// Notify if blob data is available (either in blobpool or blobhub)
	// var blob *store.StoredBlob

	// First check the blobpool - ValidateBasic handled the error, can skip the check
	hash, _ := store.ParseKey(msg.Hash)
	if !k.Has(hash) {
		ctx.Logger().Debug(
			"blob metadata tx received, without blob in blobpool",
			"blob_hash", hash.String())
	} else {
		ctx.Logger().Debug(
			"blob metadata tx received, blob already in blobpool",
			"blob_hash", hash.String())
	}

	return nil
}
