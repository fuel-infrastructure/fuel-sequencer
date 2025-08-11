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
func ObserveBlobSyncLatency(goCtx context.Context, duration time.Duration) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysBeginBlock, "blob", "sync", "latency", "ms")...,
		)
	})
}

// SetBlobhubConnectionStatus updates the connection status metric
func SetBlobhubConnectionStatus(goCtx context.Context, connected bool) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
func IncrementBlobhubReconnections(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysStore, "blob", "blobhub", "reconnections")...)
	})
}

// IncrementBlobhubErrors increments the error counter
func IncrementBlobhubErrors(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysStore, "blob", "blobhub", "errors")...)
	})
}

// ============================================================================
// Blobpool Management Metrics
// ============================================================================

// SetBlobpoolCount updates the current blobpool size
func SetBlobpoolCount(goCtx context.Context, count uint) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(count),
			append(utils.KeysBeginBlock, "blob", "pool", "count")...,
		)
	})
}

// SetBlobpoolHitRatio updates the hit ratio metric
func SetBlobpoolHitRatio(goCtx context.Context, ratio float64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBeginBlock, "blob", "pool", "hit", "ratio")...,
		)
	})
}

// SetBlobpoolMissRatio updates the miss ratio metric
func SetBlobpoolMissRatio(goCtx context.Context, ratio float64) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(ratio),
			append(utils.KeysBeginBlock, "blob", "pool", "miss", "ratio")...,
		)
	})
}

// IncrementBlobpoolEvictions increments the eviction counter
func IncrementBlobpoolEvictions(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysBeginBlock, "blob", "pool", "evictions")...)
	})
}

// ============================================================================
// Blob Processing Metrics
// ============================================================================

// ObserveBlobRetrievalTime records the time taken to retrieve a blob
func ObserveBlobRetrievalTime(goCtx context.Context, duration time.Duration) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysTxMsg, "blob", "retrieval", "time", "ms")...,
		)
	})
}

// ObserveBlobStorageLatency records the time taken to store a blob
func ObserveBlobStorageLatency(goCtx context.Context, duration time.Duration) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.SetGauge(
			float32(duration.Milliseconds()),
			append(utils.KeysTxMsg, "blob", "storage", "latency", "ms")...,
		)
	})
}

// ObserveBlobSize records the size of a blob for distribution analysis
func ObserveBlobSize(goCtx context.Context, size int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
func IncrementBlobProposalSuccess(goCtx context.Context) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
		telemetry.IncrCounter(1, append(utils.KeysTxMsg, "blob", "proposal", "success")...)
	})
}

// IncrementBlobValidationFailures increments the validation failure counter
func IncrementBlobValidationFailures(goCtx context.Context, reason string) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
func IncrementBlobThroughput(goCtx context.Context, blobSize int) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
func IncrementBlobLifecycleEvents(goCtx context.Context, eventType string, blobHash string) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
func UpdateBlobpoolHitMissRatios(goCtx context.Context, hits, misses uint) {
	utils.SafeSetMetric(goCtx, func(ctx sdk.Context) {
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
