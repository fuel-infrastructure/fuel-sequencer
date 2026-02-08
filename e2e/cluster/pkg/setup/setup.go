package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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
	SequencerVersion string `mapstructure:"sequencer_version"` // "current", branch name, or commit hash; empty = same as "current"
}

// Blob holds blob-related configuration
type Blob struct {
	RedisPath           string `mapstructure:"redis_path"`
	BlobpoolComposePath string `mapstructure:"blobpool_compose_path"`
	BlobhubDirPath      string `mapstructure:"blobhub_dir_path"`
	BlobStorageVersion  string `mapstructure:"blob_storage_version"` // "current", branch name, or commit hash; empty = same as "current"
}

// System holds system configuration and runtime state
type System struct {
	Destination `mapstructure:"destination"`
	Options     `mapstructure:"options"`
	SSH         *ssh.Client `mapstructure:"-"` // Don't serialize SSH connection
	IsLocal     bool        `mapstructure:"-"` // Whether the system is local (localhost, 127.0.0.1, or ::1)
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
	Instances int  `mapstructure:"instances"` // Number of instances to deploy on this system (default: 1)
}

var (
	// config holds the loaded configuration
	config *Config
	// systems holds the converted system configurations with SSH connections
	systems []System
	// resolved paths when building from a specific version (temp copy); empty means use config paths
	resolvedMakefileDir string
	resolvedBlobhubDir  string
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
		// Default to 1 instance if not specified
		if sysConfig.Options.Instances <= 0 {
			sysConfig.Options.Instances = 1
		}
		systems[i] = System{
			Destination: sysConfig.Destination,
			Options:     sysConfig.Options,
			SSH:         nil, // Will be set when connections are established
		}
	}

	return nil
}

// LoadedConfig returns the loaded configuration
func LoadedConfig() *Config {
	return config
}

// BinaryConfig returns the binary configuration
func BinaryConfig() Binary {
	return config.Binary
}

// SetResolvedPaths sets the effective directories for sequencer and blob-storage builds (e.g. temp copy for a version).
// Empty string means use the path from config. Cleared when temp dirs are removed.
func SetResolvedPaths(sequencerDir, blobDir string) {
	resolvedMakefileDir = sequencerDir
	resolvedBlobhubDir = blobDir
}

// ClearResolvedSequencerDir clears the resolved sequencer path (used when removing its temp dir).
func ClearResolvedSequencerDir() { resolvedMakefileDir = "" }

// ClearResolvedBlobhubDir clears the resolved blob path (used when removing its temp dir).
func ClearResolvedBlobhubDir() { resolvedBlobhubDir = "" }

// EffectiveMakefileDir returns the directory to use for building the sequencer (resolved temp or config path).
func EffectiveMakefileDir() string {
	if resolvedMakefileDir != "" {
		return resolvedMakefileDir
	}
	return config.Binary.MakefileDir
}

// EffectiveBlobhubDirPath returns the directory to use for blobhub (resolved temp or config path).
func EffectiveBlobhubDirPath() string {
	if resolvedBlobhubDir != "" {
		return resolvedBlobhubDir
	}
	return config.Blob.BlobhubDirPath
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
	return filepath.Join(EffectiveMakefileDir(), "build")
}

// DataDir returns the directory with template data
func DataDir() string {
	return filepath.Join(EffectiveMakefileDir(), "e2e/cluster/data")
}

// ServicePath returns the path to the systemd service file
func ServicePath() string {
	return filepath.Join(EffectiveMakefileDir(), "e2e/cluster/systemd/fuelsequencerd.service")
}

// BlobRedisPath returns the path to the Redis configuration file
func BlobRedisPath() string {
	return filepath.Join(EffectiveMakefileDir(), config.Blob.RedisPath)
}

// BlobpoolComposePath returns the path to the docker compose file for blobpool
func BlobpoolComposePath() string {
	return filepath.Join(EffectiveMakefileDir(), config.Blob.BlobpoolComposePath)
}

// BlobhubDirPath returns the path to the blobhub project directory
func BlobhubDirPath() string {
	return EffectiveBlobhubDirPath()
}

// RemoteChainHomeDir returns the path to the chain's home directory on the remote host
func RemoteChainHomeDir(d Destination, instanceId int) string {
	return filepath.Join(d.Dir, fmt.Sprintf(".fuelsequencer-%d", instanceId))
}

// RemoteBinaryPath returns the path where the binary should be installed on the remote host
func RemoteBinaryPath(d Destination) string {
	return filepath.Join(d.Dir, config.Binary.RemoteBinaryName)
}

// RemoteBlobDir returns the path where blob files should be stored on the remote host
func RemoteBlobDir(d Destination, instanceId int) string {
	return filepath.Join(d.Dir, fmt.Sprintf("blob-%d", instanceId))
}

// RemoteBlobStorageRedisConfPath returns the path where the Redis config file should be stored on the remote host
func RemoteBlobStorageRedisConfPath(d Destination, instanceId int) string {
	return filepath.Join(RemoteBlobDir(d, instanceId), "redis.conf")
}

// RemoteBlobpoolComposePath returns the path where the docker compose file should be stored on the remote host
func RemoteBlobpoolComposePath(d Destination, instanceId int) string {
	return filepath.Join(RemoteBlobDir(d, instanceId), "docker-compose.blobpool.yml")
}

// RemoteBlobhubDir returns the path where the blobhub project is stored
func RemoteBlobhubDir(d Destination, instanceId int) string {
	return filepath.Join(RemoteBlobDir(d, instanceId), "blob-storage")
}

func LocalBlobhubDir() string {
	return EffectiveBlobhubDirPath()
}

// RemoteBlobhubComposeDir returns the path where the blobhub compose is stored
func RemoteBlobhubComposeDir(d Destination, instanceId int) string {
	return filepath.Join(RemoteBlobhubDir(d, instanceId), "docker-compose.yml")
}

func LocalBlobhubComposeDir() string {
	return filepath.Join(LocalBlobhubDir(), "docker-compose.yml")
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
func DockerContainerName(nodeId int, instanceId int) string {
	return fmt.Sprintf("fuelsequencer%d-%d", nodeId, instanceId)
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

// IsLocal checks if a destination is local (localhost, 127.0.0.1, or ::1)
// NOTE: This was only tested on macOS; on Linux `--network host` shares the network, so might need more changes
func IsLocal(dest Destination) bool {
	host := strings.ToLower(dest.Host)
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

// TotalInstances returns the total number of instances across all systems
func TotalInstances(systems []System) int {
	total := 0
	for _, sys := range systems {
		total += sys.Options.Instances
	}
	return total
}

type Ports struct {
	P2P            int
	RPC            int
	API            int
	GRPC           int
	Prometheus     int
	BlobpoolServer int
}

func InstancePorts(instanceId int) Ports {
	return Ports{
		P2P:            26656 + instanceId*100,
		RPC:            26657 + instanceId*100,
		API:            1317 + instanceId*100,
		GRPC:           9090 + instanceId*100,
		Prometheus:     26660 + instanceId*100,
		BlobpoolServer: 21025 + instanceId*10,
	}
}
