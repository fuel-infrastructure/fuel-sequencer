package metrics

import (
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	"github.com/hashicorp/go-metrics"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// ============================================================================
// Sync Service Performance Metrics
// ============================================================================

// ObserveBlobSyncLatency records the time taken to sync blobs from blobhub
func ObserveBlobSyncLatency(duration time.Duration) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBlobhub, "sync", "latency", "ms")...,
		)
	})
}

// SetBlobhubConnectionStatus updates the connection status metric
func SetBlobhubConnectionStatus(connected bool) {
	utils.SafeSetMetric(func() {
		status := 0
		if connected {
			status = 1
		}
		telemetry.SetGauge(
			float32(status),
			append(utils.KeysBlobhub, "connection", "status")...,
		)
	})
}

// IncrementBlobhubReconnections increments the reconnection counter
func IncrementBlobhubReconnections() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysBlobhub, "reconnections")...)
	})
}

// IncrementBlobhubErrors increments the error counter
func IncrementBlobhubErrors() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysBlobhub, "errors")...)
	})
}

// ============================================================================
// Blobpool Management Metrics
// ============================================================================

// SetBlobpoolCount updates the current blobpool size
func SetBlobpoolCount(count uint) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(count),
			append(utils.KeysBlobpool, "count")...,
		)
	})
}

// SetBlobpoolHitRatio updates the hit ratio metric
func SetBlobpoolHitRatio(ratio float64) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBlobpool, "hit", "ratio")...,
		)
	})
}

// SetBlobpoolMissRatio updates the miss ratio metric
func SetBlobpoolMissRatio(ratio float64) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBlobpool, "miss", "ratio")...,
		)
	})
}

// IncrementBlobpoolEvictions increments the eviction counter
func IncrementBlobpoolEvictions() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysBlobpool, "evictions")...)
	})
}

// ============================================================================
// Blob Processing Metrics
// ============================================================================

// ObserveBlobRetrievalTime records the time taken to retrieve a blob
func ObserveBlobRetrievalTime(duration time.Duration) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBlobpool, "retrieval", "time", "ms")...,
		)
	})
}

// ObserveBlobStorageLatency records the time taken to store a blob
func ObserveBlobStorageLatency(duration time.Duration) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBlobpool, "storage", "latency", "ms")...,
		)
	})
}

// ObserveBlobSize records the size of a blob for distribution analysis
func ObserveBlobSize(size int) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(size),
			append(utils.KeysBlobpool, "size", "bytes")...,
		)
	})
}

// ============================================================================
// Consensus Integration Metrics
// ============================================================================

// IncrementBlobProposalValidationSuccess increments the successful proposal counter
func IncrementBlobProposalValidationSuccess() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "blob", "proposal", "validation", "success")...)
	})
}

// IncrementBlobProposalValidationFailures increments the validation failure counter
func IncrementBlobProposalValidationFailures(reason error) {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "blob", "proposal", "validation", "failure"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("reason", reason.Error()),
			},
		)
	})
}

// ============================================================================
// Throughput and Lifecycle Metrics
// ============================================================================

// IncrementBlobThroughput increments the throughput counter
func IncrementBlobThroughput(blobSize int) {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysBlobpool, "throughput"),
			float32(blobSize),
			[]metrics.Label{
				telemetry.NewLabel("size_bytes", strconv.Itoa(blobSize)),
			},
		)
	})
}

// IncrementBlobLifecycleEvents increments the lifecycle events counter
func IncrementBlobLifecycleEvents(eventType string, blobHash string) {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysBlobpool, "lifecycle", "event"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("event_type", eventType),
				telemetry.NewLabel("blob_hash", blobHash),
			},
		)
	})
}

// ============================================================================
// Utility Functions for Hit/Miss Ratio Calculation
// ============================================================================

// UpdateBlobpoolHitMissRatios updates hit/miss ratios based on recent operations
func UpdateBlobpoolHitMissRatios(hits, misses uint) {
	utils.SafeSetMetric(func() {
		total := hits + misses
		if total > 0 {
			hitRatio := float64(hits) / float64(total)
			missRatio := float64(misses) / float64(total)

			telemetry.SetGauge(
				float32(hitRatio),
				append(utils.KeysBlobpool, "hit", "ratio")...,
			)
			telemetry.SetGauge(
				float32(missRatio),
				append(utils.KeysBlobpool, "miss", "ratio")...,
			)
		}
	})
}

// ============================================================================
// ACK Service Metrics
// ============================================================================

// IncrementACKSubmissions increments the ACK submission counter
func IncrementACKSubmissions() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysBlobhub, "ack", "submissions")...)
	})
}

// IncrementACKErrors increments the ACK error counter
func IncrementACKErrors() {
	utils.SafeSetMetric(func() {
		telemetry.IncrCounter(1, append(utils.KeysBlobhub, "ack", "errors")...)
	})
}

// ObserveACKLatency records the time taken to submit an ACK signature
func ObserveACKLatency(duration time.Duration) {
	utils.SafeSetMetric(func() {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBlobhub, "ack", "latency", "ms")...,
		)
	})
}
