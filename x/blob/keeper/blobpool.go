package keeper

import (
	"context"
	"sync"
	"time"

	"cosmossdk.io/log"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// Blobpool manages blob storage at the application level
type Blobpool struct {
	logger log.Logger

	blobs sync.Map // hash -> blob data
	stats
}

type stats struct {
	count uint // blobpool count

	// hit/miss ratio calculation
	hits   uint
	misses uint
}

// newBlobpool creates a new blob pool
func newBlobpool(logger log.Logger) *Blobpool {
	p := &Blobpool{
		logger: logger.With("module", "blobpool"),
		blobs:  sync.Map{},
		stats:  stats{},
	}

	p.logger.Info("initialized new blobpool")
	return p
}

func (s *stats) update(ctx context.Context, hit bool) {
	if hit {
		s.hits++
	} else {
		s.misses++
	}

	// Update metrics
	// TODO: limit this to avoid excessive updates?
	metrics.UpdateBlobpoolHitMissRatios(ctx, s.hits, s.misses)
}

// Has checks if a blob is available in the pool
func (p *Blobpool) Has(ctx context.Context, hash store.Key) bool {
	_, exists := p.blobs.Load(hash)

	p.update(ctx, exists) // Update hit/miss statistics

	p.logger.Debug("checked blob existence", "hash", hash.String(), "exists", exists)
	return exists
}

// Get retrieves a blob from the pool
func (p *Blobpool) Get(ctx context.Context, hash store.Key) (*store.StoredBlob, error) {
	start := time.Now()
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

	// Record metrics
	retrievalTime := time.Since(start)
	metrics.ObserveBlobRetrievalTime(ctx, retrievalTime)
	metrics.ObserveBlobSize(ctx, len(blob.Data))

	p.logger.Debug("retrieved blob", "hash", hash.String(), "size", len(blob.Data))
	return blob, nil
}

// Insert stores a blob in the pool
func (p *Blobpool) Insert(ctx context.Context, blob *store.StoredBlob) {
	if blob == nil {
		p.logger.Error("attempted to store nil blob")
		return
	}

	start := time.Now()
	p.blobs.Store(blob.Key, blob)
	storageLatency := time.Since(start)

	// Record metrics
	blobSize := len(blob.Data)
	metrics.ObserveBlobStorageLatency(ctx, storageLatency)
	metrics.ObserveBlobSize(ctx, blobSize)
	metrics.IncrementBlobThroughput(ctx, blobSize)
	metrics.IncrementBlobLifecycleEvents(ctx, "insert", blob.Key.String())

	// Update pool size metric
	p.count++
	metrics.SetBlobpoolCount(ctx, p.count)

	p.logger.Debug("stored blob", "hash", blob.Key.String(), "size", len(blob.Data))
}
