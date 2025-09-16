package report

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProfilerReport is the main actor that manages profiling report generation
type ProfilerReport struct {
	data       *ProfilerReportData
	outputPath string
	startTime  time.Time
}

// New creates a new profiler report instance
func New(outputDir, runDirName, profileDescription string, runTimestamp time.Time) (*ProfilerReport, error) {
	outputPath := filepath.Join(outputDir, profileDescription, runDirName, "profiler_report.json")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return &ProfilerReport{
		data: &ProfilerReportData{
			Environment: ProfilerEnvironment{
				Timestamp: runTimestamp,
			},
			Events: make([]Event, 0),
			Summary: ProfilerSummary{
				StartTime: runTimestamp,
			},
		},
		outputPath: outputPath,
		startTime:  runTimestamp,
	}, nil
}
