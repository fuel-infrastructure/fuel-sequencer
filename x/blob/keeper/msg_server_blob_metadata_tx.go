package keeper

import (
	"context"
	"encoding/hex"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// PostBlobMetadata handles MsgBlobMetadataTx
func (k msgServer) PostBlobMetadata(goCtx context.Context, msg *types.MsgBlobMetadataTx) (*types.MsgBlobMetadataTxResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Process the blob metadata using the keeper
	if err := k.Keeper.ProcessBlobMetadata(ctx, msg); err != nil {
		return nil, err
	}

	// Emit events
	hashStr := hex.EncodeToString(msg.Hash)
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeBlobMetadataReceived,
			sdk.NewAttribute(types.AttributeKeyBlobHash, hashStr),
			sdk.NewAttribute(types.AttributeKeyBlobSize, fmt.Sprintf("%d", msg.GetSize_())),
			sdk.NewAttribute(types.AttributeKeyTopic, msg.Topic),
			sdk.NewAttribute(types.AttributeKeyNonce, fmt.Sprintf("%d", msg.Nonce)),
		),
	)

	return &types.MsgBlobMetadataTxResponse{}, nil
}

// ProcessBlobMetadata processes a blob metadata transaction
func (k Keeper) ProcessBlobMetadata(ctx sdk.Context, msg *types.MsgBlobMetadataTx) error {
	// Verify blob data is available (either in blobpool or blobhub)
	// var blob *store.StoredBlob

	// First check the blobpool
	hash := store.Key(msg.Hash)
	if k.blobpool.hasBlob(hash) {
		// blob, err = k.blobpool.getBlob(hash)
		// if err != nil {
		// 	return errors.Wrap(err, "failed to get blob data from blobpool")
		// }

		return nil
	}
	// else if k.blobhubClient.has(hash) {
	// 	panic("not implemented yet - blob metadata should be stored in blobpool")
	// 	return errors.Wrap(types.ErrBlobNotFound, "blob data source available")
	// }

	return nil
}
