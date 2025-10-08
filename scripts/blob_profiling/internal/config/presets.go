package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/sequencer"
	blobkeeper "github.com/fuel-infrastructure/fuel-sequencer/x/blob/keeper"
)

func DefaultConfig(logger *slog.Logger) *Config {
	// return defaultSetup(logger, constant, blobgen.Fixed100KiB)
	return defaultSetup(logger, linear, blobgen.Fixed100KiB)
}

// GetProfile returns a specific profile by type
func GetProfile(profileType string) Profile {
	switch profileType {
	case "constant":
		return constant
	case "linear":
		return linear
	default:
		return linear // default fallback
	}
}

// GetAvailableProfiles returns a list of available profile types
func GetAvailableProfiles() []string {
	return []string{"constant", "linear"}
}

const (
	defaultDuration = 10 * time.Minute
	targetRate      = 1 * size.GB / 6

	// constant configurable presets
	constantRate = 0.5 * size.MiB // Stay fixed at this rate

	// linear configurable presets
	linearIncrementRate = 5 * size.KiB // At every second, add this amount of throughput
)

var (
	// constant intends to keep a fixed throughput rate
	constant = Profile{
		Description: fmt.Sprintf("constant_%.0f_kib_s_for_%.0f_secs",
			constantRate/size.KiB,
			defaultDuration.Seconds()),
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate:     func(_ time.Duration) float64 { return constantRate }, // constant rate
		Type:     "constant",
		Purpose:  fmt.Sprintf("Constant demand of %.1f KiB/s", constantRate/size.KiB),
	}

	// linear intends a monotonic but fixed increase in the throughput
	linear = Profile{
		Description: fmt.Sprintf("linear_%d_kib_per_s_for_%.0f_secs",
			linearIncrementRate/size.KiB,
			defaultDuration.Seconds()),
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate: func(duration time.Duration) float64 {
			return (1 + duration.Seconds()) * linearIncrementRate
		},
		Type:    "linear",
		Purpose: fmt.Sprintf("Demanding an extra +%d KiB/s, per second", linearIncrementRate/size.KiB),
	}
)

func localSetup() *Config {
	return &Config{
		BlobhubURL:   "http://" + blobkeeper.LocalIP + blobkeeper.BlobhubPort,
		SequencerRPC: "http://" + blobkeeper.LocalIP + ":26657",
		BlobpoolURL:  "http://" + blobkeeper.LocalIP + blobkeeper.BlobpoolAddress,
	}
}

func benchnetSetup() *Config {
	return &Config{
		BlobhubURL:   "http://" + blobkeeper.BenchnetEUIP + blobkeeper.BlobhubPort,
		SequencerRPC: "http://" + blobkeeper.BenchnetEUIP + ":26657",
		BlobpoolURL:  "http://" + blobkeeper.BenchnetEUIP + blobkeeper.BlobpoolAddress,
	}
}

// defaultSetup returns a linear rate configuration
func defaultSetup(
	logger *slog.Logger,
	profile Profile,
	distribution blobgen.BlobSizeDistribution) *Config {

	cfg := localSetup()
	// cfg := benchnetSetup()

	cfg.Profile = profile
	cfg.ParquetDir = "../../output" // root of ./cmd/blob_profiler
	cfg.MaxLagRatio = 1.2
	cfg.LagTolerance = 2 * sequencer.BlockTime
	cfg.BufferDuration = 1 * sequencer.BlockTime
	cfg.Topic = "test-topic"
	cfg.Sender = "alice"
	cfg.SetBlobDistribution(distribution)

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	return cfg
}
