package visualizer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProfilerReport represents the structure of a profiler report JSON file
type ProfilerReport struct {
	Summary struct {
		ProfileDurationHuman string `json:"profile_duration_human"`
		ProfileDuration      int64  `json:"profile_duration"`
	} `json:"summary"`
	Environment struct {
		Config struct {
			Profile struct {
				Description string `json:"description"`
			} `json:"profile"`
		} `json:"config"`
	} `json:"environment"`
}

// ComparisonData represents a single data point for the comparison plot
type ComparisonData struct {
	System   string  // "MsgPostBlobMetadata" or "MsgPostBlob"
	BlobSize string  // e.g., "100_KiB", "1024_KiB", etc.
	Duration float64 // Duration in minutes
	RunIndex int     // Run number (1-5)
}

// GenerateComparisonPlot creates a scatter plot comparing the two systems across different blob sizes
func GenerateComparisonPlot(decoupledDir, fullpostDir string) error {
	// Create output directories
	outputDir := filepath.Join(filepath.Dir(decoupledDir), "comparison")
	imagesDir := filepath.Join(outputDir, "images")
	dataDir := filepath.Join(outputDir, "data")

	if err := os.MkdirAll(imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %v", err)
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	fmt.Printf("Generating comparison plot...\n")
	fmt.Printf("Decoupled system data: %s\n", decoupledDir)
	fmt.Printf("Fullpost system data: %s\n", fullpostDir)
	fmt.Printf("Output directory: %s\n", outputDir)

	// Extract data from both systems
	var allData []ComparisonData

	// Process decoupled system (MsgPostBlobMetadata)
	decoupledData, err := extractSystemData(decoupledDir, "MsgPostBlobMetadata")
	if err != nil {
		return fmt.Errorf("failed to extract decoupled system data: %v", err)
	}
	allData = append(allData, decoupledData...)

	// Process fullpost system (MsgPostBlob)
	fullpostData, err := extractSystemData(fullpostDir, "MsgPostBlob")
	if err != nil {
		return fmt.Errorf("failed to extract fullpost system data: %v", err)
	}
	allData = append(allData, fullpostData...)

	// Generate separate CSV files for each system
	decoupledCsvPath := filepath.Join(dataDir, "decoupled_data.csv")
	fullpostCsvPath := filepath.Join(dataDir, "fullpost_data.csv")

	var decoupledDataFiltered, fullpostDataFiltered []ComparisonData
	for _, d := range allData {
		if d.System == "MsgPostBlobMetadata" {
			decoupledDataFiltered = append(decoupledDataFiltered, d)
		} else {
			fullpostDataFiltered = append(fullpostDataFiltered, d)
		}
	}

	if err := writeComparisonCSV(decoupledCsvPath, decoupledDataFiltered); err != nil {
		return fmt.Errorf("failed to write decoupled CSV data: %v", err)
	}
	if err := writeComparisonCSV(fullpostCsvPath, fullpostDataFiltered); err != nil {
		return fmt.Errorf("failed to write fullpost CSV data: %v", err)
	}

	// Generate gnuplot script and create the plot
	if err := generateComparisonGnuplotScript(dataDir, imagesDir, decoupledCsvPath, fullpostCsvPath); err != nil {
		return fmt.Errorf("failed to generate comparison plot: %v", err)
	}

	fmt.Printf("✓ Comparison plot generated successfully\n")
	fmt.Printf("Check the '%s' directory for the generated graph.\n", imagesDir)
	return nil
}

// extractSystemData extracts comparison data from a system's output directory
func extractSystemData(systemDir, systemName string) ([]ComparisonData, error) {
	var data []ComparisonData

	// Walk through the directory structure to find profiler reports
	err := filepath.Walk(systemDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for profiler_report.json files
		if info.Name() == "profiler_report.json" {
			reportData, err := extractReportData(path, systemName)
			if err != nil {
				return fmt.Errorf("failed to extract data from %s: %v", path, err)
			}
			data = append(data, reportData)
		}

		return nil
	})

	return data, err
}

// extractReportData extracts data from a single profiler report
func extractReportData(reportPath, systemName string) (ComparisonData, error) {
	// Read the JSON file
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return ComparisonData{}, err
	}

	// Parse the JSON
	var report ProfilerReport
	if err := json.Unmarshal(data, &report); err != nil {
		return ComparisonData{}, err
	}

	// Extract blob size from the description
	// Format: "1_x_100_KiB_blobs_per_block" -> "100_KiB"
	description := report.Environment.Config.Profile.Description
	blobSize := extractBlobSize(description)

	// Convert duration from nanoseconds to minutes
	durationMinutes := float64(report.Summary.ProfileDuration) / (1000000000.0 * 60.0)

	// Extract run index from the directory structure
	runIndex := extractRunIndex(reportPath)

	return ComparisonData{
		System:   systemName,
		BlobSize: blobSize,
		Duration: durationMinutes,
		RunIndex: runIndex,
	}, nil
}

