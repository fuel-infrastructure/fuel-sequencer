package profiler

// import (
// 	"context"

// 	"github.com/fuel-infrastructure/blob-storage/pkg/store"
// 	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
// )

// type poolStatus struct {
// 	pending int // number of blobs in pending to pass through to the blobpool

// 	refs     map[store.Key]*types.TrackedBlob
// 	expected map[store.Key]bool
// }

// func (p *BlobProfiler) catchBlobpool(
// 	ctx context.Context,
// 	expect <-chan *types.TrackedBlob,
// 	stream <-chan *store.StoredBlob,
// 	consume <-chan store.Key,
// ) {
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case blob := <-expect: // blobs to expect soon in the blobpool
// 			if _, ok := p.expected[blob.Key]; ok {
// 				p.logger.Debug("blob arrived in blobpool before marked as expected", "key", blob.Key)
// 				delete(p.expected, blob.Key)
// 				blob.Submission.Status = types.InBlobpool
// 				blob.Submission.BlobpoolTime = blob.StoredAt
// 				// not modifying pending; or rather pending-1+1
// 			} else {
// 				p.refs[blob.Key] = blob
// 				p.pending++
// 			}
// 		case blob := <-stream: // blobs that are now in the blobpool
// 			tracker, ok := p.refs[blob.Key]
// 			if !ok {
// 				p.logger.Debug("unexpected blob arrived in blobpool", "key", blob.Key)
// 				p.expected[blob.Key] = true
// 				// not marking as pending; to be handled when the expected blob arrives
// 			} else {
// 				tracker.Submission.Status = types.InBlobpool
// 				tracker.Submission.BlobpoolTime = blob.StoredAt
// 				p.pending--
// 			}
// 		case key := <-consume: // blobs that are effectively finalised
// 			// consume implies blob is finalised
// 			// at this point we don't care if not caught in blobpool
// 			// i.e. don't mind no-ops
// 			delete(p.refs, key)
// 		}
// 	}
// }
