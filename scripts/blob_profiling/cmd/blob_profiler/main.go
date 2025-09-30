// Blob Profiler - Comprehensive blob profiling system with integrated visualization
//
// This tool profiles blob processing performance and generates visualization graphs.
// It can run in two modes:
//  1. Full profiling: Connects to blobhub, sequencer, and blobpool services to profile real blob processing
//  2. Visualization-only: Generates graphs from existing parquet data files
//
// Command Line Flags:
//
//	-blobhub: Override blobhub URL (default: http://localhost:31035)
//	-sequencer: Override sequencer RPC URL (default: http://localhost:26657)
//	-blobpool: Override blobpool URL (default: http://localhost:31036)
//	-parquet: Output directory for parquet files (default: ./)
//	-profile: Profile type: constant, linear, or exponential (default: linear)
//	-graphs: Generate visualization graphs (default: true)
//	-cleanup: Delete intermediate data files after graph generation (default: true)
//
// Usage Examples:
//
//	# Full profiling with visualization (default linear profile)
//	./blob_profiler -graphs -cleanup
//
//	# Constant load testing
//	./blob_profiler -profile constant -graphs -cleanup
//
//	# Exponential stress testing
//	./blob_profiler -profile exponential -graphs -cleanup
//
//	# Visualization only from existing data
//	./blob_profiler -graphs -parquet cmd/blob_profiler
//
//	# Makefile shortcuts
//	make generate-graphs    # Generate graphs from existing parquet files
//	make generate-clean     # Generate graphs and cleanup data files
//
// Generated Files:
//   - Parquet data: {parquet_dir}/{profile_description}/{timestamp}/blobs.parquet, throughput.parquet
//   - Visualization: {parquet_dir}/{profile_description}/{timestamp}/graphs/images/*.png (5 graphs: throughput, blob sizes, timeline, store-to-blobpool, store-to-finalized)
//
// Prerequisites:
//   - DuckDB: For querying parquet files (brew install duckdb / apt install duckdb)
//   - gnuplot: For generating graphs (brew install gnuplot / apt install gnuplot)
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/parquet"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/profiler"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/report"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/visualizer"
)

func main() {
	// Command line flags
	var (
		blobhubURL     = flag.String("blobhub", "", "Override blobhub URL")
		sequencerRPC   = flag.String("sequencer", "", "Override sequencer RPC URL")
		blobpoolURL    = flag.String("blobpool", "", "Override blobpool URL")
		parquetDir     = flag.String("parquet", "", "Output parquet dir path for profiling data")
		profileType    = flag.String("profile", "", "Profile type: constant, linear, or exponential (default: linear)")
		generateGraphs = flag.Bool("graphs", true, "Generate visualization graphs after profiling")
		cleanupData    = flag.Bool("cleanup", true, "Delete generated data files after creating images")
		visualizeOnly  = flag.Bool("visualize-only", false, "Only generate graphs from existing parquet data, skip profiling")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}))

	// Load configuration
	cfg := config.DefaultConfig(logger)

	// Override with command line flags if provided
	if *blobhubURL != "" {
		cfg.BlobhubURL = *blobhubURL
	}
	if *sequencerRPC != "" {
		cfg.SequencerRPC = *sequencerRPC
	}
	if *blobpoolURL != "" {
		cfg.BlobpoolURL = *blobpoolURL
	}
	if *parquetDir != "" {
		cfg.ParquetDir = *parquetDir
	}
	if *profileType != "" {
		cfg.Profile = config.GetProfile(*profileType)
		logger.Info("Using profile", "type", cfg.Profile.Type, "purpose", cfg.Profile.Purpose)
	}

	// Handle visualization-only mode
	if *visualizeOnly {
		if *parquetDir == "" {
			logger.Error("Visualization-only mode requires -parquet flag to specify data directory")
			os.Exit(1)
		}

		logger.Info("Running in visualization-only mode", "parquet_dir", *parquetDir)

		// Load profile from existing data directory
		profile := config.GetProfileFromDir(*parquetDir)
		if profile.Type == "" {
			logger.Error("Could not determine profile type from directory name")
			os.Exit(1)
		}

		// Generate visualization graphs
		logger.Info("Generating visualization graphs...")
		viz := visualizer.New(*parquetDir, profile, *cleanupData)
		if err := viz.GenerateGraphs(); err != nil {
			logger.Error("Failed to generate graphs", "error", err)
			os.Exit(1)
		}
		logger.Info("Visualization graphs generated successfully")
		return
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	runTimestamp := time.Now()
	runDirName := config.RunOutputDir(runTimestamp)

	logger.Info("Blobhub URL", "url", cfg.BlobhubURL)
	logger.Info("Sequencer RPC URL", "url", cfg.SequencerRPC)
	logger.Info("Start Rate", "rate_KiB", cfg.Rate(0)/size.KiB)
	logger.Info("Max Rate", "rate_MiB", cfg.MaxRate/size.MiB)
	logger.Info("Profile Run Timestamp", "timestamp", runTimestamp)

	ctx := context.Background()

	// Create parquet writer if output file is specified
	parquetHandler, err := parquet.New(cfg.ParquetDir, runDirName, logger, cfg.Profile.Description)
	if err != nil {
		logger.Error("failed to create parquet writer - will exit", "error", err)
		os.Exit(1)
	}

	// Create profiler report
	profilerReport, err := report.New(cfg.ParquetDir, runDirName, cfg.Profile.Description, runTimestamp)
	if err != nil {
		logger.Error("failed to create profiler report - will exit", "error", err)
		os.Exit(1)
	}
	// Initialize profiler report with environment information
	profilerReport.InitialiseEnvironment(cfg)

	// Generate visualization graphs if requested - defer now to ensure parquet handler is closed
	if *generateGraphs {
		defer func() {
			logger.Info("Generating visualization graphs...")
			viz := visualizer.New(parquetHandler.Dir(), cfg.Profile, *cleanupData)
			if err := viz.GenerateGraphs(); err != nil {
				logger.Error("Failed to generate graphs", "error", err)
				os.Exit(1)
			}
			logger.Info("Visualization graphs generated successfully")
		}()
	}
	defer parquetHandler.Close()

	// Create profiler
	profiler, err := profiler.NewBlobProfiler(ctx, cfg, logger, parquetHandler, profilerReport)
	if err != nil {
		logger.Error("Failed to create blob profiler - will exit", "error", err)
		os.Exit(1)
	}
	defer profiler.Close()

	// Run profiling
	blobs, err := profiler.RunProfile(ctx)
	if err != nil {
		logger.Error("Profiling failed - will exit", "error", err)
		os.Exit(1)
	}

	// Write profiler report
	if err := profiler.WriteReport(); err != nil {
		logger.Error("Failed to write profiler report", "error", err)
	} else {
		logger.Info("Profiler report written", "path", profiler.ReportPath())
	}

	// Log final statistics
	logger.Info("Profiling completed", "total_blobs", len(blobs))
	if parquetHandler != nil {
		stats := parquetHandler.GetStats()
		logger.Info("Parquet writer stats",
			"buffered_records", stats.BufferedRecords,
			"batch_size", stats.BatchSize,
			"profile_duration", stats.ProfileDuration)
	}
}
