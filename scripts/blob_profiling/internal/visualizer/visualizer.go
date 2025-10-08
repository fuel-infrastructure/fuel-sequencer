// Package visualizer provides comprehensive visualization capabilities for blob profiling data.
//
// The visualizer generates specific performance graphs from parquet data files:
//   - Throughput Analysis: Expected vs actual throughput with human-readable timestamps
//   - Blob Size Distribution: Distribution of blob sizes in the dataset
//   - Block Throughput: Total data in KiB per block with normalized block heights, showing individual blob sizes as stacked segments
//   - Blob Event Timeline: Timeline showing blob processing events (start, store, metadata, blobpool, finalized) sorted by submission order
//   - Store to Blobpool Timing: Duration from store completion to blobpool processing
//   - Store to Finalized Timing: Duration from store completion to finalization
//
// Prerequisites:
//   - DuckDB: For querying parquet files (brew install duckdb / apt install duckdb)
//   - gnuplot: For generating graphs (brew install gnuplot / apt install gnuplot)
//
// Usage:
//   - Standalone: ./blob_profiler -graphs -parquet <data_dir>
//   - With cleanup: ./blob_profiler -graphs -cleanup -parquet <data_dir>
//   - Makefile: make generate-graphs or make generate-clean
//
// Generated outputs are saved to {parquet_dir}/graphs/images/ as high-resolution PNG files.
package visualizer

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
)

// Visualizer handles the generation of performance graphs from parquet data.
// It uses DuckDB to query parquet files and gnuplot to generate high-resolution PNG images.
type Visualizer struct {
	parquetDir string         // Directory containing parquet files
	profile    config.Profile // Complete profile information including metadata
	cleanup    bool           // Whether to delete intermediate CSV files after graph generation
	outputDir  string         // Output directory for graphs and data (under parquet directory)
	imagesDir  string         // Images directory (under output directory)
	dataDir    string         // Data directory (under output directory)
}

// New creates a new visualizer instance with complete profile information.
func New(parquetDir string, profile config.Profile, cleanup bool) *Visualizer {
	outputDir := filepath.Join(parquetDir, "graphs")
	imagesDir := filepath.Join(outputDir, "images")
	dataDir := filepath.Join(outputDir, "data")

	return &Visualizer{
		parquetDir: parquetDir,
		profile:    profile,
		cleanup:    cleanup,
		outputDir:  outputDir,
		imagesDir:  imagesDir,
		dataDir:    dataDir,
	}
}

// GenerateGraphs creates all visualization graphs from the parquet data
func (v *Visualizer) GenerateGraphs() error {
	blobsFile := "blobs.parquet"
	throughputFile := "throughput.parquet"

	blobsPath := filepath.Join(v.parquetDir, blobsFile)
	throughputPath := filepath.Join(v.parquetDir, throughputFile)

	// Check if parquet files exist
	if _, err := os.Stat(blobsPath); os.IsNotExist(err) {
		return fmt.Errorf("blobs parquet file not found: %s", blobsPath)
	}
	if _, err := os.Stat(throughputPath); os.IsNotExist(err) {
		return fmt.Errorf("throughput parquet file not found: %s", throughputPath)
	}

	// Create output directories
	if err := os.MkdirAll(v.imagesDir, 0755); err != nil {
		return fmt.Errorf("failed to create images directory: %v", err)
	}

	fmt.Printf("Generating blob profiling visualizations...\n")
	fmt.Printf("Data directory: %s\n", v.parquetDir)
	fmt.Printf("Output directory: %s\n", v.outputDir)
	if v.cleanup {
		fmt.Printf("Cleanup mode: Visualisation data files will be deleted after image generation\n")
	}

	// Generate the specific plots requested
	visualizations := []struct {
		name        string
		description string
		generate    func() error
	}{
		{"throughput", "Expected vs Actual Throughput", func() error { return v.generateThroughputPlot(throughputPath) }},
		{"blob_sizes", "Blob Size Distribution", func() error { return v.generateBlobSizeDistribution(blobsPath) }},
		{"block_throughput", "Blob Submissions per block", func() error { return v.generateBlockThroughputPlot(blobsPath) }},
		// {"blob_timeline", "Blob Event Timeline", func() error { return v.generateBlobTimeline(blobsPath) }},
		// {"store_to_blobpool", "Time from Store to Blobpool", func() error { return v.generateStoreToBlobpoolPlot(blobsPath) }},
		// {"store_to_finalized", "Time from Store to Finalized", func() error { return v.generateStoreToFinalizedPlot(blobsPath) }},
	}

	for _, viz := range visualizations {
		fmt.Printf("Generating %s...\n", viz.description)
		if err := viz.generate(); err != nil {
			log.Printf("Error generating %s: %v", viz.name, err)
			return err
		}
		fmt.Printf("✓ %s generated successfully\n", viz.description)
	}

	// Clean up data files if requested
	if v.cleanup {
		fmt.Printf("Cleaning up data files...\n")
		if err := os.RemoveAll(v.dataDir); err != nil {
			log.Printf("Warning: Failed to clean up data files: %v", err)
		} else {
			fmt.Printf("✓ Data files cleaned up\n")
		}
	}

	fmt.Printf("\nAll visualizations completed! Check the '%s' directory for generated graphs.\n", v.imagesDir)
	return nil
}

