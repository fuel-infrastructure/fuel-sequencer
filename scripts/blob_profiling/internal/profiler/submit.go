package profiler

import (
	"context"
	"log"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// submitBlob submits a blob to blobhub and then submits metadata to sequencer
func (p *BlobProfiler) submitBlob(
	ctx context.Context, cancel context.CancelFunc, blob *types.TrackedBlob,
) {
	blob.Submission.StartTime = time.Now()

	// Submit to blobhub
	receipt, err := p.blobhubClient.PutBlob(ctx, blob.Data)
	if err != nil {
		log.Printf("Failed to submit blob to blobhub: %v", err)
		blob.Submission.Status = types.Failed
		cancel()
		return
	}
	blob.Submission.Status = types.Stored
	blob.Submission.StoreTime = receipt.StoredAt
	blob.Receipt = *receipt
}
