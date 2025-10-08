package profiler

import (
	"math"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// generateBlob creates a new blob with random data following the configured distribution
func (p *BlobProfiler) generateBlob() *types.TrackedBlob {
	data := p.generator.GenerateBlob(true)

	key := store.NewKey(data)
	p.genBlobCount++
	return &types.TrackedBlob{
		Nonce: p.genBlobCount,
		StoredBlob: store.StoredBlob{
			Receipt: store.Receipt{
				Key: key,
			},
			Data: data,
		},
		Size: len(data),
		Submission: &types.Submission{
			Status: types.Pending,
		},
	}
}

// generateBlobs creates new blobs with random data following the configured distribution.
// Generates until the total size is at least the given size.
func (p *BlobProfiler) generateBlobs(size int) ([]*types.TrackedBlob, int) {
	var blobsSize int

	blobs := make([]*types.TrackedBlob, 0)
	for blobsSize < size {
		blob := p.generateBlob()
		blobsSize += blob.Size
		blobs = append(blobs, blob)
	}

	return blobs, size
}

func throughputDiscrepancy(
	intendedThroughput, dataSubmitted float64, profileRuntime time.Duration,
) (float64, int) {
	currentThroughput := dataSubmitted / profileRuntime.Seconds()
	lackingThroughput := intendedThroughput - currentThroughput
	requireData := 0.0
	if lackingThroughput > 0 {
		requireData = math.Ceil(lackingThroughput * profileRuntime.Seconds())
	}
	return lackingThroughput, int(requireData)
}

func (p *BlobProfiler) collectBlobs(
	intendedThroughput float64, duration time.Duration, currentSize float64,
) ([]*types.TrackedBlob, int) {
	// Figure out how many blobs to send
	_, needSize := throughputDiscrepancy(intendedThroughput, currentSize, duration)

	// return early if no blobs are needed
	if needSize <= 0 {
		return nil, 0
	}
	return p.generateBlobs(needSize)
}
