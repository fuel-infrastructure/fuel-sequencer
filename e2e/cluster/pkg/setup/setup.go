package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// Config holds the cluster configuration loaded from TOML file
type Config struct {
	Binary  `mapstructure:"binary"`
	Blob    `mapstructure:"blob"`
	Systems []System `mapstructure:"systems"`
}

// Binary holds binary-related configuration
type Binary struct {
	MakefileDir      string `mapstructure:"makefile_dir"`
	WantArch         string `mapstructure:"want_arch"`
	RemoteBinaryName string `mapstructure:"remote_binary_name"`
	SystemdPath      string `mapstructure:"systemd_path"`
}

// Blob holds blob-related configuration
type Blob struct {
	RedisPath           string `mapstructure:"redis_path"`
	BlobpoolComposePath string `mapstructure:"blobpool_compose_path"`
	BlobhubDirPath      string `mapstructure:"blobhub_dir_path"`
}

// System holds system configuration and runtime state
type System struct {
	Destination `mapstructure:"destination"`
	Options     `mapstructure:"options"`
	SSH         *ssh.Client `mapstructure:"-"` // Don't serialize SSH connection
}

// Destination holds destination configuration
type Destination struct {
	PeerIP   string `mapstructure:"peer_ip"`
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Pass     string `mapstructure:"pass"`
	Dir      string `mapstructure:"dir"`
	Platform string `mapstructure:"platform"` // e.g., "linux/amd64", "linux/arm64"
}

// Options holds system options
type Options struct {
	Sequencer bool `mapstructure:"sequencer"`
	Blobpool  bool `mapstructure:"blobpool"`
	Blobhub   bool `mapstructure:"blobhub"`
}

var (
	// config holds the loaded configuration
	config *Config
	// systems holds the converted system configurations with SSH connections
	systems []System
)

// LoadConfig loads the configuration from the specified TOML file
func LoadConfig(configPath string) error {
	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("toml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	config = &Config{}
	if err := v.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Convert SystemConfig to System slice
	systems = make([]System, len(config.Systems))
	for i, sysConfig := range config.Systems {
		systems[i] = System{
			Destination: sysConfig.Destination,
			Options:     sysConfig.Options,
			SSH:         nil, // Will be set when connections are established
		}
	}

	return nil
}

func CheckParameters(logging *zap.SugaredLogger) ([]System, error) {
	systems := Systems()
	ls, lm := len(systems), len(sequencer.Mnemonics)
	if ls != lm {
		err := fmt.Errorf("check config: number of systems (%d) does not match number of mnemonics (%d)", ls, lm)
		logging.Fatalw(err.Error())
		return nil, err
	}

	return systems, nil
}

// LoadedConfig returns the loaded configuration
func LoadedConfig() *Config {
	return config
}

// BinaryConfig returns the binary configuration
func BinaryConfig() Binary {
	return config.Binary
}

// BlobConfig returns the blob configuration
func BlobConfig() Blob {
	return config.Blob
}

// SystemConfigs returns the systems configuration
func SystemConfigs() []System {
	return config.Systems
}

// Systems returns the systems slice with SSH connections
func Systems() []System {
	return systems
}

// BuildPath returns the path where binary will be built
func BuildPath() string {
	return filepath.Join(config.Binary.MakefileDir, "build")
}

// DataDir returns the directory with template data
func DataDir() string {
	return filepath.Join(config.Binary.MakefileDir, "e2e/cluster/data")
}

// ServicePath returns the path to the systemd service file
func ServicePath() string {
	return filepath.Join(config.Binary.MakefileDir, "e2e/cluster/systemd/fuelsequencerd.service")
}

// BlobRedisPath returns the path to the Redis configuration file
func BlobRedisPath() string {
	return filepath.Join(config.Binary.MakefileDir, config.Blob.RedisPath)
}

// BlobpoolComposePath returns the path to the docker compose file for blobpool
func BlobpoolComposePath() string {
	return filepath.Join(config.Binary.MakefileDir, config.Blob.BlobpoolComposePath)
}

// BlobhubDirPath returns the path to the docker compose file for blobhub
func BlobhubDirPath() string {
	return filepath.Join(config.Blob.BlobhubDirPath)
}

// RemoteChainHomeDir returns the path to the chain's home directory on the remote host
func RemoteChainHomeDir(d Destination) string {
	return filepath.Join(d.Dir, ".fuelsequencer")
}

// RemoteBinaryPath returns the path where the binary should be installed on the remote host
func RemoteBinaryPath(d Destination) string {
	return filepath.Join(d.Dir, config.Binary.RemoteBinaryName)
}

// RemoteBlobDir returns the path where blob files should be stored on the remote host
func RemoteBlobDir(d Destination) string {
	return filepath.Join(d.Dir, "blob")
}

// RemoteBlobStorageRedisConfPath returns the path where the Redis config file should be stored on the remote host
func RemoteBlobStorageRedisConfPath(d Destination) string {
	return filepath.Join(RemoteBlobDir(d), "redis.conf")
}

// RemoteBlobpoolComposePath returns the path where the docker compose file should be stored on the remote host
func RemoteBlobpoolComposePath(d Destination) string {
	return filepath.Join(RemoteBlobDir(d), "docker-compose.blobpool.yml")
}

// RemoteBlobhubDir returns the path where the blobhub project is stored
func RemoteBlobhubDir(d Destination) string {
	return filepath.Join(RemoteBlobDir(d), "blob-storage")
}

// RemoteBlobhubComposeDir returns the path where the blobhub compose is stored
func RemoteBlobhubComposeDir(d Destination) string {
	return filepath.Join(RemoteBlobhubDir(d), "docker-compose.yml")
}

// DockerImageName returns the Docker image name for the sequencer
func DockerImageName() string {
	return "fuel-infrastructure/fuel-sequencer"
}

// DockerImageTag returns the Docker image tag (defaults to latest)
// The value may change if the tag is updated by the caller.
var latestDockerImageTag = "latest" // TODO: replace this from global variable
func DockerImageTag(updateTag string) string {
	if updateTag != "" {
		latestDockerImageTag = updateTag
	}
	return latestDockerImageTag
}

func DockerImageTagLatest(d Destination) (string, string) {
	imageName := DockerImageName()
	imageTag := DockerImageTag("")
	platform := DockerPlatform(d)

	// Use platform-specific tag to avoid conflicts between different architectures
	platformTag := fmt.Sprintf("%s-%s", imageTag, platform)
	platformTag = strings.ReplaceAll(platformTag, "/", "-") // Replace / with - for valid tag name
	platformTag = strings.ReplaceAll(platformTag, ".", "_") // Replace . with _ for valid tag name
	fullImageName := fmt.Sprintf("%s:%s", imageName, platformTag)

	return fullImageName, platformTag
}

// DockerContainerName returns the container name for a sequencer instance
func DockerContainerName(nodeId int) string {
	return fmt.Sprintf("fuelsequencer%d", nodeId)
}

// DockerNetworkName returns the Docker network name for sequencer communication
func DockerNetworkName() string {
	return "fuel-sequencer-network"
}

// DockerPlatform returns the Docker platform for a destination, defaulting to linux/amd64 if not specified
func DockerPlatform(d Destination) string {
	if d.Platform != "" {
		return d.Platform
	}
	return "linux/amd64" // Default platform
}
