package visualizer

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
