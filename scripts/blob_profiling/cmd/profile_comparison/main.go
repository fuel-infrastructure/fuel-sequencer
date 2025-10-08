package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/report"
)

// ProfileRun represents a single profiling run
type ProfileRun struct {
	CommitHash                 string
	CommitHashShort            string
	BlobStorageCommitHash      string
	BlobStorageCommitHashShort string
	ProfileType                string
	BlobSize                   string
	MaxThroughput              float64
	MinThroughput              float64
	AvgThroughput              float64
	ThroughputVariance         float64
	SuccessRate                float64
	Duration                   time.Duration
	Timestamp                  time.Time
	RunPath                    string
}

// ProfileComparison groups runs by commit hash, profile type, and blob size
type ProfileComparison struct {
	CommitHash                 string
	CommitHashShort            string
	BlobStorageCommitHash      string
	BlobStorageCommitHashShort string
	ProfileType                string
	BlobSize                   string
	MaxThroughput              float64
	MinThroughput              float64
	AvgThroughput              float64
	ThroughputVariance         float64
	SuccessRate                float64
	Duration                   time.Duration
	RunCount                   int
	Runs                       []ProfileRun
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <output_directory> [options]")
		fmt.Println("Example: go run main.go ../../output")
		fmt.Println("Options:")
		fmt.Println("  --csv        Output in CSV format")
		fmt.Println("  --individual Show individual runs instead of grouping")
		fmt.Println("  --help       Show this help message")
		os.Exit(1)
	}

	outputDir := os.Args[1]
	outputFormat := "table" // default format
	showIndividual := false

	// Parse command line options
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--csv":
			outputFormat = "csv"
		case "--individual":
			showIndividual = true
		case "--help":
			fmt.Println("Usage: go run main.go <output_directory> [options]")
			fmt.Println("Example: go run main.go ../../output")
			fmt.Println("Options:")
			fmt.Println("  --csv        Output in CSV format")
			fmt.Println("  --individual Show individual runs instead of grouping")
			fmt.Println("  --help       Show this help message")
			os.Exit(0)
		default:
			fmt.Printf("Unknown option: %s\n", os.Args[i])
			os.Exit(1)
		}
	}

	// Parse all profiler reports
	runs, err := parseProfilerReports(outputDir)
	if err != nil {
		log.Fatalf("Error parsing profiler reports: %v", err)
	}

	if len(runs) == 0 {
		fmt.Println("No profiler reports found in the specified directory")
		os.Exit(1)
	}

	// Group runs by commit hash
	comparisons := groupRunsByCommit(runs)

	// Generate output in requested format
	if showIndividual {
		switch outputFormat {
		case "csv":
			generateIndividualCSVOutput(runs)
		default:
			generateIndividualTableOutput(runs)
		}
	} else {
		switch outputFormat {
		case "csv":
			generateCSVOutput(comparisons)
		default:
			generateComparisonTable(comparisons)
		}
	}
}

func parseProfilerReports(outputDir string) ([]ProfileRun, error) {
	var runs []ProfileRun

	err := filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Look for profiler_report.json files
		if d.Name() == "profiler_report.json" {
			run, err := parseProfilerReport(path)
			if err != nil {
				log.Printf("Warning: Failed to parse %s: %v", path, err)
				return nil // Continue processing other files
			}
			runs = append(runs, run)
		}

		return nil
	})

	return runs, err
}

func parseProfilerReport(reportPath string) (ProfileRun, error) {
	data, err := os.ReadFile(reportPath)
	if err != nil {
		return ProfileRun{}, err
	}

	var reportData report.ProfilerReportData
	if err := json.Unmarshal(data, &reportData); err != nil {
		return ProfileRun{}, err
	}

	// Extract profile information
	profileType := reportData.Environment.Config.Profile.Type
	blobSize := reportData.Environment.Config.BlobDistributionHuman

	// Calculate success rate
	successRate := 0.0
	if reportData.Summary.TotalBlobs > 0 {
		successRate = float64(reportData.Summary.SuccessfulBlobs) / float64(reportData.Summary.TotalBlobs) * 100
	}

	// Get throughput from summary (convert KiB/s to MiB/s)
	avgThroughput := reportData.Summary.AverageThroughputKiBPerSec / 1024.0

	// For now, set min throughput to average and variance to 0
	// TODO: Calculate actual min/max/variance from detailed throughput data
	minThroughput := avgThroughput
	maxThroughput := avgThroughput
	variance := 0.0

	return ProfileRun{
		CommitHash:                 reportData.Environment.CommitHash,
		CommitHashShort:            reportData.Environment.CommitHashShort,
		BlobStorageCommitHash:      reportData.Environment.BlobStorageCommitHash,
		BlobStorageCommitHashShort: reportData.Environment.BlobStorageCommitHashShort,
		ProfileType:                profileType,
		BlobSize:                   blobSize,
		MaxThroughput:              maxThroughput,
		MinThroughput:              minThroughput,
		AvgThroughput:              avgThroughput,
		ThroughputVariance:         variance,
		SuccessRate:                successRate,
		Duration:                   reportData.Summary.ProfileDuration,
		Timestamp:                  reportData.Environment.Timestamp,
		RunPath:                    reportPath,
	}, nil
}

