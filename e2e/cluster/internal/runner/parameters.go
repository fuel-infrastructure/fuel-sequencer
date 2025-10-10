// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

// TODO: This is a temporary file to store parameters for the cluster.
// It should be replaced with a proper configuration management system in the future.
// In the meantime, primary parameters to configure are:
// - makefileDir: The directory where the makefile is located
// - destinations: The list of destinations to deploy to
//
// The rest are applied remotely or intended to be consistent across a deployment. Ultimately depends on the usecase.

import (
	"path/filepath"
)

const (
	// Binary Parameters
	makefileDir      = "/home/user/fuel-sequencer"                                 // Absolute path to the directory where makefile is located
	wantArch         = "linux-amd64"                                               // Arch specified from build binary suffix
	buildPath        = makefileDir + "/build"                                      // Path where binary will be built
	dataDir          = makefileDir + "/e2e/cluster/data"                           // Directory with template data
	servicePath      = makefileDir + "/e2e/cluster/systemd/fuelsequencerd.service" // Path to the systemd service file
	systemdPath      = "/etc/systemd/system/fuelsequencerd.service"                // Path to the systemd service file on the remote machine
	remoteBinaryName = "fuelsequencerd"                                            // Name of the binary on remote

	// Blob Parameters
	blobComposePath = makefileDir + "/e2e/cluster/blob/docker-compose.blobpool.yml" // Path to the docker compose file
	blobRedisPath   = makefileDir + "/e2e/cluster/blob/redis.conf"                  // Path to the Redis configuration file
)

var (
	// defaultOptions are the intended options for each system, which may be overridden as needed
	defaultOptions = options{
		sequencer: true,
		blobpool:  true,
	}

	// Remote System Parameters
	systems = []system{
		{
			destination: destination{
				peer_ip: "127.0.0.1",
				host:    "localhost",
				user:    "benchmarks",
				pass:    "password",
				dir:     "/home/benchmarks",
			},
			options: defaultOptions,
		},
	}
)

// chainHomeDir returns the path to the chain's home directory on the remote host
func chainHomeDir(d destination) string {
	return filepath.Join(d.dir, ".fuelsequencer")
}

// remoteBinaryPath returns the path where the binary should be installed on the remote host
func remoteBinaryPath(d destination) string {
	return filepath.Join(d.dir, remoteBinaryName)
}

// remoteBlobDir returns the path where blob files should be stored on the remote host
func remoteBlobDir(d destination) string {
	return filepath.Join(d.dir, "blob")
}

// remoteBlobpoolComposePath returns the path where the docker compose file should be stored on the remote host
func remoteBlobpoolComposePath(d destination) string {
	return filepath.Join(remoteBlobDir(d), "docker-compose.blobpool.yml")
}

// remoteBlobStorageRedisConfPath returns the path where the Redis config file should be stored on the remote host
func remoteBlobStorageRedisConfPath(d destination) string {
	return filepath.Join(remoteBlobDir(d), "redis.conf")
}
