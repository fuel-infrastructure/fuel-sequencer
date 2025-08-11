package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlobMetricsFunctions(t *testing.T) {
	// Test that all metric functions can be called without panicking
	ctx := context.TODO()

	// Test sync service performance metrics
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(ctx, time.Second)
		SetBlobhubConnectionStatus(ctx, true)
		SetBlobhubConnectionStatus(ctx, false)
		IncrementBlobhubReconnections(ctx)
		IncrementBlobhubErrors(ctx)
	})

	// Test blobpool management metrics
	assert.NotPanics(t, func() {
		SetBlobpoolCount(ctx, 100)
		SetBlobpoolHitRatio(ctx, 0.8)
		SetBlobpoolMissRatio(ctx, 0.2)
		IncrementBlobpoolEvictions(ctx)
		UpdateBlobpoolHitMissRatios(ctx, 80, 20)
	})

	// Test blob processing metrics
	assert.NotPanics(t, func() {
		ObserveBlobRetrievalTime(ctx, time.Millisecond)
		ObserveBlobStorageLatency(ctx, time.Millisecond)
		ObserveBlobSize(ctx, 1024)
	})

	// Test consensus integration metrics
	assert.NotPanics(t, func() {
		IncrementBlobProposalSuccess(ctx)
		IncrementBlobValidationFailures(ctx, "test_reason")
	})

	// Test throughput and lifecycle metrics
	assert.NotPanics(t, func() {
		IncrementBlobThroughput(ctx, 1024)
		IncrementBlobLifecycleEvents(ctx, "test_event", "test_hash")
	})
}

func TestBlobMetricsEdgeCases(t *testing.T) {
	ctx := context.TODO()

	// Test with zero durations
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(ctx, 0)
		ObserveBlobRetrievalTime(ctx, 0)
		ObserveBlobStorageLatency(ctx, 0)
	})

	// Test with very large durations
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(ctx, 24*time.Hour)
		ObserveBlobRetrievalTime(ctx, 24*time.Hour)
		ObserveBlobStorageLatency(ctx, 24*time.Hour)
	})

	// Test with negative sizes (should handle gracefully)
	assert.NotPanics(t, func() {
		ObserveBlobSize(ctx, -1)
		ObserveBlobSize(ctx, 0)
	})

	// Test hit/miss ratios with edge cases
	assert.NotPanics(t, func() {
		UpdateBlobpoolHitMissRatios(ctx, 0, 0)   // Both zero
		UpdateBlobpoolHitMissRatios(ctx, 1, 0)   // Only hits
		UpdateBlobpoolHitMissRatios(ctx, 0, 1)   // Only misses
		UpdateBlobpoolHitMissRatios(ctx, 100, 0) // 100% hit rate
		UpdateBlobpoolHitMissRatios(ctx, 0, 100) // 0% hit rate
	})

	// Test ratio setting with edge cases
	assert.NotPanics(t, func() {
		SetBlobpoolHitRatio(ctx, 0.0)
		SetBlobpoolHitRatio(ctx, 1.0)
		SetBlobpoolHitRatio(ctx, 0.5)
		SetBlobpoolMissRatio(ctx, 0.0)
		SetBlobpoolMissRatio(ctx, 1.0)
		SetBlobpoolMissRatio(ctx, 0.5)
	})
}

func TestBlobMetricsConcurrentAccess(t *testing.T) {
	ctx := context.TODO()
	done := make(chan bool)
	numGoroutines := 10

	// Test concurrent access to metrics functions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				IncrementBlobThroughput(ctx, j*1024)
				SetBlobpoolCount(ctx, uint(j))
				ObserveBlobSize(ctx, j*1024)
				ObserveBlobRetrievalTime(ctx, time.Duration(j)*time.Millisecond)
				SetBlobhubConnectionStatus(ctx, j%2 == 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Should not panic under concurrent access
	assert.NotPanics(t, func() {
		IncrementBlobThroughput(ctx, 100)
		SetBlobpoolCount(ctx, 100)
		SetBlobhubConnectionStatus(ctx, true)
	})
}

func TestBlobMetricsNilContext(t *testing.T) {
	ctx := context.TODO()

	// Test that metrics functions handle nil context gracefully
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(ctx, time.Second)
		SetBlobpoolCount(ctx, 100)
		IncrementBlobThroughput(ctx, 1024)
	})
}

func TestBlobMetricsLargeValues(t *testing.T) {
	ctx := context.TODO()

	// Test with very large values
	assert.NotPanics(t, func() {
		SetBlobpoolCount(ctx, 1000000)
		ObserveBlobSize(ctx, 1000000000) // 1GB
		SetBlobpoolHitRatio(ctx, 0.999999)
		SetBlobpoolMissRatio(ctx, 0.000001)
	})
}

func TestBlobMetricsStringValues(t *testing.T) {
	ctx := context.TODO()

	// Test with various string values for labels
	assert.NotPanics(t, func() {
		IncrementBlobValidationFailures(ctx, "")
		IncrementBlobValidationFailures(ctx, "very_long_reason_string_that_might_cause_issues")
		IncrementBlobValidationFailures(ctx, "special_chars_!@#$%^&*()")
		IncrementBlobLifecycleEvents(ctx, "event_type", "")
		IncrementBlobLifecycleEvents(ctx, "event_type", "very_long_hash_string_that_might_cause_issues")
		IncrementBlobLifecycleEvents(ctx, "event_type", "special_chars_!@#$%^&*()")
	})
}