func groupRunsByCommit(runs []ProfileRun) map[string]*ProfileComparison {
	comparisons := make(map[string]*ProfileComparison)

	for _, run := range runs {
		// Create composite key from commit hashes, profile type, and blob size
		key := fmt.Sprintf("%s-%s-%s-%s", run.CommitHash, run.BlobStorageCommitHash, run.ProfileType, run.BlobSize)
		if comp, exists := comparisons[key]; exists {
			comp.Runs = append(comp.Runs, run)
		} else {
			comparisons[key] = &ProfileComparison{
				CommitHash:                 run.CommitHash,
				CommitHashShort:            run.CommitHashShort,
				BlobStorageCommitHash:      run.BlobStorageCommitHash,
				BlobStorageCommitHashShort: run.BlobStorageCommitHashShort,
				ProfileType:                run.ProfileType,
				BlobSize:                   run.BlobSize,
				Runs:                       []ProfileRun{run},
			}
		}
	}

	// Calculate aggregated metrics for each comparison
	for _, comp := range comparisons {
		comp.calculateAggregatedMetrics()
	}

	return comparisons
}

// calculateAggregatedMetrics calculates min, max, average, variance, and other aggregated metrics
func (pc *ProfileComparison) calculateAggregatedMetrics() {
	if len(pc.Runs) == 0 {
		return
	}

	// Initialize with first run
	firstRun := pc.Runs[0]
	pc.MaxThroughput = firstRun.MaxThroughput
	pc.MinThroughput = firstRun.MinThroughput
	pc.AvgThroughput = firstRun.AvgThroughput
	pc.SuccessRate = firstRun.SuccessRate
	pc.Duration = firstRun.Duration
	pc.RunCount = len(pc.Runs)

	// Calculate min, max, and average
	var totalThroughput float64
	var totalSuccessRate float64
	var totalDuration time.Duration

	for _, run := range pc.Runs {
		totalThroughput += run.AvgThroughput
		totalSuccessRate += run.SuccessRate
		totalDuration += run.Duration

		if run.MaxThroughput > pc.MaxThroughput {
			pc.MaxThroughput = run.MaxThroughput
		}
		if run.MinThroughput < pc.MinThroughput {
			pc.MinThroughput = run.MinThroughput
		}
	}

	pc.AvgThroughput = totalThroughput / float64(len(pc.Runs))
	pc.SuccessRate = totalSuccessRate / float64(len(pc.Runs))
	pc.Duration = totalDuration / time.Duration(len(pc.Runs))

	// Calculate variance
	var sumSquaredDiffs float64
	for _, run := range pc.Runs {
		diff := run.AvgThroughput - pc.AvgThroughput
		sumSquaredDiffs += diff * diff
	}
	pc.ThroughputVariance = sumSquaredDiffs / float64(len(pc.Runs))
}