// extractBlobSize extracts the blob size from a description string
func extractBlobSize(description string) string {
	// Format: "1_x_100_KiB_blobs_per_block" -> "100_KiB"
	parts := strings.Split(description, "_")
	if len(parts) >= 3 {
		return parts[2] + "_" + parts[3] // e.g., "100_KiB"
	}
	return description
}

// extractRunIndex extracts the run index from the file path
func extractRunIndex(reportPath string) int {
	// Extract timestamp from the directory structure
	// Path format: .../1_x_100_KiB_blobs_per_block/2025-10-08_T_19_09_32/profiler_report.json
	dir := filepath.Dir(reportPath)
	timestampDir := filepath.Base(dir)

	// Extract hour and minute from timestamp (e.g., "19_09" from "2025-10-08_T_19_09_32")
	parts := strings.Split(timestampDir, "_")
	if len(parts) >= 4 {
		hourMin := parts[2] + "_" + parts[3] // e.g., "19_09"

		// Map timestamp to run index based on the observed pattern
		switch hourMin {
		case "19_09":
			return 1
		case "19_19":
			return 2
		case "19_29":
			return 3
		case "19_38":
			return 4
		case "19_40":
			return 5
		case "22_15":
			return 1
		case "22_29":
			return 2
		case "22_33":
			return 3
		case "22_35":
			return 4
		case "22_37":
			return 5
		}
	}

	// Default fallback
	return 1
}

// writeComparisonCSV writes the comparison data to a CSV file
func writeComparisonCSV(csvPath string, data []ComparisonData) error {
	file, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header
	file.WriteString("System,BlobSize,DurationMinutes,RunIndex\n")

	// Write data rows
	for _, d := range data {
		line := fmt.Sprintf("%s,%s,%.2f,%d\n", d.System, d.BlobSize, d.Duration, d.RunIndex)
		file.WriteString(line)
	}

	return nil
}

// generateComparisonGnuplotScript creates and executes a gnuplot script for the comparison plot
func generateComparisonGnuplotScript(dataDir, imagesDir, decoupledCsvPath, fullpostCsvPath string) error {
	// Convert to absolute paths
	absImagesDir, err := filepath.Abs(imagesDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for images dir: %v", err)
	}
	absDecoupledCsvPath, err := filepath.Abs(decoupledCsvPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for decoupled csv file: %v", err)
	}
	absFullpostCsvPath, err := filepath.Abs(fullpostCsvPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for fullpost csv file: %v", err)
	}

	script := `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '` + absImagesDir + `/system_comparison.png'
		set datafile separator ","
		set grid
		
		# Define colors and styles for the two systems
		set style line 1 lc rgb '#1f77b4' pt 7 ps 1.5 lw 2
		set style line 2 lc rgb '#ff7f0e' pt 9 ps 1.5 lw 2
		
		set title "Original Complete Blobs Posted vs Ongoing Decoupled Blobs (Metadata Only)" font "Arial,16"
		set xlabel "Blob Size Configuration" font "Arial,14"
		set ylabel "Total Duration (minutes)" font "Arial,14"
		
		# Set y-axis range to focus on 1-4 minutes
		set yrange [0:4]
		
		# Create a mapping for x-axis positions based on blob sizes
		set xrange [0:6]
		set xtics ("1x100 KiB" 1, "2x300 KiB" 2, "1x1 MiB" 3, "2x3 MiB" 4, "1x5 MiB" 5)
		
		# Plot the data with different symbols for each system
		plot '` + absDecoupledCsvPath + `' using (stringcolumn(2) eq "100_KiB" ? 1 : stringcolumn(2) eq "300_KiB" ? 2 : stringcolumn(2) eq "1024_KiB" ? 3 : stringcolumn(2) eq "3072_KiB" ? 4 : stringcolumn(2) eq "10240_KiB" ? 5 : 0):3 with points ls 1 title "MsgPostBlobMetadata", \
		     '` + absFullpostCsvPath + `' using (stringcolumn(2) eq "100_KiB" ? 1 : stringcolumn(2) eq "300_KiB" ? 2 : stringcolumn(2) eq "1024_KiB" ? 3 : stringcolumn(2) eq "3072_KiB" ? 4 : stringcolumn(2) eq "10240_KiB" ? 5 : 0):3 with points ls 2 title "MsgPostBlob"
	`

	// Write script to temporary file
	scriptPath := filepath.Join(dataDir, "comparison.gp")
	if err := os.WriteFile(scriptPath, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write gnuplot script: %v", err)
	}

	// Execute gnuplot
	gnuplotPath := "/opt/homebrew/bin/gnuplot"
	if _, err := exec.LookPath("gnuplot"); err == nil {
		gnuplotPath = "gnuplot"
	}
	cmd := exec.Command(gnuplotPath, scriptPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gnuplot failed: %v", err)
	}

	// Clean up script file
	os.Remove(scriptPath)
	return nil
}
