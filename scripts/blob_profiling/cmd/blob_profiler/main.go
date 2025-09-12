package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/parquet"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/profiler"
)

func main() {
	// Command line flags
	var (
		blobhubURL   = flag.String("blobhub", "", "Override blobhub URL")
		sequencerRPC = flag.String("sequencer", "", "Override sequencer RPC URL")
		blobpoolURL  = flag.String("blobpool", "", "Override blobpool URL")
		parquetDir   = flag.String("parquet", "", "Output parquet dir path for profiling data")
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	logger.Info("Blobhub URL", "url", cfg.BlobhubURL)
	logger.Info("Sequencer RPC URL", "url", cfg.SequencerRPC)
	logger.Info("Start Rate", "rate_KiB", cfg.Rate(0)/size.KiB)
	logger.Info("Max Rate", "rate_MiB", cfg.MaxRate/size.MiB)

	ctx := context.Background()

	// Create parquet writer if output file is specified
	parquetHandler, err := parquet.New(cfg.ParquetDir, logger)
	if err != nil {
		logger.Error("failed to create parquet writer - will exit", "error", err)
		os.Exit(1)
	}
	defer parquetHandler.Close()

	// Create profiler
	profiler, err := profiler.NewBlobProfiler(ctx, cfg, logger, parquetHandler)
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
