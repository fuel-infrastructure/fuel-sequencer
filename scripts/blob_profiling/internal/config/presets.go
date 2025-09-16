package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	"github.com/fuel-infrastructure/blob-storage/pkg/size"
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
	nanosPerSecond  = int(time.Second)
	defaultDuration = 5 * time.Second
	targetRate      = 1 * size.GB / 6

	// constant configurable presets
	constantRate = int(0.5 * size.MiB) // Stay fixed at this rate

	// linear configurable presets
	linearIncrementRate = 5 * size.KiB // At every second, add this amount of throughput
)

var (
	// constant intends to keep a fixed throughput rate
	constant = Profile{
		Description: fmt.Sprintf("constant_%d_kib_s_for_%.0f_secs",
			constantRate/size.KiB,
			defaultDuration.Seconds()),
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate:     func(_ time.Duration) int { return constantRate }, // constant rate
		Size: func(x time.Duration) int {
			seconds := int(x / time.Second)
			return constantRate * seconds
		},
		Type:    "constant",
		Purpose: fmt.Sprintf("Constant demand of %d KiB/s", constantRate/size.KiB),
	}

	// linear intends a monotonic but fixed increase in the throughput
	linear = Profile{
		Description: fmt.Sprintf("linear_%d_kib_per_s_for_%.0f_secs",
			linearIncrementRate/size.KiB,
			defaultDuration.Seconds()),
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate: func(duration time.Duration) int {
			return int((1 + duration.Seconds()) * linearIncrementRate)
		},
		Size: func(x time.Duration) int {
			seconds := x.Seconds()
			// integral of (1 + t) * linearIncrementRate
			// = linearIncrementRate * (t + 0.5*t²)
			total := float64(linearIncrementRate) * (seconds + 0.5*seconds*seconds)
			return int(total)
		},
		Type:    "linear",
		Purpose: fmt.Sprintf("Demanding an extra +%d KiB/s, per second", linearIncrementRate/size.KiB),
	}
)

// defaultSetup returns a linear rate configuration
func defaultSetup(
	logger *slog.Logger,
	profile Profile,
	distribution blobgen.BlobSizeDistribution) *Config {
	cfg := &Config{
		BlobhubURL:     "http://localhost:31035",
		SequencerRPC:   "http://localhost:26657",
		BlobpoolURL:    "http://localhost" + blobkeeper.BlobpoolAddress,
		ParquetDir:     "./",
		Profile:        profile,
		MaxLagRatio:    1.2,
		LagTolerance:   10 * time.Second,
		BufferDuration: 5 * time.Second,
		Topic:          "test-topic",
		Sender:         "eve",
	}

	// Set blob distribution and generate BlobSizeInfo dynamically
	cfg.SetBlobDistribution(distribution)

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	return cfg
}
