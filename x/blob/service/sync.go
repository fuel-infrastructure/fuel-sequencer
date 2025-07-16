package service

import (
	"context"
	"encoding/hex"
	"time"

	"cosmossdk.io/log"

	blobstore "github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/blob-storage/pkg/store/mock"
	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

// BlobNotification represents a blob availability notification
type BlobNotification struct {
	Hash     []byte
	Metadata types.BlobMetadata
	Error    error
}

// BSNSyncService handles blob synchronization from BSN
type BSNSyncService struct {
	blobPool  *BlobPool
	bsnClient blobstore.BlobTxStore
	logger    log.Logger
	stopCh    chan struct{}
}

// NewBSNSyncService creates a new BSN sync service
func NewBSNSyncService(blobPool *BlobPool, logger log.Logger) *BSNSyncService {
	return &BSNSyncService{
		blobPool:  blobPool,
		bsnClient: mock.NewMockBlobTxStore(), // Default to mock implementation
		logger:    logger.With("module", "bsn-sync"),
		stopCh:    make(chan struct{}),
	}
}

// SetBSNClient allows setting a custom BSN client implementation
func (bss *BSNSyncService) SetBSNClient(client blobstore.BlobTxStore) {
	bss.bsnClient = client
}

// Start starts the BSN sync service
func (bss *BSNSyncService) Start(ctx context.Context) error {
	bss.logger.Info("starting BSN sync service")
	go bss.continuousSync(ctx)
	return nil
}

// Stop stops the BSN sync service
func (bss *BSNSyncService) Stop() error {
	bss.logger.Info("stopping BSN sync service")
	close(bss.stopCh)
	return nil
}

// Subscribe returns a channel for blob notifications
func (bss *BSNSyncService) Subscribe() <-chan BlobNotification {
	// For now, return a simple channel
	// In a real implementation, this would be more sophisticated
	ch := make(chan BlobNotification, 10)
	return ch
}

// continuousSync continuously syncs blobs from BSN
func (bss *BSNSyncService) continuousSync(ctx context.Context) {
	stream, err := bss.bsnClient.Stream(ctx)
	if err != nil {
		bss.logger.Error("failed to start BSN stream", "error", err)
		return
	}

	for {
		select {
		case <-bss.stopCh:
			bss.logger.Info("BSN sync service stopped")
			return
		case <-ctx.Done():
			bss.logger.Info("BSN sync service context cancelled")
			return
		case result, ok := <-stream:
			if !ok {
				bss.logger.Warn("BSN stream closed, attempting to reconnect")
				time.Sleep(5 * time.Second)
				stream, err = bss.bsnClient.Stream(ctx)
				if err != nil {
					bss.logger.Error("failed to reconnect to BSN stream", "error", err)
					time.Sleep(10 * time.Second)
				}
				continue
			}

			if result.IsErr() {
				bss.logger.Error("BSN stream error", "error", result.Error())
				continue
			}

			blobTx := result.Value()
			for _, blob := range blobTx.Blobs {
				// Use the blob's NamespaceId as the hash for consistency
				hash := hex.EncodeToString(blob.NamespaceId)
				metadata := &types.BlobMetadata{
					Hash:      blob.NamespaceId,
					Size_:     uint64(len(blob.Data)),
					Timestamp: time.Now(),
				}

				bss.blobPool.StoreBlob(hash, blob.Data, metadata)
				bss.logger.Debug("stored blob from BSN", "hash", hash, "size", len(blob.Data))
			}
		}
	}
}
