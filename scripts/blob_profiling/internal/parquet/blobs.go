package parquet

import (
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// WriteBlobs adds multiple blobs to the write buffer and flushes if batch size is reached
func (w *Writer) writeBlobs(blobs []*types.TrackedBlob) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	for _, blob := range blobs {
		record := FromTrackedBlob(blob, w.profileStartTime)
		w.buffer = append(w.buffer, record)
	}

	// Flush if buffer is full
	if len(w.buffer) >= w.batchSize {
		return w.flushUnsafe()
	}

	return nil
}
