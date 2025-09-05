package profiler

import (
	"context"
	"log"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// castBlob is a fire-and-forget function that handles the blob submission.
// All while keeping track of the status along the blob lifetime.
//
// This process includes:
//   - T0: Start of blob being submitted (pre-requisites done)
//   - T1: Submit to blobhub first
//   - T2: Submit to sequencer
//   - T3: Check if blob is in blobpool
//   - T4: Check if blob is in proposal
//   - T5: Check if blob is in validated block
//   - T6: Check if blob is in finalized block
func (p *BlobProfiler) castBlob(
	ctx context.Context, cancel context.CancelFunc, blob *types.TrackedBlob,
	order uint64,
) {
	var err error
	defer func() {
		if err != nil {
			blob.Submission.Status = types.Failed
			cancel()
		}
	}()

	blob.Submission.StartTime = time.Now() // T0: Start of blob being submitted

	// T1: Submit to blobhub first
	receipt, err := p.blobhub.PutBlob(ctx, blob.StoredBlob.Data)
	if err != nil {
		log.Printf("failed to submit blob to blobhub: %v", err)
		return
	}
	blob.Submission.Status = types.Stored
	blob.Submission.StoreTime = receipt.StoredAt
	blob.Receipt = *receipt

	// T2: Submit to sequencer
	resp, err := p.sequencer.SubmitBlobMetadataTx(ctx, blob, order)
	if err != nil {
		log.Printf("failed to submit blob metadata to sequencer: %v", err)
		return
	}
	blob.Submission.Status = types.Submitted
	blob.TxResponse = resp // Only has tx hash since not finalised yet
	blob.Submission.MetadataTime = time.Now()

	// TODO: T3: Check if blob is in blobpool
	// TODO: T4: Check if blob is in proposal
	// TODO: T5: Check if blob is in validated block
	// TODO: T6: Check if blob is in finalized block
}
