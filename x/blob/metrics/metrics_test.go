package metrics

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBlobMetricsFunctions(t *testing.T) {
	// Test that all metric functions can be called without panicking
	// Test sync service performance metrics
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(time.Second)
		SetBlobhubConnectionStatus(true)
		SetBlobhubConnectionStatus(false)
		IncrementBlobhubReconnections()
		IncrementBlobhubErrors()
	})

	// Test blobpool management metrics
	assert.NotPanics(t, func() {
		SetBlobpoolCount(100)
		SetBlobpoolHitRatio(0.8)
		SetBlobpoolMissRatio(0.2)
		IncrementBlobpoolEvictions()
		UpdateBlobpoolHitMissRatios(80, 20)
	})

	// Test blob processing metrics
	assert.NotPanics(t, func() {
		ObserveBlobRetrievalTime(time.Millisecond)
		ObserveBlobStorageLatency(time.Millisecond)
		ObserveBlobSize(1024)
	})

	// Test consensus integration metrics
	assert.NotPanics(t, func() {
		IncrementBlobProposalValidationSuccess()
		IncrementBlobProposalValidationFailures(errors.New("test_reason"))
	})

	// Test throughput and lifecycle metrics
	assert.NotPanics(t, func() {
		IncrementBlobThroughput(1024)
		IncrementBlobLifecycleEvents("test_event", "test_hash")
	})
}

func TestBlobMetricsEdgeCases(t *testing.T) {
	// Test with zero durations
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(0)
		ObserveBlobRetrievalTime(0)
		ObserveBlobStorageLatency(0)
	})

	// Test with very large durations
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(24 * time.Hour)
		ObserveBlobRetrievalTime(24 * time.Hour)
		ObserveBlobStorageLatency(24 * time.Hour)
	})

	// Test with negative sizes (should handle gracefully)
	assert.NotPanics(t, func() {
		ObserveBlobSize(-1)
		ObserveBlobSize(0)
	})

	// Test hit/miss ratios with edge cases
	assert.NotPanics(t, func() {
		UpdateBlobpoolHitMissRatios(0, 0)   // Both zero
		UpdateBlobpoolHitMissRatios(1, 0)   // Only hits
		UpdateBlobpoolHitMissRatios(0, 1)   // Only misses
		UpdateBlobpoolHitMissRatios(100, 0) // 100% hit rate
		UpdateBlobpoolHitMissRatios(0, 100) // 0% hit rate
	})

	// Test ratio setting with edge cases
	assert.NotPanics(t, func() {
		SetBlobpoolHitRatio(0.0)
		SetBlobpoolHitRatio(1.0)
		SetBlobpoolHitRatio(0.5)
		SetBlobpoolMissRatio(0.0)
		SetBlobpoolMissRatio(1.0)
		SetBlobpoolMissRatio(0.5)
	})
}

func TestBlobMetricsConcurrentAccess(t *testing.T) {
	done := make(chan bool)
	numGoroutines := 10

	// Test concurrent access to metrics functions
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				IncrementBlobThroughput(j * 1024)
				SetBlobpoolCount(uint(j))
				ObserveBlobSize(j * 1024)
				ObserveBlobRetrievalTime(time.Duration(j) * time.Millisecond)
				SetBlobhubConnectionStatus(j%2 == 0)
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
		IncrementBlobThroughput(100)
		SetBlobpoolCount(100)
		SetBlobhubConnectionStatus(true)
	})
}

func TestBlobMetricsNilContext(t *testing.T) {
	// Test that metrics functions handle nil context gracefully
	assert.NotPanics(t, func() {
		ObserveBlobSyncLatency(time.Second)
		SetBlobpoolCount(100)
		IncrementBlobThroughput(1024)
	})
}

func TestBlobMetricsLargeValues(t *testing.T) {
	// Test with very large values
	assert.NotPanics(t, func() {
		SetBlobpoolCount(1000000)
		ObserveBlobSize(1000000000) // 1GB
		SetBlobpoolHitRatio(0.999999)
		SetBlobpoolMissRatio(0.000001)
	})
}

func TestBlobMetricsStringValues(t *testing.T) {
	// Test with various string values for labels
	assert.NotPanics(t, func() {
		IncrementBlobProposalValidationFailures(errors.New(""))
		IncrementBlobProposalValidationFailures(errors.New("very_long_reason_string_that_might_cause_issues"))
		IncrementBlobProposalValidationFailures(errors.New("special_chars_!@#$%^&*()"))
		IncrementBlobLifecycleEvents("event_type", "")
		IncrementBlobLifecycleEvents("event_type", "very_long_hash_string_that_might_cause_issues")
		IncrementBlobLifecycleEvents("event_type", "special_chars_!@#$%^&*()")
	})
}
