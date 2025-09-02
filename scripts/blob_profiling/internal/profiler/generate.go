package profiler

import (
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// generateBlob creates a new blob with random data following the configured distribution
func (p *BlobProfiler) generateBlob() *types.TrackedBlob {

	blob := p.generator.GenerateBlob(true)

	key := store.NewKey(blob)
	return &types.TrackedBlob{
		StoredBlob: store.StoredBlob{
			Receipt: store.Receipt{
				Key: key,
			},
			Data: blob,
		},
		Size: int64(len(blob)),
		Submission: &types.Submission{
			Status: types.Pending,
		},
	}
}
