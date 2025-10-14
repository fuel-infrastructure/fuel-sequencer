// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
	"go.uber.org/zap"
)

// manageSystems handles the management of all systems including
// service configuration, binary deployment, data management and blob file deployment.
// Returns an error if management of any system fails.
func manageSystems(systems []System, localBinaryPath string) error {
	l := logging.Named("Manage")

	localServiceHash, err := calculateFileHash(ServicePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local service hash: %w", err)
	}

	localBinaryHash, err := calculateFileHash(localBinaryPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get local binary hash: %w", err)
	}

	localBlobRedisHash, err := calculateFileHash(BlobRedisPath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob redis hash: %w", err)
	}

	localBlobpoolComposeHash, err := calculateFileHash(BlobpoolComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob compose hash: %w", err)
	}

	localBlobhubComposeHash, err := calculateFileHash(BlobhubComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blobhub compose hash: %w", err)
	}

	for i, sys := range systems {
		if err := manage(l, sys, i, localBinaryPath, localServiceHash, localBinaryHash,
			localBlobRedisHash, localBlobpoolComposeHash, localBlobhubComposeHash); err != nil {
			return fmt.Errorf("failed to manage destination %s: %w", sys.Destination.Host, err)
		}
	}
	return nil
}

// manage handles the management of a single destination including service,
// binary, data and blob file management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, sys System, nodeId int,
	localBinaryPath string, localServiceHash, localBinaryHash,
	localBlobRedisHash, localBlobpoolComposeHash, localBlobhubComposeHash []byte,
) error {
	// Sequencer Related
	if err := manageService(l, sys, localServiceHash); err != nil {
		return fmt.Errorf("failed to manage service: %w", err)
	}

	if err := manageBinary(l, sys, localBinaryHash, localBinaryPath); err != nil {
		return fmt.Errorf("failed to manage binary: %w", err)
	}

	if err := manageData(l, sys, nodeId); err != nil {
		return fmt.Errorf("failed to manage data: %w", err)
	}

	// Blobpool & Blobhub Related
	if err := manageBlobStorageRedisConf(l, sys, localBlobRedisHash); err != nil {
		return fmt.Errorf("failed to manage blobpool storage redis file: %w", err)
	}

	if err := manageBlobpoolCompose(l, sys, localBlobpoolComposeHash); err != nil {
		return fmt.Errorf("failed to manage blobpool storage compose file: %w", err)
	}

	if err := manageBlobhubCompose(l, sys, localBlobhubComposeHash); err != nil {
		return fmt.Errorf("failed to manage blobhub storage compose file: %w", err)
	}

	return nil
}

// manageService handles the systemd service configuration for a destination.
// Checks if service exists, compares hashes and transfers updated service file if needed.
// Returns an error if service management fails.
func manageService(l *zap.SugaredLogger, sys System, localHash []byte) error {
	onlyShutdown := !sys.Options.Sequencer

	// Check if systemd service exists
	checkCmd := "test -f /etc/systemd/system/fuelsequencerd.service"
	l.Infow("checking for fuelsequencerd service...", "host", sys.Destination.Host, "cmd", checkCmd)
	if err := remotely(l, sys.SSH, checkCmd); err != nil {
		if onlyShutdown {
			l.Infow("only shutdown, no need to transfer service", "host", sys.Destination.Host)
			return nil
		}
		l.Infof("no fuelsequencerd service found on %s - will transfer service file...", sys.Destination.Host)
		if err := transferService(l, sys); err != nil {
			return fmt.Errorf("failed to transfer service file: %w", err)
		}
	} else {
		// Service exists, make sure it is disabled and stopped
		disableCmd := "systemctl disable fuelsequencerd"
		l.Infow("disabling fuelsequencerd...", "host", sys.Destination.Host, "cmd", disableCmd)
		if err := remotely(l, sys.SSH, withSudo(disableCmd, sys.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to disable service on %s: %w", sys.Destination.Host, err)
		}

		stopCmd := "systemctl stop fuelsequencerd"
		l.Infow("stopping fuelsequencerd...", "host", sys.Destination.Host, "cmd", stopCmd)
		if err := remotely(l, sys.SSH, withSudo(stopCmd, sys.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to stop service on %s: %w", sys.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to compare hashes", "host", sys.Destination.Host)
			return nil // only shutdown, no need to compare hashes
		}

		// Compare local and remote service file hashes
		l.Infow("comparing service file hashes...", "host", sys.Destination.Host)

		// Get remote service file hash
		remoteHash, err := calculateFileHash(BinaryConfig().SystemdPath, sys.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote service file hash: %w", err)
		}

		l.Debugw("hashes", "connection", sys.Destination.Host, "local", localHash, "remote", remoteHash)

		// Compare and replace if different
		if string(localHash) != string(remoteHash) {
			l.Infow("service file differs! replacing...", "host", sys.Destination.Host)
			if err := transferService(l, sys); err != nil {
				return fmt.Errorf("failed to transfer service file: %w", err)
			}
		}
	}
	return nil
}

// transferService transfers the systemd service file to a destination.
// Returns an error if transfer fails.
func transferService(l *zap.SugaredLogger, conn System) error {
	// First transfer to temporary location
	tmpServicePath := filepath.Join(conn.Destination.Dir, "tmp-fuelsequencerd.service")
	l.Debugw("transferring service to tmp file...", "from", ServicePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, tmpServicePath))

	if err := transfer(l, conn, ServicePath(), tmpServicePath); err != nil {
		return fmt.Errorf("failed to transfer service file to temp location: %w", err)
	}

	// Then move to final location with sudo
	moveCmd := fmt.Sprintf("mv %s %s", tmpServicePath, BinaryConfig().SystemdPath)
	if err := remotely(l, conn.SSH, withSudo(moveCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to move service file to system directory: %w", err)
	}

	l.Infow("...transferred service file to system directory", "from", ServicePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, BinaryConfig().SystemdPath))
	return nil
}

