package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/profiler"
)

func main() {
	// Command line flags
	var (
		blobhubURL = flag.String("blobhub", "", "Override blobhub URL")
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

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	logger.Info("Blobhub URL", "url", cfg.BlobhubURL)
	logger.Info("Start Rate", "rate_KiB", cfg.Rate(0)/size.KiB)
	logger.Info("Max Rate", "rate_MiB", cfg.MaxRate/size.MiB)

	// Create profiler
	profiler, err := profiler.NewBlobProfiler(cfg, logger)
	if err != nil {
		logger.Error("Failed to create blob profiler - will exit", "error", err)
		os.Exit(1)
	}
	defer profiler.Close()

	ctx := context.Background()

	// Run profiling
	blobs, err := profiler.RunProfile(ctx)
	if err != nil {
		logger.Error("Profiling failed - will exit", "error", err)
		os.Exit(1)
	}

	// Analyze the results
	_ = blobs
}
