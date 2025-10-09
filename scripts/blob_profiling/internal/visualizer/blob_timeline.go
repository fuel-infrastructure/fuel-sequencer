package visualizer

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
		`+v.generateProfileLabels(12)+`
		
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
