package fullcluster

import (
	"log"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/runner"
)

func Setup() {
	// Default configuration path relative to the cluster directory
	configPath := filepath.Join("e2e", "cluster", "cluster.toml")
	if err := SetupWithConfig(configPath); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
}

// SetupWithConfig allows specifying a custom configuration file path
func SetupWithConfig(configPath string) error {
	return runner.Setup(configPath)
}
