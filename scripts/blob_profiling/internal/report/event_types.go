package report

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// RecordCastingEvent records casting-related events
func (pr *ProfilerReport) RecordCastingEvent(txHash string, blobCount int, err error) {
	context := map[string]interface{}{
		"tx_hash":    txHash,
		"blob_count": blobCount,
	}
	pr.RecordEvent(EventTypeCasting, "profiler", "Blob casting operation", err, context)
}

func eventContext(txHash string, blobs []*types.TrackedBlob) map[string]interface{} {
	context := map[string]interface{}{
		"tx_hash":    txHash,
		"blob_count": len(blobs),
	}

	// Add blob details if available
	if len(blobs) > 0 {
		blobDetails := make([]map[string]interface{}, 0, len(blobs))
		for _, blob := range blobs {
			blobInfo := map[string]interface{}{
				"hash": blob.Key.String(),
				"size": blob.Size,
			}
			blobDetails = append(blobDetails, blobInfo)
		}
		context["blobs"] = blobDetails
	}

	return context
}

// RecordCastingEventWithBlobs records casting-related events with blob details
func (pr *ProfilerReport) RecordCastingEventWithBlobs(txHash string, blobs []*types.TrackedBlob, err error) {
	context := eventContext(txHash, blobs)
	pr.RecordEvent(EventTypeCasting, "profiler", "Blob casting operation", err, context)
}

// RecordCatchingEvent records catching/tracking-related events
func (pr *ProfilerReport) RecordCatchingEvent(txHash string, blobCount int, err error) {
	context := map[string]interface{}{
		"tx_hash":    txHash,
		"blob_count": blobCount,
	}
	pr.RecordEvent(EventTypeCatching, "profiler", "Blob catching/tracking operation", err, context)
}

// RecordCatchingEventWithBlobs records catching/tracking-related events with blob details
func (pr *ProfilerReport) RecordCatchingEventWithBlobs(txHash string, blobs []*types.TrackedBlob, err error) {
	context := eventContext(txHash, blobs)
	pr.RecordEvent(EventTypeCatching, "profiler", "Blob catching/tracking operation", err, context)
}

// RecordConnectionEvent records connection-related events
func (pr *ProfilerReport) RecordConnectionEvent(service string, url string, err error) {
	context := map[string]interface{}{
		"service": service,
		"url":     url,
	}
	pr.RecordEvent(EventTypeConnection, service, "Service connection operation", err, context)
}

// RecordParquetEvent records parquet writing events
func (pr *ProfilerReport) RecordParquetEvent(operation string, err error) {
	context := map[string]interface{}{
		"operation": operation,
	}
	pr.RecordEvent(EventTypeParquet, "handler", fmt.Sprintf("Parquet %s operation", operation), err, context)
}

// RecordSequencerEvent records sequencer-related events
func (pr *ProfilerReport) RecordSequencerEvent(operation string, err error) {
	context := map[string]interface{}{
		"operation": operation,
	}
	pr.RecordEvent(EventTypeSequencer, "client", fmt.Sprintf("Sequencer %s operation", operation), err, context)
}

// RecordBlobpoolEvent records blobpool-related events
func (pr *ProfilerReport) RecordBlobpoolEvent(operation string, err error) {
	context := map[string]interface{}{
		"operation": operation,
	}
	pr.RecordEvent(EventTypeBlobpool, "client", fmt.Sprintf("Blobpool %s operation", operation), err, context)
}

// RecordInfoEvent records informational events
func (pr *ProfilerReport) RecordInfoEvent(component, message string, context map[string]interface{}) {
	pr.RecordEvent(EventTypeInfo, component, message, nil, context)
}

// RecordWarningEvent records warning events
func (pr *ProfilerReport) RecordWarningEvent(component, message string, context map[string]interface{}) {
	pr.RecordEvent(EventTypeWarning, component, message, nil, context)
}
