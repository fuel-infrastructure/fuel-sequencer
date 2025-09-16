package profiler

import (
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// updateReportSummary updates the profiler report with final statistics
func (p *BlobProfiler) updateReportSummary(blobs []*types.TrackedBlob, dataSubmittedBytes int) {
	totalBlobs := len(blobs)
	successfulBlobs := 0
	failedBlobs := 0

	for _, blob := range blobs {
		switch blob.Submission.Status {
		case types.Finalized:
			successfulBlobs++
		case types.Failed:
			failedBlobs++
		}
	}

	p.report.UpdateSummary(totalBlobs, successfulBlobs, failedBlobs, dataSubmittedBytes)
}

// WriteReport writes the final profiler report to file
func (p *BlobProfiler) WriteReport() error {
	return p.report.Write()
}

// ReportPath returns the path where the profiler report will be written
func (p *BlobProfiler) ReportPath() string {
	return p.report.Path()
}
