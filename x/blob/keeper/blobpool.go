package keeper

import (
	"sync"

	"cosmossdk.io/log"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// blobpool manages blob storage at the application level
type blobpool struct {
	logger log.Logger

	blobs sync.Map // hash -> blob data
}

// newBlobpool creates a new blob pool
func newBlobpool(logger log.Logger) *blobpool {
	p := &blobpool{
		logger: logger.With("module", "blobpool"),
		blobs:  sync.Map{},
	}

	p.logger.Info("initialized new blobpool")
	return p
}

// hasBlob checks if a blob is available in the pool
func (p *blobpool) hasBlob(hash store.Key) bool {
	_, exists := p.blobs.Load(hash)
	p.logger.Debug("checked blob existence", "hash", hash.String(), "exists", exists)
	return exists
}

// getBlob retrieves a blob from the pool
func (p *blobpool) getBlob(hash store.Key) (*store.StoredBlob, error) {
	aBlob, exists := p.blobs.Load(hash)
	if !exists {
		p.logger.Debug("blob not found", "hash", hash.String())
		return nil, types.ErrBlobNotFound
	}

	blob, ok := aBlob.(*store.StoredBlob)
	if !ok {
		p.logger.Error("invalid blob type in pool", "hash", hash.String(), "type", aBlob)
		return nil, types.ErrBlobNotFound
	}
	p.logger.Debug("retrieved blob", "hash", hash.String(), "size", len(blob.Data))
	return blob, nil
}

// storeBlob stores a blob in the pool
func (p *blobpool) storeBlob(blob *store.StoredBlob) {
	if blob == nil {
		p.logger.Error("attempted to store nil blob")
		return
	}
	p.blobs.Store(blob.Key, blob)
	p.logger.Debug("stored blob", "hash", blob.Key.String(), "size", len(blob.Data))
}
