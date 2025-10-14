package runner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/setup"
)

func TestSetupLoadsConfig(t *testing.T) {
	// Create a temporary config file for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.toml")

	// Write test configuration
	testConfig := `[binary]
makefile_dir = "/test/path"
want_arch = "linux-amd64"
remote_binary_name = "test-binary"
systemd_path = "/test/systemd/path"

[blob]
redis_path = "test/redis.conf"
pool_compose_path = "test/pool-compose.yml"
hub_compose_path = "test/hub-compose.yml"

[[systems]]
[systems.destination]
peer_ip = "127.0.0.1"
host = "localhost"
user = "testuser"
pass = "testpass"
dir = "/test/dir"

[systems.options]
sequencer = true
blob_pool = true
blob_hub = false
`

	err := os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Test that LoadConfig works
	err = setup.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify that Systems() now returns the loaded systems
	systems := setup.Systems()
	if len(systems) != 1 {
		t.Errorf("Expected 1 system, got %d", len(systems))
	}

	if len(systems) > 0 {
		system := systems[0]
		if system.Destination.PeerIP != "127.0.0.1" {
			t.Errorf("Expected PeerIP to be '127.0.0.1', got '%s'", system.Destination.PeerIP)
		}
		if system.Destination.Host != "localhost" {
			t.Errorf("Expected Host to be 'localhost', got '%s'", system.Destination.Host)
		}
		if !system.Options.Sequencer {
			t.Error("Expected Sequencer to be true")
		}
	}
}
