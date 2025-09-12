package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	blobkeeper "github.com/fuel-infrastructure/fuel-sequencer/x/blob/keeper"
)

const (
	nanosPerSecond  = int(time.Second)
	defaultDuration = 1 * time.Minute
	targetRate      = 170 * size.MiB
)

var (
	// constant intends to keep a fixed throughput rate
	constant = Profile{
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate:     func(_ time.Duration) int { return 100 * size.KiB }, // constant rate
		Size: func(x time.Duration) int {
			seconds := int(x / time.Second)
			return 100 * size.KiB * seconds
		},
	}

	// linear intends a monotonic but fixed increase in the throughput
	_linear = Profile{
		Duration: defaultDuration,
		MaxRate:  targetRate,
		Rate: func(duration time.Duration) int {
			// 100*KiB + (100*KiB * nanos / nanosPerSecond)
			// 100*KiB + (100*KiB * x)
			nanos := int(duration)
			baseRate := 100 * size.KiB
			variableRate := (baseRate * nanos) / nanosPerSecond
			return baseRate + variableRate
		},
		Size: func(x time.Duration) int {
			// integral(100*KiB + (100*KiB * x)):
			// 100*KiB * x + 100*KiB * 0.5 * x^2
			// 100*KiB * (x + 0.5 * x^2)
			seconds := x.Seconds()
			baseRate := float64(100 * size.KiB)

			// Do calculation in float64, then convert once at the end
			total := baseRate * (seconds + 0.5*seconds*seconds)
			return int(total)
		},
	}

	// todo: exponential
)

// defaultSetup returns a linear rate configuration
func defaultSetup(
	logger *slog.Logger,
	profile Profile,
	distribution blobgen.BlobSizeDistribution) *Config {
	cfg := &Config{
		BlobhubURL:       "http://localhost:31035",
		SequencerRPC:     "http://localhost:26657",
		BlobpoolURL:      "http://localhost" + blobkeeper.BlobpoolAddress,
		ParquetDir:       "./",
		Profile:          profile,
		BlobDistribution: distribution,
		MaxLagRatio:      1.1,
		LagTolerance:     10 * time.Second,
		BufferDuration:   5 * time.Second,
		Topic:            "test-topic",
		Sender:           "eve",
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("Invalid configuration - will exit", "error", err)
		os.Exit(1)
	}

	return cfg
}

func DefaultConfig(logger *slog.Logger) *Config {
	return defaultSetup(logger, constant, blobgen.Fixed10KiB)
	// return defaultSetup(logger, linear, blobgen.Fixed10KiB)
}
