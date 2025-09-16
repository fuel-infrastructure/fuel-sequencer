package report

import (
	"os"
	"runtime"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
)

// InitialiseEnvironment sets up the environment information
func (pr *ProfilerReport) InitialiseEnvironment(cfg *config.Config) {
	pr.data.Environment.CommitHash = CommitHash()
	pr.data.Environment.CommitHashShort = CommitHashShort()
	pr.data.Environment.GitBranch = Branch()
	pr.data.Environment.GitStatus = Status()
	pr.data.Environment.GoVersion = runtime.Version()
	pr.data.Environment.OS = runtime.GOOS
	pr.data.Environment.Architecture = runtime.GOARCH
	pr.data.Environment.Config = convertToSerializableConfig(cfg) // convert to serializable version

	// Runtime information
	pr.data.Environment.RuntimeInfo = map[string]interface{}{
		"num_cpu":    runtime.NumCPU(),
		"goroutines": runtime.NumGoroutine(),
		"memory_mb":  memoryUsageMB(),
	}

	// Environment variables (filter sensitive ones)
	pr.data.Environment.EnvironmentVars = filteredEnvVars()
}

// Helper functions for environment data

func memoryUsageMB() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int(m.Alloc / 1024 / 1024)
}

func filteredEnvVars() map[string]string {
	envVars := make(map[string]string)

	// List of environment variables to include (avoid sensitive ones)
	safeVars := []string{
		"HOME", "USER", "SHELL", "PATH", "PWD", "LANG", "LC_ALL",
		"GOPATH", "GOROOT", "GOOS", "GOARCH", "GO111MODULE",
		"TERM", "DISPLAY", "SSH_CLIENT", "SSH_CONNECTION",
	}

	for _, key := range safeVars {
		if value := os.Getenv(key); value != "" {
			envVars[key] = value
		}
	}

	return envVars
}

// convertToSerializableConfig converts a config.Config to a SerializableConfig
func convertToSerializableConfig(cfg *config.Config) SerializableConfig {
	return SerializableConfig{
		BlobhubURL:   cfg.BlobhubURL,
		SequencerRPC: cfg.SequencerRPC,
		BlobpoolURL:  cfg.BlobpoolURL,
		ParquetDir:   cfg.ParquetDir,
		Profile: SerializableProfile{
			Description:   cfg.Profile.Description,
			Duration:      cfg.Profile.Duration,
			DurationHuman: formatDuration(cfg.Profile.Duration),
			MaxRate:       cfg.Profile.MaxRate,
			MaxRateHuman:  formatBytesPerSecond(cfg.Profile.MaxRate),
			Type:          cfg.Profile.Type,
			Purpose:       cfg.Profile.Purpose,
			Explanation:   cfg.Profile.Explanation,
			BlobSizeInfo:  cfg.Profile.BlobSizeInfo,
		},
		BlobDistribution:      cfg.BlobDistribution,
		BlobDistributionHuman: config.GenerateBlobSizeInfo(cfg.BlobDistribution),
		MaxLagRatio:           cfg.MaxLagRatio,
		LagTolerance:          cfg.LagTolerance,
		LagToleranceHuman:     formatDuration(cfg.LagTolerance),
		BufferDuration:        cfg.BufferDuration,
		BufferDurationHuman:   formatDuration(cfg.BufferDuration),
		Topic:                 cfg.Topic,
		Sender:                cfg.Sender,
	}
}