func generateComparisonTable(comparisons map[string]*ProfileComparison) {
	// Convert to slice and sort by max throughput (descending)
	var sortedComparisons []*ProfileComparison
	for _, comp := range comparisons {
		sortedComparisons = append(sortedComparisons, comp)
	}

	sort.Slice(sortedComparisons, func(i, j int) bool {
		return sortedComparisons[i].MaxThroughput > sortedComparisons[j].MaxThroughput
	})

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "SEQUENCER COMMIT\tSEQ SHORT\tBLOB STORAGE COMMIT\tBLOB SHORT\tPROFILE TYPE\tBLOB SIZE\tMAX THROUGHPUT (MiB/s)\tMIN THROUGHPUT (MiB/s)\tAVG THROUGHPUT (MiB/s)\tVARIANCE\tSUCCESS RATE (%)\tDURATION\tRUNS")
	fmt.Fprintln(w, "---------------\t---------\t-------------------\t----------\t-----------\t---------\t-------------------\t-------------------\t-------------------\t--------\t---------------\t--------\t----")

	// Print data rows
	for _, comp := range sortedComparisons {
		// Safely truncate commit hashes
		sequencerCommit := comp.CommitHash
		if len(sequencerCommit) > 8 {
			sequencerCommit = sequencerCommit[:8]
		}
		blobStorageCommit := comp.BlobStorageCommitHash
		if len(blobStorageCommit) > 8 {
			blobStorageCommit = blobStorageCommit[:8]
		}
		if blobStorageCommit == "" {
			blobStorageCommit = "unknown"
		}

		// Format blob size for display
		blobSize := comp.BlobSize
		if blobSize == "" {
			blobSize = "unknown"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.1f\t%s\t%d\n",
			sequencerCommit,
			comp.CommitHashShort,
			blobStorageCommit,
			comp.BlobStorageCommitHashShort,
			comp.ProfileType,
			blobSize,
			comp.MaxThroughput,
			comp.MinThroughput,
			comp.AvgThroughput,
			comp.ThroughputVariance,
			comp.SuccessRate,
			formatDuration(comp.Duration),
			comp.RunCount,
		)
	}

	// Print summary statistics
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "SUMMARY:")
	fmt.Fprintf(w, "Total comparisons analyzed: %d\n", len(comparisons))

	// Count unique commits
	commits := make(map[string]bool)
	for _, comp := range comparisons {
		commits[comp.CommitHash] = true
	}
	fmt.Fprintf(w, "Unique commits: %d\n", len(commits))

	// Count unique profile types
	profileTypes := make(map[string]bool)
	for _, comp := range comparisons {
		profileTypes[comp.ProfileType] = true
	}
	fmt.Fprintf(w, "Profile types found: %s\n", strings.Join(getKeys(profileTypes), ", "))
}

func generateCSVOutput(comparisons map[string]*ProfileComparison) {
	fmt.Println("SEQUENCER_COMMIT,SEQ_SHORT,BLOB_STORAGE_COMMIT,BLOB_SHORT,PROFILE_TYPE,BLOB_SIZE,MAX_THROUGHPUT_MIB_PER_SEC,MIN_THROUGHPUT_MIB_PER_SEC,AVG_THROUGHPUT_MIB_PER_SEC,VARIANCE,SUCCESS_RATE_PERCENT,DURATION,RUNS")

	// Convert to slice and sort by max throughput (descending)
	var sortedComparisons []*ProfileComparison
	for _, comp := range comparisons {
		sortedComparisons = append(sortedComparisons, comp)
	}

	sort.Slice(sortedComparisons, func(i, j int) bool {
		return sortedComparisons[i].MaxThroughput > sortedComparisons[j].MaxThroughput
	})

	// Print data rows
	for _, comp := range sortedComparisons {
		// Safely truncate commit hashes
		sequencerCommit := comp.CommitHash
		if len(sequencerCommit) > 8 {
			sequencerCommit = sequencerCommit[:8]
		}
		blobStorageCommit := comp.BlobStorageCommitHash
		if len(blobStorageCommit) > 8 {
			blobStorageCommit = blobStorageCommit[:8]
		}
		if blobStorageCommit == "" {
			blobStorageCommit = "unknown"
		}

		// Format blob size for display
		blobSize := comp.BlobSize
		if blobSize == "" {
			blobSize = "unknown"
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.1f,%s,%d\n",
			sequencerCommit,
			comp.CommitHashShort,
			blobStorageCommit,
			comp.BlobStorageCommitHashShort,
			comp.ProfileType,
			blobSize,
			comp.MaxThroughput,
			comp.MinThroughput,
			comp.AvgThroughput,
			comp.ThroughputVariance,
			comp.SuccessRate,
			formatDuration(comp.Duration),
			comp.RunCount,
		)
	}
}

