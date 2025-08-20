package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/hashicorp/go-metrics"

	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// ============================================================================
// Sync Service Performance Metrics
// ============================================================================

// ObserveBlobSyncLatency records the time taken to sync blobs from blobhub
func ObserveBlobSyncLatency(duration time.Duration) {
	utils.SafeSetFinalizedMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBeginBlock, "blob", "sync", "latency", "ms")...,
		)
	})
}

// SetBlobhubConnectionStatus updates the connection status metric
func SetBlobhubConnectionStatus(connected bool) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		status := 0
		if connected {
			status = 1
		}
		telemetry.SetGauge(
			float32(status),
			append(utils.KeysStore, "blob", "blobhub", "connection", "status")...,
		)
	})
}

// IncrementBlobhubReconnections increments the reconnection counter
func IncrementBlobhubReconnections() {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysStore, "blob", "blobhub", "reconnections")...)
	})
}

// IncrementBlobhubErrors increments the error counter
func IncrementBlobhubErrors() {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysStore, "blob", "blobhub", "errors")...)
	})
}

// ============================================================================
// Blobpool Management Metrics
// ============================================================================

// SetBlobpoolCount updates the current blobpool size
func SetBlobpoolCount(count uint) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(count),
			append(utils.KeysBlobpool, "count")...,
		)
	})
}

// SetBlobpoolHitRatio updates the hit ratio metric
func SetBlobpoolHitRatio(ratio float64) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBlobpool, "hit", "ratio")...,
		)
	})
}

// SetBlobpoolMissRatio updates the miss ratio metric
func SetBlobpoolMissRatio(ratio float64) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBlobpool, "miss", "ratio")...,
		)
	})
}

// IncrementBlobpoolEvictions increments the eviction counter
func IncrementBlobpoolEvictions() {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysBlobpool, "evictions")...)
	})
}

// ============================================================================
// Blob Processing Metrics
// ============================================================================

// ObserveBlobRetrievalTime records the time taken to retrieve a blob
func ObserveBlobRetrievalTime(duration time.Duration) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysTxMsg, "blob", "retrieval", "time", "ms")...,
		)
	})
}

// ObserveBlobStorageLatency records the time taken to store a blob
func ObserveBlobStorageLatency(duration time.Duration) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysTxMsg, "blob", "storage", "latency", "ms")...,
		)
	})
}

// ObserveBlobSize records the size of a blob for distribution analysis
func ObserveBlobSize(size int) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.SetGauge(
			float32(size),
			append(utils.KeysTxMsg, "blob", "size", "bytes")...,
		)
	})
}

// ============================================================================
// Consensus Integration Metrics
// ============================================================================

// IncrementBlobProposalSuccess increments the successful proposal counter
func IncrementBlobProposalSuccess() {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "blob", "proposal", "success")...)
	})
}

// IncrementBlobValidationFailures increments the validation failure counter
func IncrementBlobValidationFailures(reason string) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "blob", "validation", "failure"),
			1,
			[]metrics.Label{
				telemetry.NewLabel("reason", reason),
			},
		)
	})
}

// ============================================================================
// Throughput and Lifecycle Metrics
// ============================================================================

// IncrementBlobThroughput increments the throughput counter
func IncrementBlobThroughput(blobSize int) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "blob", "throughput"),
			float32(blobSize),
			[]metrics.Label{
				telemetry.NewLabel("size_bytes", strconv.Itoa(blobSize)),
			},
		)
	})
}

// IncrementBlobLifecycleEvents increments the lifecycle events counter
func IncrementBlobLifecycleEvents(eventType string, blobHash string) {
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		telemetry.IncrCounterWithLabels(
			append(utils.KeysTxMsg, "blob", "lifecycle", "event"),
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
	utils.SafeSetMetric(context.Background(), func(_ sdk.Context) {
		total := hits + misses
		if total > 0 {
			hitRatio := float64(hits) / float64(total)
			missRatio := float64(misses) / float64(total)

			telemetry.SetGauge(
				float32(hitRatio),
				append(utils.KeysBeginBlock, "blob", "pool", "hit", "ratio")...,
			)
			telemetry.SetGauge(
				float32(missRatio),
				append(utils.KeysBeginBlock, "blob", "pool", "miss", "ratio")...,
			)
		}
	})
}
