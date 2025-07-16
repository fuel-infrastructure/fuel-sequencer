package service

import (
	"sync"

	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// BlobPool manages blob storage at the application level
type BlobPool struct {
	blobs    map[string][]byte // hash -> blob data
	metadata map[string]*types.BlobMetadata
	mu       sync.RWMutex
	logger   log.Logger

	// Channel to notify when blobs arrive
	blobReady chan string // blob hash notifications

	// Pending transactions waiting for blobs
	pendingTxs map[string][]sdk.Msg
}

// NewBlobPool creates a new blob pool
func NewBlobPool(logger log.Logger) *BlobPool {
	return &BlobPool{
		blobs:      make(map[string][]byte),
		metadata:   make(map[string]*types.BlobMetadata),
		logger:     logger.With("module", "blob-pool"),
		blobReady:  make(chan string, 100), // Buffer size of 100
		pendingTxs: make(map[string][]sdk.Msg),
	}
}

// HasBlob checks if a blob is available in the pool
func (bp *BlobPool) HasBlob(hash string) bool {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	_, exists := bp.blobs[hash]
	return exists
}

// GetBlob retrieves a blob from the pool
func (bp *BlobPool) GetBlob(hash string) ([]byte, error) {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	blob, exists := bp.blobs[hash]
	if !exists {
		return nil, types.ErrBlobNotFound
	}
	return blob, nil
}

// StoreBlob stores a blob in the pool
func (bp *BlobPool) StoreBlob(hash string, blob []byte, metadata *types.BlobMetadata) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	bp.blobs[hash] = blob
	bp.metadata[hash] = metadata

	// Notify waiting transactions
	select {
	case bp.blobReady <- hash:
	default:
		// Channel is full, drop notification
		bp.logger.Warn("blob ready channel full, dropping notification", "hash", hash)
	}
}

// ShelveTransaction stores a transaction that's waiting for a blob
func (bp *BlobPool) ShelveTransaction(blobHash string, msg sdk.Msg) {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	if bp.pendingTxs[blobHash] == nil {
		bp.pendingTxs[blobHash] = make([]sdk.Msg, 0)
	}
	bp.pendingTxs[blobHash] = append(bp.pendingTxs[blobHash], msg)
	bp.logger.Info("transaction shelved", "blob_hash", blobHash)
}

// ProcessShelvedTransactions processes transactions that were waiting for blobs
func (bp *BlobPool) ProcessShelvedTransactions(ctx sdk.Context, reprocessFunc func(sdk.Context, sdk.Msg) error) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	for blobHash, msgs := range bp.pendingTxs {
		if bp.hasBlob(blobHash) { // internal method without lock
			// Process all waiting transactions for this blob
			for _, msg := range msgs {
				if err := reprocessFunc(ctx, msg); err != nil {
					bp.logger.Error("failed to reprocess message", "error", err, "blob_hash", blobHash)
					continue
				}
			}
			delete(bp.pendingTxs, blobHash)
			bp.logger.Info("processed shelved transactions", "blob_hash", blobHash, "count", len(msgs))
		}
	}
}

// hasBlob is an internal method that doesn't acquire locks (caller must hold lock)
func (bp *BlobPool) hasBlob(hash string) bool {
	_, exists := bp.blobs[hash]
	return exists
}

// GetBlobReadyChannel returns the channel for blob ready notifications
func (bp *BlobPool) GetBlobReadyChannel() <-chan string {
	return bp.blobReady
}

// GetPendingTransactionCount returns the number of pending transactions
func (bp *BlobPool) GetPendingTransactionCount() int {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	count := 0
	for _, msgs := range bp.pendingTxs {
		count += len(msgs)
	}
	return count
}
