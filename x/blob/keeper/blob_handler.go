package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// ProcessBlobMetadata processes a blob metadata transaction
func (k Keeper) ProcessBlobMetadata(ctx sdk.Context, msg *types.MsgBlobMetadataTx) error {
	// Verify blob data is available (either in transient store or via blob store service)
	var blobData []byte
	var err error

	// First check transient store
	if k.HasBlobData(ctx, msg.BlobHash) {
		blobData, err = k.GetBlobData(ctx, msg.BlobHash)
		if err != nil {
			return errors.Wrap(err, "failed to get blob data from transient store")
		}
	} else if k.blobStoreService != nil {
		// Check blob store service
		blobData, err = k.blobStoreService.Get(ctx, msg.BlobHash)
		if err != nil {
			return errors.Wrap(types.ErrBlobNotFound, "blob not available in store service")
		}

		// Store in transient store for this block's execution
		k.SetBlobData(ctx, msg.BlobHash, blobData)
	} else {
		return errors.Wrap(types.ErrBlobNotFound, "no blob data source available")
	}

	// Create metadata from the message
	metadata := types.BlobMetadata{
		Hash:      msg.BlobHash,
		Size_:     msg.GetSize_(),
		Topic:     msg.Topic,
		Nonce:     msg.Nonce,
		Timestamp: ctx.BlockTime(),
	}

	// Store metadata persistently
	if err := k.SetBlobMetadata(ctx, msg.BlobHash, metadata); err != nil {
		return errors.Wrap(err, "failed to store blob metadata")
	}

	return nil
}

// ValidateBlobAvailability checks if a blob is available for processing
func (k Keeper) ValidateBlobAvailability(ctx sdk.Context, blobHash []byte) error {
	// Check transient store first
	if k.HasBlobData(ctx, blobHash) {
		return nil
	}

	// Check blob store service
	if k.blobStoreService != nil && k.blobStoreService.Has(ctx, blobHash) {
		return nil
	}

	return errors.Wrap(types.ErrBlobNotFound, "blob not available")
}

// StoreBlobForBlock stores blob data in transient store for block execution
func (k Keeper) StoreBlobForBlock(ctx sdk.Context, blobHash []byte, data []byte) error {
	// Validate that the hash matches the data
	// Note: In a real implementation, you'd want to verify the hash

	// Store in transient store
	k.SetBlobData(ctx, blobHash, data)

	k.Logger().Debug("stored blob for block execution",
		"hash", string(blobHash),
		"size", len(data))

	return nil
}

// GetBlobFromAnySource attempts to retrieve blob data from any available source
func (k Keeper) GetBlobFromAnySource(ctx sdk.Context, blobHash []byte) ([]byte, error) {
	// First try transient store
	if k.HasBlobData(ctx, blobHash) {
		return k.GetBlobData(ctx, blobHash)
	}

	// Then try blob store service
	if k.blobStoreService != nil {
		data, err := k.blobStoreService.Get(ctx, blobHash)
		if err == nil {
			// Cache in transient store for this block
			k.SetBlobData(ctx, blobHash, data)
			return data, nil
		}
	}

	return nil, errors.Wrap(types.ErrBlobNotFound, "blob not found in any source")
}

// SetBlobMetadata stores blob metadata in the persistent store
func (k Keeper) SetBlobMetadata(ctx sdk.Context, blobHash []byte, metadata types.BlobMetadata) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BlobMetadataKeyPrefix(blobHash)

	bz, err := k.cdc.Marshal(&metadata)
	if err != nil {
		return errors.Wrap(err, "failed to marshal blob metadata")
	}

	return store.Set(key, bz)
}

// GetBlobMetadata retrieves blob metadata from the persistent store
func (k Keeper) GetBlobMetadata(ctx sdk.Context, blobHash []byte) (types.BlobMetadata, error) {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BlobMetadataKeyPrefix(blobHash)

	bz, err := store.Get(key)
	if err != nil {
		return types.BlobMetadata{}, err
	}

	if bz == nil {
		return types.BlobMetadata{}, errors.Wrap(types.ErrBlobNotFound, string(blobHash))
	}

	var metadata types.BlobMetadata
	if err := k.cdc.Unmarshal(bz, &metadata); err != nil {
		return types.BlobMetadata{}, errors.Wrap(err, "failed to unmarshal blob metadata")
	}

	return metadata, nil
}

// HasBlobMetadata checks if blob metadata exists in the persistent store
func (k Keeper) HasBlobMetadata(ctx sdk.Context, blobHash []byte) bool {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BlobMetadataKeyPrefix(blobHash)

	has, err := store.Has(key)
	if err != nil {
		return false
	}

	return has
}

// DeleteBlobMetadata removes blob metadata from the persistent store
func (k Keeper) DeleteBlobMetadata(ctx sdk.Context, blobHash []byte) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.BlobMetadataKeyPrefix(blobHash)

	return store.Delete(key)
}

// GetBlobData retrieves blob data from the transient store
func (k Keeper) GetBlobData(ctx sdk.Context, blobHash []byte) ([]byte, error) {
	tStore := ctx.TransientStore(k.tStoreKey)

	data := tStore.Get(blobHash)
	if data == nil {
		return nil, errors.Wrap(types.ErrBlobNotFound, string(blobHash))
	}

	return data, nil
}

// HasBlobData checks if blob data exists in the transient store
func (k Keeper) HasBlobData(ctx sdk.Context, blobHash []byte) bool {
	tStore := ctx.TransientStore(k.tStoreKey)
	return tStore.Has(blobHash)
}

// SetBlobData stores blob data in the transient store (for block execution)
func (k Keeper) SetBlobData(ctx sdk.Context, blobHash []byte, data []byte) {
	tStore := ctx.TransientStore(k.tStoreKey)
	tStore.Set(blobHash, data)
}
