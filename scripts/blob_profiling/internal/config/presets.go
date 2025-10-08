package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/sequencer"
	blobkeeper "github.com/fuel-infrastructure/fuel-sequencer/x/blob/keeper"
)

const (
	defaultDuration = 1 * time.Minute
	targetRate      = 1 * size.GB / 6
)

// GetProfile returns a specific profile by type
func GetProfile(logger *slog.Logger, profileType string) Profile {
	var p Profile

	switch profileType {
	case "constant":
		p = constant(0.5 * size.MiB)
	case "perblock":
		p = perblock(1, 1*size.MiB)
	case "linear":
		p = linear(5 * size.KiB)
	default:
		panic("no profile defined")
	}

	// if NoBlobSize == -1 {
	// 	logger.Info("no blob size specified - using realistic distribution")
	// 	p.BlobDistribution = blobgen.RealisticDistribution
	// }

	return p
}

// GetAvailableProfiles returns a list of available profile types
func GetAvailableProfiles() []string {
	return []string{"constant", "perblock", "linear"}
}

func DefaultConfig(logger *slog.Logger, profileType string) *Config {
	profile := GetProfile(logger, profileType)
	logger.Info("Using profile", "type", profile.Type, "purpose", profile.Purpose)
	return defaultSetup(logger, profile)
}

var (
	// constant intends to keep a fixed throughput rate
	constant = func(fixedRate float64) Profile {
		return Profile{
			Description: fmt.Sprintf("constant_%.0f_kib_s_for_%.0f_secs",
				fixedRate/size.KiB,
				defaultDuration.Seconds()),
			Duration: defaultDuration,
			MaxRate:  targetRate,
			Rate:     func(_ time.Duration) float64 { return fixedRate }, // constant rate
			Type:     "constant",
			Purpose:  fmt.Sprintf("Constant demand of %.1f KiB/s", fixedRate/size.KiB),
		}
	}

	// per block intends to keep a fixed throughput rate
	// the same as constant, but defined per <blocktime> seconds, instead of just 1.
	// tldr: sugared constant preset
	perblock = func(blobCount, fixedBlobSize int) Profile {
		return Profile{
			Description: fmt.Sprintf("%d_x_%.0f_KiB_blobs_per_block", blobCount, float64(fixedBlobSize/size.KiB)),
			Duration:    defaultDuration,
			BlobSize:    fixedBlobSize,
			MaxRate:     targetRate,
			Rate: func(_ time.Duration) float64 {
				return float64(blobCount*fixedBlobSize) / sequencer.BlockTime.Seconds()
			},
			Type:    "perblock",
			Purpose: fmt.Sprintf("Posting %dx%.0fKiB Blobs at each Block", blobCount, float64(fixedBlobSize/size.KiB)),
		}
	}

	// linear intends a monotonic but fixed increase in the throughput
	linear = func(incrementRate float64) Profile {
		return Profile{
			Description: fmt.Sprintf("linear_%d_kib_per_s_for_%.0f_secs",
				int(incrementRate/size.KiB),
				defaultDuration.Seconds()),
			Duration: defaultDuration,
			MaxRate:  targetRate,
			Rate: func(duration time.Duration) float64 {
				return (1 + duration.Seconds()) * incrementRate
			},
			Type:    "linear",
			Purpose: fmt.Sprintf("Demanding an extra +%0.0f KiB/s, per second", incrementRate/size.KiB),
		}
	}
)

func localSetup() *Config {
	return &Config{
		ParquetDir:   "../../localhost_output", // root of ./cmd/blob_profiler
		ChainID:      "fuelsequencer-1",
		BlobhubURL:   "http://" + blobkeeper.LocalIP + blobkeeper.BlobhubPort,
		SequencerRPC: "http://" + blobkeeper.LocalIP + ":26657",
		BlobpoolURL:  "http://" + blobkeeper.LocalIP + blobkeeper.BlobpoolAddress,
	}
}

func benchnetSetup() *Config {
	return &Config{
		ParquetDir:   "../../benchnet_eu_output", // root of ./cmd/blob_profiler
		ChainID:      "seq-benchnet-1",
		BlobhubURL:   "http://" + blobkeeper.BenchnetEUIP + blobkeeper.BlobhubPort,
		SequencerRPC: "http://" + blobkeeper.BenchnetEUIP + ":26657",
		BlobpoolURL:  "http://" + blobkeeper.BenchnetEUIP + blobkeeper.BlobpoolAddress,
	}
}

// defaultSetup returns a linear rate configuration
func defaultSetup(logger *slog.Logger, profile Profile) *Config {
	cfg := localSetup()
	// cfg := benchnetSetup()

	cfg.Profile = profile
	cfg.MaxLagRatio = 1.2
	cfg.LagTolerance = 2 * sequencer.BlockTime
	cfg.BufferDuration = 1 * sequencer.BlockTime
	cfg.Topic = "test-topic"
	cfg.Sender = "alice"

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	return cfg
}
