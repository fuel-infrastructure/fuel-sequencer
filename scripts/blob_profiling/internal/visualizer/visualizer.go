// Package visualizer provides comprehensive visualization capabilities for blob profiling data.
//
// The visualizer generates five specific performance graphs from parquet data files:
//   - Throughput Analysis: Expected vs actual throughput with human-readable timestamps
//   - Blob Size Distribution: Distribution of blob sizes in the dataset
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
		{"blob_timeline", "Blob Event Timeline", func() error { return v.generateBlobTimeline(blobsPath) }},
		{"store_to_blobpool", "Time from Store to Blobpool", func() error { return v.generateStoreToBlobpoolPlot(blobsPath) }},
		{"store_to_finalized", "Time from Store to Finalized", func() error { return v.generateStoreToFinalizedPlot(blobsPath) }},
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

// generateThroughputPlot creates a graph showing expected vs actual throughput over time.
// Timestamps are converted to human-readable format starting from 0.
func (v *Visualizer) generateThroughputPlot(throughputPath string) error {
	// Generate throughput data with human-readable timestamps starting from 0
	if err := v.runDuckDBQuery(throughputPath, `
		COPY (
			WITH throughput_data AS (
				SELECT 
					"name=timestamp" as timestamp_ns,
					"name=expected_kib_per_sec" as expected_kib_per_sec,
					"name=actual_kib_per_sec" as actual_kib_per_sec
				FROM read_parquet('`+throughputPath+`')
				ORDER BY "name=timestamp"
			),
			min_timestamp AS (
				SELECT MIN(timestamp_ns) as min_ts FROM throughput_data
			)
			SELECT 
				(timestamp_ns - min_ts) / 1000000000.0 as time_sec,
				expected_kib_per_sec,
				actual_kib_per_sec
			FROM throughput_data, min_timestamp
		) TO '`+v.dataDir+`/throughput_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	return v.generateGnuplotScript("throughput", `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '`+v.imagesDir+`/throughput_analysis.png'
		set datafile separator ","
		set grid
		set style line 1 lc rgb '#1f77b4' lw 2
		set style line 2 lc rgb '#ff7f0e' lw 2
		
		set title "Expected vs Actual Throughput" font "Arial,16"
		set xlabel "Time (seconds)" font "Arial,14"
		set ylabel "Throughput (KiB/s)" font "Arial,14"
		
		# Add profile information outside graph area with text wrapping
		`+v.generateProfileLabels()+`
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		plot '`+v.dataDir+`/throughput_data.csv' using 1:2 with lines ls 1 title "Expected Throughput", \
		     '`+v.dataDir+`/throughput_data.csv' using 1:3 with lines ls 2 title "Actual Throughput"
	`)
}

// generateBlobSizeDistribution creates a histogram showing the distribution of blob sizes.
// Uses buckets aligned with the realistic distribution ranges from blob generation.
func (v *Visualizer) generateBlobSizeDistribution(blobsPath string) error {
	// Generate histogram data with buckets aligned to realistic distribution ranges
	// Buckets: 1-10KB, 10-100KB, 100KB-1MB, 1-10MB, 10-100MB, 100MB+
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH size_buckets AS (
				SELECT 
					CASE 
						WHEN "name=size" <= 10240 THEN '1-10 KB'
						WHEN "name=size" <= 102400 THEN '10-100 KB'
						WHEN "name=size" <= 1048576 THEN '100 KB-1 MB'
						WHEN "name=size" <= 10485760 THEN '1-10 MB'
						WHEN "name=size" <= 104857600 THEN '10-100 MB'
						ELSE '100+ MB'
					END as bucket,
					CASE 
						WHEN "name=size" <= 10240 THEN 1
						WHEN "name=size" <= 102400 THEN 2
						WHEN "name=size" <= 1048576 THEN 3
						WHEN "name=size" <= 10485760 THEN 4
						WHEN "name=size" <= 104857600 THEN 5
						ELSE 6
					END as bucket_order,
					COUNT(*) as count
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=size" > 0
				GROUP BY bucket, bucket_order
			)
			SELECT bucket, bucket_order, count
			FROM size_buckets
			ORDER BY bucket_order
		) TO '`+v.dataDir+`/blob_sizes_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	return v.generateGnuplotScript("blob_sizes", `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '`+v.imagesDir+`/blob_size_distribution.png'
		set datafile separator ","
		set grid
		set style fill solid 0.7
		set style line 1 lc rgb '#2ca02c'
		
		set title "Blob Size Distribution (Histogram)" font "Arial,16"
		set xlabel "Size Range" font "Arial,14"
		set ylabel "Count" font "Arial,14"
		
		# Add profile information outside graph area with text wrapping
		`+v.generateProfileLabels()+`
		
		# Auto-scale to fit data
		set autoscale y
		
		# Format y-axis to show integers
		set format y "%.0f"
		
		# Set x-axis to use categorical labels
		set xtics rotate by -45
		set xrange [0.5:6.5]
		
		# Create histogram-style bars
		plot '`+v.dataDir+`/blob_sizes_data.csv' using 2:3:xtic(1) with boxes ls 1 title "Blob Count"
	`)
}

// generateBlobTimeline creates a timeline graph showing blob processing events over time.
// Blobs are sorted by submission order, and timestamps are relative to the first blob's start time.
func (v *Visualizer) generateBlobTimeline(blobsPath string) error {
	// Generate blob timeline data with human-readable timestamps starting from 0
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH timeline_data AS (
				SELECT 
					"name=count" as blob_index,
					"name=start_time",
					"name=store_time",
					"name=metadata_time",
					"name=blobpool_time",
					"name=finalized_time"
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=finalized_time" > 0 AND "name=start_time" > 0
				ORDER BY "name=start_time"
			),
			min_start_time AS (
				SELECT MIN("name=start_time") as min_start FROM timeline_data
			)
			SELECT 
				blob_index,
				("name=start_time" - min_start) / 1000000000.0 as start_time_sec,
				("name=store_time" - min_start) / 1000000000.0 as store_time_sec,
				("name=metadata_time" - min_start) / 1000000000.0 as metadata_time_sec,
				("name=blobpool_time" - min_start) / 1000000000.0 as blobpool_time_sec,
				("name=finalized_time" - min_start) / 1000000000.0 as finalized_time_sec
			FROM timeline_data, min_start_time
		) TO '`+v.dataDir+`/blob_timeline_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	return v.generateGnuplotScript("blob_timeline", `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '`+v.imagesDir+`/blob_timeline.png'
		set datafile separator ","
		set grid
		set key top left
		set style line 1 lc rgb '#1f77b4' lw 2
		set style line 2 lc rgb '#ff7f0e' lw 2
		set style line 3 lc rgb '#2ca02c' lw 2
		set style line 4 lc rgb '#d62728' lw 2
		set style line 5 lc rgb '#9467bd' lw 2
		
		set title "Blob Event Timeline (Sorted by Submission Order)" font "Arial,16"
		set xlabel "Blob Index (Submission Order)" font "Arial,14"
		set ylabel "Time (seconds)" font "Arial,14"
		
		# Add profile information outside graph area with text wrapping
		`+v.generateProfileLabels()+`
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		# Force x-axis to show only integer ticks
		set xtics 1
		set format x "%.0f"
		# Format y-axis to show seconds with appropriate precision
		set format y "%.1f"
		
		plot '`+v.dataDir+`/blob_timeline_data.csv' using 1:2 with lines ls 1 title "Start Time", \
		     '`+v.dataDir+`/blob_timeline_data.csv' using 1:3 with lines ls 2 title "Store Time", \
		     '`+v.dataDir+`/blob_timeline_data.csv' using 1:4 with lines ls 3 title "Metadata Time", \
		     '`+v.dataDir+`/blob_timeline_data.csv' using 1:5 with lines ls 4 title "Blobpool Time", \
		     '`+v.dataDir+`/blob_timeline_data.csv' using 1:6 with lines ls 5 title "Finalized Time"
	`)
}

// generateStoreToBlobpoolPlot creates a scatter plot showing the duration from store completion to blobpool processing.
func (v *Visualizer) generateStoreToBlobpoolPlot(blobsPath string) error {
	// Generate store to blobpool timing data
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH timing_data AS (
				SELECT 
					"name=count" as blob_index,
					"name=store_time",
					"name=blobpool_time"
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=finalized_time" > 0 AND "name=start_time" > 0 AND "name=blobpool_time" > 0
				ORDER BY "name=start_time"
			),
			min_start_time AS (
				SELECT MIN("name=store_time") as min_start FROM timing_data
			)
			SELECT 
				blob_index,
				("name=store_time" - min_start) / 1000000000.0 as store_time_sec,
				("name=blobpool_time" - "name=store_time") / 1000000000.0 as duration_sec
			FROM timing_data, min_start_time
		) TO '`+v.dataDir+`/store_to_blobpool_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	return v.generateGnuplotScript("store_to_blobpool", `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '`+v.imagesDir+`/store_to_blobpool.png'
		set datafile separator ","
		set grid
		set style line 1 lc rgb '#ff7f0e' lw 2 pt 7 ps 0.5
		
		set title "Time from Store to Blobpool" font "Arial,16"
		set xlabel "Blob Index (Submission Order)" font "Arial,14"
		set ylabel "Duration (seconds)" font "Arial,14"
		
		# Add profile information outside graph area with text wrapping
		`+v.generateProfileLabels()+`
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		# Force x-axis to show only integer ticks
		set xtics 1
		set format x "%.0f"
		# Format y-axis to show seconds with appropriate precision
		set format y "%.3f"
		
		plot '`+v.dataDir+`/store_to_blobpool_data.csv' using 1:3 with points ls 1 title "Store to Blobpool Duration"
	`)
}

// generateStoreToFinalizedPlot creates a scatter plot showing the duration from store completion to finalization.
func (v *Visualizer) generateStoreToFinalizedPlot(blobsPath string) error {
	// Generate store to finalized timing data
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH timing_data AS (
				SELECT 
					"name=count" as blob_index,
					"name=store_time",
					"name=finalized_time"
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=finalized_time" > 0 AND "name=start_time" > 0
				ORDER BY "name=start_time"
			),
			min_start_time AS (
				SELECT MIN("name=store_time") as min_start FROM timing_data
			)
			SELECT 
				blob_index,
				("name=store_time" - min_start) / 1000000000.0 as store_time_sec,
				("name=finalized_time" - "name=store_time") / 1000000000.0 as duration_sec
			FROM timing_data, min_start_time
		) TO '`+v.dataDir+`/store_to_finalized_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	return v.generateGnuplotScript("store_to_finalized", `
		set terminal pngcairo size 1200,800 enhanced font 'Arial,12'
		set output '`+v.imagesDir+`/store_to_finalized.png'
		set datafile separator ","
		set grid
		set style line 1 lc rgb '#d62728' lw 2 pt 7 ps 0.5
		
		set title "Time from Store to Finalized" font "Arial,16"
		set xlabel "Blob Index (Submission Order)" font "Arial,14"
		set ylabel "Duration (seconds)" font "Arial,14"
		
		# Add profile information outside graph area with text wrapping
		`+v.generateProfileLabels()+`
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		# Force x-axis to show only integer ticks
		set xtics 1
		set format x "%.0f"
		# Format y-axis to show seconds with appropriate precision
		set format y "%.3f"
		
		plot '`+v.dataDir+`/store_to_finalized_data.csv' using 1:3 with points ls 1 title "Store to Finalized Duration"
	`)
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

	// Run DuckDB query
	cmd := exec.Command("duckdb", "-c", fmt.Sprintf(".read %s", tempQueryFile))
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

	// Run gnuplot
	cmd := exec.Command("gnuplot", tempScriptFile)
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
	// Profile title with spacing
	return fmt.Sprintf(`set label "` + v.getProfileTitle() + `" at screen 0.5, screen 0.9375 center font "Arial,12"`)
}
