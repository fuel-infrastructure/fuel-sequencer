package keeper

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/blob-storage/pkg/store/syncmap"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// Blobpool manages blob storage at the application level
type Blobpool struct {
	logger log.Logger

	store store.Store
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
		store:  syncmap.Store(nil, false),
		stats:  stats{},
	}

	metrics.SetBlobpoolCount(p.count)
	p.logger.Info("initialized new blobpool")
	return p
}

func (s *stats) update(hit bool) {
	if hit {
		s.hits++
	} else {
		s.misses++
	}

	// Update metrics
	// TODO: limit this to avoid excessive updates?
	metrics.UpdateBlobpoolHitMissRatios(s.hits, s.misses)
}

// Has checks if a blob is available in the pool
func (p *Blobpool) Has(ctx context.Context, hash store.Key) bool {
	exists, err := p.store.Has(ctx, hash)
	if err != nil {
		p.logger.Error("failed to check blob existence", "hash", hash.String(), "error", err)
		return false
	}

	p.update(exists) // Update hit/miss statistics

	p.logger.Debug("checked blob existence", "hash", hash.String(), "exists", exists)
	return exists
}

// Get retrieves a blob from the pool
func (p *Blobpool) Get(ctx context.Context, hash store.Key) (*store.StoredBlob, error) {
	start := time.Now()
	blob, err := p.store.Get(ctx, hash)
	if err != nil {
		if err == store.ErrNotFound {
			p.logger.Debug("blob not found", "hash", hash.String())
		} else {
			p.logger.Error("failed to get blob", "hash", hash.String(), "error", err)
		}
		return nil, fmt.Errorf("%w: %w", types.ErrBlobNotFound, err)
	}

	// Record metrics
	retrievalTime := time.Since(start)
	metrics.ObserveBlobRetrievalTime(retrievalTime)
	metrics.ObserveBlobSize(len(blob.Data))

	p.logger.Debug("retrieved blob", "hash", hash.String(), "size", len(blob.Data))
	return blob, nil
}

// Insert stores a blob in the pool
func (p *Blobpool) Insert(ctx context.Context, data []byte) {
	if data == nil {
		p.logger.Error("attempted to store nil blob")
		return
	}

	start := time.Now()
	receipt, err := p.store.Put(ctx, data)
	if err != nil {
		p.logger.Error("failed to store blob", "error", err)
		return
	}
	storageLatency := time.Since(start)

	// Record metrics
	blobSize := len(data)
	metrics.ObserveBlobStorageLatency(storageLatency)
	metrics.ObserveBlobSize(blobSize)
	metrics.IncrementBlobThroughput(blobSize)
	metrics.IncrementBlobLifecycleEvents("insert", receipt.Key.String())

	// Update pool size metric
	p.count++
	metrics.SetBlobpoolCount(p.count)

	p.logger.Debug("stored blob", "hash", receipt.Key.String(), "size", len(data))
}
