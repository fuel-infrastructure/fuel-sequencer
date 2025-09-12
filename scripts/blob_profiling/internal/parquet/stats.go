package parquet

import (
	"time"
)

// Stats contains statistics about the parquet handler, writer(s)
type Stats struct {
	BufferedRecords int
	BatchSize       int
	ProfileDuration time.Duration
}

// GetStats returns statistics about the writer
func (w *Writer) GetStats() Stats {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return Stats{
		BufferedRecords: len(w.buffer),
		BatchSize:       w.batchSize,
		ProfileDuration: time.Since(w.profileStartTime),
	}
}
