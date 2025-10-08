package parquet

import (
	"time"
)

// ThroughputRecord represents throughput metrics for parquet storage
type ThroughputRecord struct {
	// Timing information (nanoseconds since epoch)
	Timestamp int64 `parquet:"name=timestamp,type=INT64"`

	// Throughput metrics
	ExpectedKiBPerSec int64 `parquet:"name=expected_kib_per_sec,type=INT64"`
	ActualKiBPerSec   int64 `parquet:"name=actual_kib_per_sec,type=INT64"`

	// Transaction metrics
	SubmittedTxs   int64 `parquet:"name=submitted_txs,type=INT64"`
	SubmittedCount int64 `parquet:"name=submitted_count,type=INT64"`
	SubmittedKiB   int64 `parquet:"name=submitted_kib,type=INT64"`

	// Blobpool metrics
	PendingBlobpoolCount int64 `parquet:"name=pending_blobpool_count,type=INT64"`

	// Duration
	DurationSeconds float64 `parquet:"name=duration_seconds,type=DOUBLE"`
}

// NewThroughputRecord creates a new throughput record from the logged data
func NewThroughputRecord(
	expectedKiBPerSec, actualKiBPerSec int64,
	submittedTxs, submittedCount, submittedKiB int64,
	pendingBlobpoolCount int64,
	durationSeconds float64,
	timestamp time.Time,
) *ThroughputRecord {
	return &ThroughputRecord{
		Timestamp:            timestamp.UnixNano(),
		ExpectedKiBPerSec:    expectedKiBPerSec,
		ActualKiBPerSec:      actualKiBPerSec,
		SubmittedTxs:         submittedTxs,
		SubmittedCount:       submittedCount,
		SubmittedKiB:         submittedKiB,
		PendingBlobpoolCount: pendingBlobpoolCount,
		DurationSeconds:      durationSeconds,
	}
}

func writeThroughput(record *ThroughputRecord) func(
	buffer []*ThroughputRecord, profileStartTime time.Time) []*ThroughputRecord {
	return func(buffer []*ThroughputRecord, profileStartTime time.Time) []*ThroughputRecord {
		buffer = append(buffer, record)
		return buffer
	}
}
