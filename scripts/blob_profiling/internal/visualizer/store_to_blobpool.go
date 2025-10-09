package visualizer

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
		`+v.generateProfileLabels(12)+`
		
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
