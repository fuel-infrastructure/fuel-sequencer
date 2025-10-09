package visualizer

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
		`+v.generateProfileLabels(12)+`
		
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
