package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	"github.com/fuel-infrastructure/blob-storage/pkg/size"
)

var (
	// constant intends to keep a fixed throughput rate
	constant = ProfileRate{
		Rate:    func(_ time.Duration) int64 { return 100 * size.KiB }, // same as StartRate
		MaxRate: size.MiB,
	}

	// linear intends a monotonic but fixed increase in the throughput
	_linear = ProfileRate{
		Rate: func(duration time.Duration) int64 {
			// every second, the rate increases by 100 KiB
			ns := float32(duration.Nanoseconds())
			s := ns / float32(time.Second)
			rate := (1 + s) * 100 * size.KiB
			return int64(rate)
		},
		MaxRate: 200 * size.MiB,
	}

	// todo: exponential
)

// defaultSetup returns a linear rate configuration
func defaultSetup(
	logger *slog.Logger,
	profileRate ProfileRate,
	distribution blobgen.BlobSizeDistribution) *Config {
	cfg := &Config{
		BlobhubURL:       "http://localhost:31035",
		SequencerGRPC:    "localhost:9090",
		ProfileRate:      profileRate,
		BlobDistribution: distribution,
		MaxLatency:       25 * time.Second,
		Topic:            "test-topic",
		Sender:           "test-sender",
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
