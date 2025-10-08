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
	// Create stacked data where each blob is a separate segment within each block
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
			stacked_data AS (
				SELECT 
					bd.block_height - min_h + 1 as normalized_height,
					bd.blob_size_mib,
					bd.blob_order_in_block,
					SUM(bd.blob_size_mib) OVER (
						PARTITION BY bd.block_height 
						ORDER BY bd.blob_order_in_block 
						ROWS UNBOUNDED PRECEDING
					) - bd.blob_size_mib as stack_bottom,
					SUM(bd.blob_size_mib) OVER (
						PARTITION BY bd.block_height 
						ORDER BY bd.blob_order_in_block 
						ROWS UNBOUNDED PRECEDING
					) as stack_top,
					bd.block_height as original_height
				FROM block_data bd, min_height
				ORDER BY normalized_height, blob_order_in_block
			)
			SELECT 
				normalized_height,
				blob_size_mib,
				blob_order_in_block,
				stack_bottom,
				stack_top,
				original_height
			FROM stacked_data
		) TO '`+v.dataDir+`/block_throughput_data.csv' (HEADER, DELIMITER ',');
	`); err != nil {
		return err
	}

	script := fmt.Sprintf(`
		set terminal pngcairo size %d,3000 enhanced font 'Arial,14'
		set output '%s/block_throughput.png'
		set datafile separator ","
		set grid
		set style fill solid 0.7
		set boxwidth 0.8
		
		set title "Blob Data Submitted per Block" font "Arial,54"
		set xlabel "Normalized Block Height" font "Arial,33" offset 0,-30
		set ylabel "Total Data per Block (MiB)" font "Arial,60" offset -1,0
		
		# Add profile information outside graph area with text wrapping
		%s
		
		# Auto-scale to fit data
		set autoscale x
		set autoscale y
		
		# Format y-axis to show KiB values with fractional part to avoid duplicates
		set format y "%%.1f"
		set ytics font "Arial,36"
		
		# Set x-axis to show integer block heights with smaller font and better spacing
		set xtics 1 font "Arial,24"
		set format x "%%.0f"
		
		# Rotate x-axis labels to prevent overlap
		set xtics rotate by -45
		
		# Set margins to accommodate rotated labels and larger fonts
		set bmargin 6
		set lmargin 24
		
		# Create stacked histogram using boxes with distinct colors
		# Each blob segment gets a different color based on its order in the block
		set style line 1 lc rgb '#1f77b4'  # Blue
		set style line 2 lc rgb '#ff7f0e'  # Orange  
		set style line 3 lc rgb '#2ca02c'  # Green
		set style line 4 lc rgb '#d62728'  # Red
		set style line 5 lc rgb '#9467bd'  # Purple
		set style line 6 lc rgb '#8c564b'  # Brown
		set style line 7 lc rgb '#e377c2'  # Pink
		set style line 8 lc rgb '#7f7f7f'  # Gray
		set style line 9 lc rgb '#bcbd22'  # Olive
		set style line 10 lc rgb '#17becf' # Cyan
		
		# Plot total data per block with distinct styling
		plot '%s/block_throughput_data.csv' using 1:5 with boxes ls 1 title "Total Data per Block (MiB)"
	`, imageWidth*3, v.imagesDir, v.generateProfileLabels(), v.dataDir)

	return v.generateGnuplotScript("block_throughput", script)
}