// manageBinary handles the binary deployment for a destination.
// Checks if binary exists, compares hashes and transfers updated binary if needed.
// Returns an error if binary management fails.
func manageBinary(l *zap.SugaredLogger, conn System, localBinaryHash []byte, localBinaryPath string) error {
	if !conn.Options.Sequencer {
		l.Infow("only shutdown, no need to compare hashes nor transfer binary", "host", conn.Destination.Host)
		return nil // only shutdown, no need to compare hashes nor transfer binary
	}

	// check if binary exists
	remoteBinaryPath := remoteBinaryPath(conn.Destination)
	checkCmd := "test -f " + remoteBinaryPath
	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infow("no binary found - will transfer binary...", "host", conn.Destination.Host)
		if err := transferBinary(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
			return fmt.Errorf("failed to transfer binary: %w", err)
		}
	} else {
		// if it does, compare local and remote binary hashes
		l.Debugw("comparing binary hashes...", "host", conn.Destination.Host)

		// Get remote binary hash
		remoteHash, err := calculateFileHash(remoteBinaryPath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote binary hash: %w", err)
		}

		l.Debugw("hashes", "connection", conn.Destination.Host, "local", localBinaryHash, "remote", remoteHash)

		// if they are different, replace the remote binary with the local one
		if string(localBinaryHash) != string(remoteHash) {
			l.Infof("binary differs on %s, replacing...", conn.Destination.Host)
			if err := transferBinary(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
				return fmt.Errorf("failed to transfer binary: %w", err)
			}
		}
	}
	return nil
}

