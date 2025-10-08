package visualizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// generateBlockThroughputPlot creates a stacked bar chart showing total data in KiB per block.
// Block heights are normalized starting from 1, and each bar represents the total data posted in that block.
// Each blob within a block is shown as a separate segment in the stacked bar.
func (v *Visualizer) generateBlockThroughputPlot(blobsPath string) error {
	// First, get the data range to calculate appropriate image width
	var minHeight, maxHeight int64
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH block_data AS (
				SELECT 
					"name=height" as block_height
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=height" > 0 AND "name=size" > 0
			),
			min_height AS (
				SELECT MIN(block_height) as min_h FROM block_data
			),
			max_height AS (
				SELECT MAX(block_height) as max_h FROM block_data
			)
			SELECT 
				min_h as min_height,
				max_h as max_height,
				max_h - min_h + 1 as height_range
			FROM min_height, max_height
		) TO '`+v.dataDir+`/block_range_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	// Read the range data to calculate image width
	rangeData, err := os.ReadFile(filepath.Join(v.dataDir, "block_range_data.csv"))
	if err != nil {
		return fmt.Errorf("failed to read block range data: %v", err)
	}

	lines := strings.Split(string(rangeData), "\n")
	if len(lines) < 2 {
		return fmt.Errorf("invalid range data format")
	}

	// Parse the CSV data (skip header)
	parts := strings.Split(lines[1], ",")
	if len(parts) < 3 {
		return fmt.Errorf("invalid range data format")
	}

	minHeight, err = strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse min height: %v", err)
	}

	maxHeight, err = strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse max height: %v", err)
	}

	heightRange := maxHeight - minHeight + 1

	// Calculate image width: 10px per block + 50px base
	imageWidth := int(heightRange*10 + 50)
	if imageWidth < 800 {
		imageWidth = 800 // Minimum width
	}
	if imageWidth > 4000 {
		imageWidth = 4000 // Maximum width to prevent extremely large images
	}

	// Generate block throughput data with normalized block heights and individual blob sizes
	// Create line chart data with cumulative blob sizes and separators between blobs
	if err := v.runDuckDBQuery(blobsPath, `
		COPY (
			WITH block_data AS (
				SELECT 
					"name=height" as block_height,
					"name=size" as blob_size_bytes,
					"name=size" / 1048576.0 as blob_size_mib,
					"name=count" as blob_index,
					ROW_NUMBER() OVER (PARTITION BY "name=height" ORDER BY "name=count") as blob_order_in_block
				FROM read_parquet('`+blobsPath+`')
				WHERE "name=height" > 0 AND "name=size" > 0
				ORDER BY "name=height", "name=count"
			),
			min_height AS (
				SELECT MIN(block_height) as min_h FROM block_data
			),
			cumulative_data AS (
				SELECT 
					bd.block_height - min_h + 1 as normalized_height,
					bd.blob_size_mib,
					bd.blob_order_in_block,
					SUM(bd.blob_size_mib) OVER (
						PARTITION BY bd.block_height 
						ORDER BY bd.blob_order_in_block 
						ROWS UNBOUNDED PRECEDING
					) - bd.blob_size_mib as blob_bottom_mib,
					SUM(bd.blob_size_mib) OVER (
						PARTITION BY bd.block_height 
						ORDER BY bd.blob_order_in_block 
						ROWS UNBOUNDED PRECEDING
					) as blob_top_mib,
					bd.block_height as original_height
				FROM block_data bd, min_height
				ORDER BY normalized_height, blob_order_in_block
			),
			zero_point AS (
				SELECT 
					0 as normalized_height,
					0.0 as blob_size_mib,
					0 as blob_order_in_block,
					0.0 as blob_bottom_mib,
					0.0 as blob_top_mib,
					0 as original_height
			)
			SELECT * FROM zero_point
			UNION ALL
			SELECT 
				normalized_height,
				blob_size_mib,
				blob_order_in_block,
				blob_bottom_mib,
				blob_top_mib,
				original_height
			FROM cumulative_data
		) TO '`+v.dataDir+`/block_throughput_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	script := fmt.Sprintf(`
		set terminal pngcairo size %d,3000 enhanced font 'Arial,14'
		set output '%s/block_throughput.png'
		set datafile separator ","
		set grid
		
		set title "Blob Data Submitted per Block" font "Arial,54"
		set xlabel "Normalised Block Height" font "Arial,60" offset 0,-3
		set ylabel "Blob Data (MiB)" font "Arial,60" offset -12,0
		
		# Add profile information outside graph area with text wrapping
		%s
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		# Format y-axis to show MiB values with 0.25 MiB increments
		set format y "%%.2f"
		set ytics 0.25 font "Arial,36"
		
		# Set x-axis to show integer block heights with smaller font and better spacing
		set xtics 1 font "Arial,18"
		set format x "%%.0f"
		
		# Rotate x-axis labels to prevent overlap
		set xtics rotate by -45
		
		# Set margins to accommodate rotated labels and larger fonts
		set bmargin 12
		set lmargin 24
		
		# Create vertical line chart with no horizontal connections
		# Each blob is represented as a vertical line, with visible markers showing blob boundaries
		set style line 1 lc rgb 'midnight-blue' lw 18  # Very thick blue lines
		set style line 2 lc rgb 'gold' lw 1 pt 7 ps 1.0  # Large red markers for blob boundaries
		
		# Increase legend font size
		set key font "Arial,24"
		
		# Plot vertical lines only - no horizontal connections between blocks
		# Use impulses to create vertical lines without horizontal connections
		# Plot only top boundaries of each blob
		plot '%s/block_throughput_data.csv' using 1:5 with impulses ls 1 title "Cumulative Blob Data (MiB)", \
		     '' using 1:5 with points ls 2 title "Blob Boundaries"
	`, imageWidth*3, v.imagesDir, v.generateProfileLabels(36), v.dataDir)

	return v.generateGnuplotScript("block_throughput", script)
}
