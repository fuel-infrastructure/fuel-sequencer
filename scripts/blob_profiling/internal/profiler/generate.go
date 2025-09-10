package profiler

import (
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// generateBlob creates a new blob with random data following the configured distribution
func (p *BlobProfiler) generateBlob() *types.TrackedBlob {
	data := p.generator.GenerateBlob(true)

	key := store.NewKey(data)
	return &types.TrackedBlob{
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

func (p *BlobProfiler) supplementBuffer(runtime time.Duration) {
	upcomingSize := p.config.Size(runtime + p.config.BufferDuration) // Appropriate buffer size at the current runtime
	requireBlobsSize := upcomingSize - p.bufferSize                  // Additional size that needs to be generated given the current buffer

	newBlobs, newBlobsSize := p.generateBlobs(requireBlobsSize)
	p.buffer = append(p.buffer, newBlobs...)
	p.bufferSize += newBlobsSize
}
