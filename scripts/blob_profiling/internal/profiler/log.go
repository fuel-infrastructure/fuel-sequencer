package profiler

import (
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/parquet"
)

const (
	logFrequency = 1 * time.Second
)

func (p *BlobProfiler) logThroughput(
	plannedRate, currentThroughput float64,
	txCount uint64,
	dataSubmitted float64,
	blobCount int,
	seconds float64,
	timestamp time.Time,
) {
	expectedKiBPerSec := plannedRate / size.KiB
	actualKiBPerSec := currentThroughput / size.KiB
	submittedTxs := txCount - p.sequencer.Sender.Sequence
	submittedKiB := dataSubmitted / size.KiB

	p.logger.Info("throughput",
		"expected_KiB/s", expectedKiBPerSec,
		"actual_KiB/s", actualKiBPerSec,
		"submitted_txs", submittedTxs,
		"submitted_count", blobCount,
		"submitted_KiB", submittedKiB,
		// "pending_blobpool_count", p.pending, // concurrent access, but should be safe enough
		"duration_s", seconds,
	)

	// Write throughput data to parquet
	throughputRecord := parquet.NewThroughputRecord(
		int64(expectedKiBPerSec), int64(actualKiBPerSec),
		int64(submittedTxs), int64(blobCount), int64(submittedKiB),
		// int64(p.pending),
		seconds, timestamp,
	)
	if err := p.handler.WriteThroughput(throughputRecord); err != nil {
		p.report.RecordParquetEvent("write_throughput", err)
		p.logger.Error("failed to write throughput data", "error", err)
	}
}
