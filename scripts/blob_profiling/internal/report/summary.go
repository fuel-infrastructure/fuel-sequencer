package report

import (
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
)

// UpdateSummary updates the summary statistics
func (pr *ProfilerReport) UpdateSummary(totalBlobs, successfulBlobs, failedBlobs, dataSubmittedBytes int) {
	pr.data.Summary.TotalBlobs = totalBlobs
	pr.data.Summary.SuccessfulBlobs = successfulBlobs
	pr.data.Summary.FailedBlobs = failedBlobs
	pr.data.Summary.TotalEvents = len(pr.data.Events)
	pr.data.Summary.EndTime = time.Now()
	pr.data.Summary.ProfileDuration = pr.data.Summary.EndTime.Sub(pr.data.Summary.StartTime)
	pr.data.Summary.ProfileDurationHuman = formatDuration(pr.data.Summary.ProfileDuration)
	pr.data.Summary.DataSubmittedBytes = dataSubmittedBytes
	pr.data.Summary.DataSubmittedKiB = dataSubmittedBytes / size.KiB

	if pr.data.Summary.ProfileDuration.Seconds() > 0 {
		pr.data.Summary.AverageThroughput = float64(dataSubmittedBytes) / pr.data.Summary.ProfileDuration.Seconds()
		pr.data.Summary.AverageThroughputKiBPerSec = float64(dataSubmittedBytes/size.KiB) / pr.data.Summary.ProfileDuration.Seconds()
	}

	// Count events by type
	errorCount := 0
	warningCount := 0
	for _, event := range pr.data.Events {
		switch event.EventType {
		case EventTypeError, EventTypeCasting, EventTypeCatching, EventTypeConnection, EventTypeParquet, EventTypeSequencer, EventTypeBlobpool:
			errorCount++
		case EventTypeWarning:
			warningCount++
		}
	}
	pr.data.Summary.ErrorCount = errorCount
	pr.data.Summary.WarningCount = warningCount
}