// transferBinary transfers the fuelsequencerd binary to a destination.
// Returns an error if transfer fails.
func transferBinary(l *zap.SugaredLogger, conn System, localBinaryPath, remoteBinaryPath string) error {
	logging.Infow("transferring binary...", "from", localBinaryPath, "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteBinaryPath))
	if err := transfer(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
		return fmt.Errorf("failed to transfer binary to %s: %w", conn.Destination.Host, err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remoteBinaryPath)
	l.Infow("chowning binary to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown binary on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageData handles the chain data management for a destination.
// Cleans existing data and transfers new configuration.
// Returns an error if data management fails.
func manageData(l *zap.SugaredLogger, conn System, nodeId int) error {
	onlyDelete := !conn.Options.Sequencer

	// if data on remote exists, remove it
	homeDir := chainHomeDir(conn.Destination)
	checkCmd := "test -d " + homeDir
	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infof("no chain home directory found on %s", conn.Destination.Host)
		if onlyDelete {
			l.Infow("no chain home directory found on %s, but only delete, so no need to transfer config", conn.Destination.Host)
			return nil // only delete, found nothing, and no need to transfer config
		}
	} else {
		removeCmd := "rm -rf " + homeDir
		l.Infow("removing chain home directory...", "host", conn.Destination.Host, "cmd", removeCmd)
		if err := remotely(l, conn.SSH, withSudo(removeCmd, conn.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to remove home directory on %s: %w", conn.Destination.Host, err)
		}
		if onlyDelete {
			l.Infow("chain home directory found on %s, but only delete, so no need to transfer config", conn.Destination.Host)
			return nil
		}
	}

	// now transfer the updated config
	if err := transferConfig(l, conn, nodeId); err != nil {
		return fmt.Errorf("failed to transfer config: %w", err)
	}
	return nil
}

// transferConfig transfers the chain configuration files to a destination node.
// Takes a logger, connection details, and node ID.
// Transfers the configuration files and sets appropriate permissions.
// Returns an error if transfer fails.
func transferConfig(l *zap.SugaredLogger, conn System, nodeId int) error {
	instanceDir := filepath.Join(DataDir(), sequencer.ChainName, fmt.Sprintf("fuelsequencer%d", nodeId))
	remoteDataDir := chainHomeDir(conn.Destination)

	l.Infow("transferring config as chain home directory...", "from", instanceDir, "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteDataDir))
	if err := transfer(l, conn, instanceDir, remoteDataDir); err != nil {
		return fmt.Errorf("failed to transfer config to %s: %w", conn.Destination.Host, err)
	}

	chownCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDataDir)
	l.Infow("chowning chain home directory to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown chain home directory on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageBlobStorageRedisConf handles the redis.conf file deployment for a destination.
// Checks if file exists, compares hashes and transfers updated file if needed.
// Returns an error if blob redis management fails.
func manageBlobStorageRedisConf(l *zap.SugaredLogger, conn System, localHash []byte) error {
	remotePath := remoteBlobStorageRedisConfPath(conn.Destination)
	checkCmd := "test -f " + remotePath
	onlyDelete := !conn.Options.BlobPool && !conn.Options.BlobHub // only delete if neither blobpool nor blobhub are enabled

	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infof("no blob-storage redis conf file found on %s - will transfer...", conn.Destination.Host)
		if onlyDelete {
			l.Infow("only delete, no need to transfer blob-storage redis conf file", "host", conn.Destination.Host)
			return nil
		}
		if err := transferBlobStorageRedisConf(l, conn); err != nil {
			return fmt.Errorf("failed to transfer blob-storage redis conf file: %w", err)
		}
	} else {
		if onlyDelete {
			removeCmd := "rm -f " + remotePath
			l.Infow("blob-storage redis conf file found on %s, but only delete, so removing blob-storage redis conf file...", "host", conn.Destination.Host, "cmd", removeCmd)
			if err := remotely(l, conn.SSH, withSudo(removeCmd, conn.Destination.Pass)); err != nil {
				return fmt.Errorf("failed to remove blob-storage redis conf file on %s: %w", conn.Destination.Host, err)
			}
			return nil
		}

		// File exists, compare local and remote file hashes
		l.Debugw("comparing blob-storage redis conf file hashes...", "host", conn.Destination.Host)

		// Get remote blob redis file hash
		remoteHash, err := calculateFileHash(remotePath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote blob-storage redis conf file hash: %w", err)
		}

		l.Debugw("blob-storage redis conf hashes", "connection", conn.Destination.Host, "local", localHash, "remote", remoteHash)

		// Compare and replace if different
		if string(localHash) != string(remoteHash) {
			l.Infow("blob-storage redis conf file differs! replacing...", "host", conn.Destination.Host)
			if err := transferBlobStorageRedisConf(l, conn); err != nil {
				return fmt.Errorf("failed to transfer blob redis file: %w", err)
			}
		}
	}
	return nil
}

// transferBlobStorageRedisConf transfers the redis.conf file to a destination.
// Returns an error if transfer fails.
func transferBlobStorageRedisConf(l *zap.SugaredLogger, conn System) error {
	remoteDir := remoteBlobDir(conn.Destination)
	remotePath := remoteBlobStorageRedisConfPath(conn.Destination)

	// Ensure remote blob directory exists (should already exist from compose file transfer)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob-storage directory: %w", err)
	}

	l.Infow("transferring blob-storage redis conf file...", "from", BlobRedisPath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, BlobRedisPath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blob-storage redis conf file: %w", err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remotePath)
	l.Infow("chowning blob-storage redis conf file to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blob-storage redis conf file on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageBlobpoolCompose handles the docker-compose.BlobPool.yml file deployment for a destination.
// Checks if file exists, compares hashes and transfers updated file if needed.
// Returns an error if blob compose management fails.
func manageBlobpoolCompose(l *zap.SugaredLogger, conn System, localHash []byte) error {
	remotePath := remoteBlobpoolComposePath(conn.Destination)
	checkCmd := "test -f " + remotePath
	onlyShutdown := !conn.Options.BlobPool

	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		if onlyShutdown {
			l.Infow("only shutdown, no need to transfer blobpool storage compose file", "host", conn.Destination.Host)
			return nil
		}
		l.Infof("no blobpool storage compose file found on %s - will transfer...", conn.Destination.Host)
		if err := transferBlobpoolCompose(l, conn); err != nil {
			return fmt.Errorf("failed to transfer blobpool storage compose file: %w", err)
		}
	} else {
		// Shutdown existing blobpool containers and reset their volumes (will be redeployed later)
		composeDir := remoteBlobpoolComposePath(conn.Destination)
		downCmd := fmt.Sprintf("docker compose -f=%s down -v", composeDir)
		l.Infow("shutting down and resetting blobpool storage...", "host", conn.Destination.Host, "cmd", downCmd)
		if err := remotely(l, conn.SSH, withSudo(downCmd, conn.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to shutdown and reset blobpool storage on %s: %w", conn.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to compare hashes", "host", conn.Destination.Host)
			return nil
		}

		// File exists, compare local and remote file hashes
		l.Debugw("comparing blobpool storage compose file hashes...", "host", conn.Destination.Host)

		// Get remote blob compose file hash
		remoteHash, err := calculateFileHash(remotePath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote blobpool storage compose file hash: %w", err)
		}

		l.Debugw("blobpool storage compose hashes", "connection", conn.Destination.Host, "local", localHash, "remote", remoteHash)

		// Compare and replace if different
		if string(localHash) != string(remoteHash) {
			l.Infow("blobpool storage compose file differs! replacing...", "host", conn.Destination.Host)
			if err := transferBlobpoolCompose(l, conn); err != nil {
				return fmt.Errorf("failed to transfer blobpool storage compose file: %w", err)
			}
		}
	}
	return nil
}

// transferBlobpoolCompose transfers the docker-compose.BlobPool.yml file to a destination.
// Returns an error if transfer fails.
func transferBlobpoolCompose(l *zap.SugaredLogger, conn System) error {
	remoteDir := remoteBlobDir(conn.Destination)
	remotePath := remoteBlobpoolComposePath(conn.Destination)

	// Ensure remote blob directory exists
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob directory: %w", err)
	}

	l.Infow("transferring blobpool storage compose file...", "from", BlobpoolComposePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, BlobpoolComposePath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blobpool storage compose file: %w", err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remotePath)
	l.Infow("chowning blobpool storage compose file to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blobpool storage compose file on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageBlobhubCompose handles the docker-compose.BlobHub.yml file deployment for a destination.
// Checks if file exists, compares hashes and transfers updated file if needed.
// Returns an error if blob compose management fails.
func manageBlobhubCompose(l *zap.SugaredLogger, conn System, localHash []byte) error {
	remotePath := remoteBlobhubComposePath(conn.Destination)
	checkCmd := "test -f " + remotePath
	onlyShutdown := !conn.Options.BlobHub

	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		if onlyShutdown {
			l.Infow("only shutdown, no need to transfer blobhub storage compose file", "host", conn.Destination.Host)
			return nil
		}
		l.Infof("no blobhub storage compose file found on %s - will transfer...", conn.Destination.Host)
		if err := transferBlobhubCompose(l, conn); err != nil {
			return fmt.Errorf("failed to transfer blobhub storage compose file: %w", err)
		}
	} else {
		// Shutdown existing blobhub containers and reset their volumes (will be redeployed later)
		composeDir := remoteBlobhubComposePath(conn.Destination)
		downCmd := fmt.Sprintf("docker compose -f=%s down -v", composeDir)
		l.Infow("shutting down and resetting blobhub storage...", "host", conn.Destination.Host, "cmd", downCmd)
		if err := remotely(l, conn.SSH, withSudo(downCmd, conn.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to shutdown and reset blobhub storage on %s: %w", conn.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to compare hashes", "host", conn.Destination.Host)
			return nil
		}

		// File exists, compare local and remote file hashes
		l.Debugw("comparing blobhub storage compose file hashes...", "host", conn.Destination.Host)

		// Get remote blob compose file hash
		remoteHash, err := calculateFileHash(remotePath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote blobhub storage compose file hash: %w", err)
		}

		l.Debugw("blobhub storage compose hashes", "connection", conn.Destination.Host, "local", localHash, "remote", remoteHash)

		// Compare and replace if different
		if string(localHash) != string(remoteHash) {
			l.Infow("blobhub storage compose file differs! replacing...", "host", conn.Destination.Host)
			if err := transferBlobhubCompose(l, conn); err != nil {
				return fmt.Errorf("failed to transfer blobhub storage compose file: %w", err)
			}
		}
	}
	return nil
}

// transferBlobhubCompose transfers the docker-compose.BlobHub.yml file to a destination.
// Returns an error if transfer fails.
func transferBlobhubCompose(l *zap.SugaredLogger, conn System) error {
	remoteDir := remoteBlobDir(conn.Destination)
	remotePath := remoteBlobhubComposePath(conn.Destination)

	// Ensure remote blob directory exists
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob directory: %w", err)
	}

	l.Infow("transferring blobhub storage compose file...", "from", BlobhubComposePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, BlobhubComposePath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blobhub storage compose file: %w", err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remotePath)
	l.Infow("chowning blobhub storage compose file to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blobhub storage compose file on %s: %w", conn.Destination.Host, err)
	}
	return nil
}
