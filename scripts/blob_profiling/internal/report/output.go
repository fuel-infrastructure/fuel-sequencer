package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Write writes the complete report to JSON file
func (pr *ProfilerReport) Write() error {
	// Update final summary
	pr.data.Summary.EndTime = time.Now()
	pr.data.Summary.ProfileDuration = pr.data.Summary.EndTime.Sub(pr.data.Summary.StartTime)

	// Write to file
	data, err := json.MarshalIndent(pr.data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	if err := os.WriteFile(pr.outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	return nil
}

// Path returns the path where the report will be written
func (pr *ProfilerReport) Path() string {
	return pr.outputPath
}
