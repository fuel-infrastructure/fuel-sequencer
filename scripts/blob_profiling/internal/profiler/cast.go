package profiler

import (
	"context"
	"log"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// castBlobs handles the blob submission.
// Intended to be run sequentially.
//
// This process includes:
//   - T0: Start of blob being submitted (pre-requisites done)
//   - T1: Submit to blobhub first
//   - T2: Submit to sequencer
func (p *BlobProfiler) castBlobs(
	ctx context.Context, cancel context.CancelFunc, blobs []*types.TrackedBlob,
	txSequence uint64, blobNonce int,
) (string, error) {
	var err error

	for _, blob := range blobs {
		blob.Submission.StartTime = time.Now() // T0: Start of blob being submitted

		// T1: Submit to blobhub first
		receipt, err := p.blobhub.PutBlob(ctx, blob.StoredBlob.Data)
		if err != nil {
			log.Printf("failed to submit blob to blobhub: %v", err)
			blob.Submission.Status = types.Failed
			cancel()
		}
		blob.Submission.Status = types.Stored
		blob.Submission.StoreTime = receipt.StoredAt
		blob.Receipt = *receipt
	}

	// T2: Submit to sequencer
	resp, err := p.sequencer.SubmitBlobMetadataTx(ctx, blobs, txSequence, blobNonce)
	if err != nil {
		log.Printf("failed to submit blob metadata to sequencer: %v", err)
		for _, blob := range blobs {
			blob.Submission.Status = types.Failed
		}
		cancel()
		return "", err
	}
	txHash := resp.TxHash
	metadataTime := time.Now()
	for _, blob := range blobs {
		blob.Submission.Status = types.Submitted
		blob.MetadataHash = txHash // Only has tx hash since not finalised yet
		blob.Submission.MetadataTime = metadataTime
	}

	return txHash, nil
}