func generateIndividualTableOutput(runs []ProfileRun) {
	// Sort runs by max throughput (descending)
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].MaxThroughput > runs[j].MaxThroughput
	})

	// Create tabwriter for formatted output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	fmt.Fprintln(w, "SEQUENCER COMMIT\tSEQ SHORT\tBLOB STORAGE COMMIT\tBLOB SHORT\tPROFILE TYPE\tBLOB SIZE\tMAX THROUGHPUT (MiB/s)\tMIN THROUGHPUT (MiB/s)\tAVG THROUGHPUT (MiB/s)\tVARIANCE\tSUCCESS RATE (%)\tDURATION\tTIMESTAMP")
	fmt.Fprintln(w, "---------------\t---------\t-------------------\t----------\t-----------\t---------\t-------------------\t-------------------\t-------------------\t--------\t---------------\t--------\t---------")

	// Print individual runs
	for _, run := range runs {
		// Safely truncate commit hashes
		sequencerCommit := run.CommitHash
		if len(sequencerCommit) > 8 {
			sequencerCommit = sequencerCommit[:8]
		}
		blobStorageCommit := run.BlobStorageCommitHash
		if len(blobStorageCommit) > 8 {
			blobStorageCommit = blobStorageCommit[:8]
		}
		if blobStorageCommit == "" {
			blobStorageCommit = "unknown"
		}

		// Format blob size for display
		blobSize := run.BlobSize
		if blobSize == "" {
			blobSize = "unknown"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%.2f\t%.2f\t%.2f\t%.2f\t%.1f\t%s\t%s\n",
			sequencerCommit,
			run.CommitHashShort,
			blobStorageCommit,
			run.BlobStorageCommitHashShort,
			run.ProfileType,
			blobSize,
			run.MaxThroughput,
			run.MinThroughput,
			run.AvgThroughput,
			run.ThroughputVariance,
			run.SuccessRate,
			formatDuration(run.Duration),
			run.Timestamp.Format("2006-01-02 15:04:05"),
		)
	}

	// Print summary statistics
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "SUMMARY:")
	fmt.Fprintf(w, "Total runs analyzed: %d\n", len(runs))

	// Count unique commits
	commits := make(map[string]bool)
	for _, run := range runs {
		commits[run.CommitHash] = true
	}
	fmt.Fprintf(w, "Unique commits: %d\n", len(commits))

	// Count unique profile types
	profileTypes := make(map[string]bool)
	for _, run := range runs {
		profileTypes[run.ProfileType] = true
	}
	fmt.Fprintf(w, "Profile types found: %s\n", strings.Join(getKeys(profileTypes), ", "))
}

func generateIndividualCSVOutput(runs []ProfileRun) {
	fmt.Println("SEQUENCER_COMMIT,SEQ_SHORT,BLOB_STORAGE_COMMIT,BLOB_SHORT,PROFILE_TYPE,BLOB_SIZE,MAX_THROUGHPUT_MIB_PER_SEC,MIN_THROUGHPUT_MIB_PER_SEC,AVG_THROUGHPUT_MIB_PER_SEC,VARIANCE,SUCCESS_RATE_PERCENT,DURATION,TIMESTAMP")

	// Sort runs by max throughput (descending)
	sort.Slice(runs, func(i, j int) bool {
		return runs[i].MaxThroughput > runs[j].MaxThroughput
	})

	// Print individual runs
	for _, run := range runs {
		// Safely truncate commit hashes
		sequencerCommit := run.CommitHash
		if len(sequencerCommit) > 8 {
			sequencerCommit = sequencerCommit[:8]
		}
		blobStorageCommit := run.BlobStorageCommitHash
		if len(blobStorageCommit) > 8 {
			blobStorageCommit = blobStorageCommit[:8]
		}
		if blobStorageCommit == "" {
			blobStorageCommit = "unknown"
		}

		// Format blob size for display
		blobSize := run.BlobSize
		if blobSize == "" {
			blobSize = "unknown"
		}

		fmt.Printf("%s,%s,%s,%s,%s,%s,%.2f,%.2f,%.2f,%.2f,%.1f,%s,%s\n",
			sequencerCommit,
			run.CommitHashShort,
			blobStorageCommit,
			run.BlobStorageCommitHashShort,
			run.ProfileType,
			blobSize,
			run.MaxThroughput,
			run.MinThroughput,
			run.AvgThroughput,
			run.ThroughputVariance,
			run.SuccessRate,
			formatDuration(run.Duration),
			run.Timestamp.Format("2006-01-02 15:04:05"),
		)
	}
}

func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	} else if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.1fh", d.Hours())
	}
}