// runDuckDBQuery executes a DuckDB query on the specified parquet file.
// It creates a temporary SQL file, runs DuckDB, and cleans up the temporary file.
func (v *Visualizer) runDuckDBQuery(inputFile, query string) error {
	// Ensure data directory exists
	if err := os.MkdirAll(v.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// Write query to temporary file
	tempQueryFile := filepath.Join(v.dataDir, "temp_query.sql")
	if err := os.WriteFile(tempQueryFile, []byte(query), 0644); err != nil {
		return fmt.Errorf("failed to write query file: %v", err)
	}

	// Run DuckDB query - try full path first, then fallback to PATH
	duckdbPath := "/opt/homebrew/bin/duckdb"
	if _, err := exec.LookPath("duckdb"); err == nil {
		duckdbPath = "duckdb"
	}
	cmd := exec.Command(duckdbPath, "-c", fmt.Sprintf(".read %s", tempQueryFile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("duckdb query failed: %v", err)
	}

	// Clean up temporary query file
	os.Remove(tempQueryFile)
	return nil
}

// generateGnuplotScript creates a gnuplot script and executes it to generate a PNG image.
// It creates a temporary script file, runs gnuplot, and cleans up the temporary file.
func (v *Visualizer) generateGnuplotScript(name, script string) error {
	// Write gnuplot script to temporary file
	tempScriptFile := filepath.Join(v.dataDir, name+".gp")
	if err := os.WriteFile(tempScriptFile, []byte(script), 0644); err != nil {
		return fmt.Errorf("failed to write gnuplot script: %v", err)
	}

	// Run gnuplot - try full path first, then fallback to PATH
	gnuplotPath := "/opt/homebrew/bin/gnuplot"
	if _, err := exec.LookPath("gnuplot"); err == nil {
		gnuplotPath = "gnuplot"
	}
	cmd := exec.Command(gnuplotPath, tempScriptFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gnuplot failed: %v", err)
	}

	// Clean up temporary script file
	os.Remove(tempScriptFile)
	return nil
}

// getProfileTitle returns a formatted title for the profile
func (v *Visualizer) getProfileTitle() string {
	return fmt.Sprintf("Profile: %s (%s)", v.profile.Type, v.profile.Purpose)
}

// generateProfileLabels creates gnuplot label commands for profile information
func (v *Visualizer) generateProfileLabels() string {
	// Profile title with spacing and larger font
	return fmt.Sprintf(`set label "` + v.getProfileTitle() + `" at screen 0.5, screen 0.9375 center font "Arial,36"`)
}
