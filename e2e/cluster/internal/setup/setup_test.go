package setup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/setup"
)

func TestLoadConfig(t *testing.T) {
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

	// Load the configuration
	err = setup.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the configuration was loaded correctly
	config := setup.LoadedConfig()
	if config == nil {
		t.Fatal("Config is nil")
	}

	// Test binary configuration
	binaryConfig := setup.BinaryConfig()
	if binaryConfig.MakefileDir != "/test/path" {
		t.Errorf("Expected MakefileDir to be '/test/path', got '%s'", binaryConfig.MakefileDir)
	}
	if binaryConfig.WantArch != "linux-amd64" {
		t.Errorf("Expected WantArch to be 'linux-amd64', got '%s'", binaryConfig.WantArch)
	}
	if binaryConfig.RemoteBinaryName != "test-binary" {
		t.Errorf("Expected RemoteBinaryName to be 'test-binary', got '%s'", binaryConfig.RemoteBinaryName)
	}

	// Test blob configuration
	blobConfig := setup.BlobConfig()
	if blobConfig.RedisPath != "test/redis.conf" {
		t.Errorf("Expected RedisPath to be 'test/redis.conf', got '%s'", blobConfig.RedisPath)
	}

	// Test systems configuration
	systems := setup.SystemConfigs()
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
		if !system.Options.Blobpool {
			t.Error("Expected BlobPool to be true")
		}
		if system.Options.Blobhub {
			t.Error("Expected BlobHub to be false")
		}
	}

	// Test systems with connections
	systemsWithConnections := setup.Systems()
	if len(systemsWithConnections) != 1 {
		t.Errorf("Expected 1 system with connections, got %d", len(systemsWithConnections))
	}
}

func TestHelperFunctions(t *testing.T) {
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

	// Load the configuration
	err = setup.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test helper functions
	expectedBuildPath := filepath.Join("/test/path", "build")
	if setup.BuildPath() != expectedBuildPath {
		t.Errorf("Expected BuildPath to be '%s', got '%s'", expectedBuildPath, setup.BuildPath())
	}

	expectedDataDir := filepath.Join("/test/path", "e2e/cluster/data")
	if setup.DataDir() != expectedDataDir {
		t.Errorf("Expected DataDir to be '%s', got '%s'", expectedDataDir, setup.DataDir())
	}

	expectedServicePath := filepath.Join("/test/path", "e2e/cluster/systemd/fuelsequencerd.service")
	if setup.ServicePath() != expectedServicePath {
		t.Errorf("Expected ServicePath to be '%s', got '%s'", expectedServicePath, setup.ServicePath())
	}

	expectedBlobRedisPath := filepath.Join("/test/path", "test/redis.conf")
	if setup.BlobRedisPath() != expectedBlobRedisPath {
		t.Errorf("Expected BlobRedisPath to be '%s', got '%s'", expectedBlobRedisPath, setup.BlobRedisPath())
	}

	expectedBlobpoolComposePath := filepath.Join("/test/path", "test/pool-compose.yml")
	if setup.BlobpoolComposePath() != expectedBlobpoolComposePath {
		t.Errorf("Expected BlobpoolComposePath to be '%s', got '%s'", expectedBlobpoolComposePath, setup.BlobpoolComposePath())
	}

	expectedBlobhubDirPath := filepath.Join("/test/path", "test/hub-compose.yml")
	if setup.BlobhubDirPath() != expectedBlobhubDirPath {
		t.Errorf("Expected BlobhubDirPath to be '%s', got '%s'", expectedBlobhubDirPath, setup.BlobhubDirPath())
	}
}
