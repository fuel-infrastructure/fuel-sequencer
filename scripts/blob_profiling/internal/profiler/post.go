package profiler

import (
	"context"
	"log"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// postBlobs handles the blob submission directly to sequencer using MsgPostBlob.
// Intended to be run sequentially.
//
// This process includes:
//   - T0: Start of blob being submitted (pre-requisites done)
//   - T1: Submit directly to sequencer using MsgPostBlob
func (p *BlobProfiler) postBlobs(
	ctx context.Context, cancel context.CancelFunc, blobs []*types.TrackedBlob,
	txSequence uint64, blobNonce int,
) (string, error) {
	for _, blob := range blobs {
		blob.Submission.StartTime = time.Now() // T0: Start of blob being submitted
	}

	// Submit all blobs directly to sequencer using MsgPostBlob
	resp, err := p.sequencer.SubmitPostBlob(ctx, blobs, txSequence, blobNonce)
	if err != nil {
		log.Printf("failed to submit blobs directly to sequencer: %v", err)
		for _, blob := range blobs {
			blob.Submission.Status = types.Failed
		}
		cancel()
		return "", err
	}

	txHash := resp.TxHash
	submissionTime := time.Now()
	for _, blob := range blobs {
		blob.Submission.Status = types.Submitted
		blob.MetadataHash = txHash // Store tx hash for tracking
		blob.Submission.MetadataTime = submissionTime
	}

	return txHash, nil
}
