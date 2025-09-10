package profiler

import (
	"context"
	"log"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// catchBlob is a tracking function that monitors the blob submission.
// It keeps track of the status along the blob lifetime.
//
// This process includes:
//   - T3: Check if blob is in blobpool
//   - T4: Check if blob is in proposal
//   - T5: Check if blob is in validated block
//   - T6: Check if blob is in finalized block
func (p *BlobProfiler) catchBlobs(
	ctx context.Context, cancel context.CancelFunc,
	txHash string, blobs []*types.TrackedBlob,
) {
	var err error

	// TODO: T3: Check if blob is in blobpool
	// TODO: T4: Check if blob is in proposal
	// TODO: T5: Check if blob is in validated block

	// T6: Lookup blob metadata in finalized block
	txResult, err := p.sequencer.WaitForMetadataTxFinalisation(ctx, txHash)
	if err != nil {
		log.Printf("failed to wait for transaction confirmation: %v", err)
		for _, blob := range blobs {
			blob.Submission.Status = types.Failed
		}
		cancel()
		return
	}
	finalizedTime := p.blocktime(ctx, txResult.Height)
	for _, blob := range blobs {
		blob.Submission.Status = types.Finalized
		blob.MetadataTx = txResult
		blob.Submission.FinalizedTime = finalizedTime
	}
}
