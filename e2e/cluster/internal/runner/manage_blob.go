package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageBlobStorageRedisConf handles the redis.conf file deployment for a destination.
// Checks if file exists, compares hashes and transfers updated file if needed.
// Returns an error if blob redis management fails.
func manageBlobStorageRedisConf(l *zap.SugaredLogger, conn setup.System, localHash []byte) error {
	remotePath := setup.RemoteBlobStorageRedisConfPath(conn.Destination)
	checkCmd := "test -f " + remotePath
	onlyDelete := !conn.Options.Blobpool && !conn.Options.Blobhub // only delete if neither blobpool nor blobhub are enabled

	if err := execute.Remotely(l, conn.SSH, checkCmd); err != nil {
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
			if err := execute.Remotely(l, conn.SSH, execute.WithSudo(removeCmd, conn.Destination.Pass)); err != nil {
				return fmt.Errorf("failed to remove blob-storage redis conf file on %s: %w", conn.Destination.Host, err)
			}
			return nil
		}

		// File exists, compare local and remote file hashes
		l.Debugw("comparing blob-storage redis conf file hashes...", "host", conn.Destination.Host)

		// Get remote blob redis file hash
		remoteHash, err := execute.CalculateFileHash(remotePath, conn.SSH)
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
func transferBlobStorageRedisConf(l *zap.SugaredLogger, conn setup.System) error {
	remoteDir := setup.RemoteBlobDir(conn.Destination)
	remotePath := setup.RemoteBlobStorageRedisConfPath(conn.Destination)

	// Ensure remote blob directory exists (should already exist from compose file transfer)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := execute.Remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob-storage directory: %w", err)
	}

	l.Infow("transferring blob-storage redis conf file...", "from", setup.BlobRedisPath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, setup.BlobRedisPath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blob-storage redis conf file: %w", err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remotePath)
	l.Infow("chowning blob-storage redis conf file to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blob-storage redis conf file on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageBlobpoolCompose handles the docker-compose.blobpool.yml file deployment for a destination.
// Checks if file exists, compares hashes and transfers updated file if needed.
// Returns an error if blob compose management fails.
func manageBlobpoolCompose(l *zap.SugaredLogger, conn setup.System, localHash []byte) error {
	remotePath := setup.RemoteBlobpoolComposePath(conn.Destination)
	checkCmd := "test -f " + remotePath
	onlyShutdown := !conn.Options.Blobpool

	if err := execute.Remotely(l, conn.SSH, checkCmd); err != nil {
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
		composeDir := setup.RemoteBlobpoolComposePath(conn.Destination)
		downCmd := fmt.Sprintf("docker compose -f=%s down -v", composeDir)
		l.Infow("shutting down and resetting blobpool storage...", "host", conn.Destination.Host, "cmd", downCmd)
		if err := execute.Remotely(l, conn.SSH, execute.WithSudo(downCmd, conn.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to shutdown and reset blobpool storage on %s: %w", conn.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to compare hashes", "host", conn.Destination.Host)
			return nil
		}

		// File exists, compare local and remote file hashes
		l.Debugw("comparing blobpool storage compose file hashes...", "host", conn.Destination.Host)

		// Get remote blob compose file hash
		remoteHash, err := execute.CalculateFileHash(remotePath, conn.SSH)
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

// transferBlobpoolCompose transfers the docker-compose.blobpool.yml file to a destination.
// Returns an error if transfer fails.
func transferBlobpoolCompose(l *zap.SugaredLogger, conn setup.System) error {
	remoteDir := setup.RemoteBlobDir(conn.Destination)
	remotePath := setup.RemoteBlobpoolComposePath(conn.Destination)

	// Ensure remote blob directory exists
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := execute.Remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob directory: %w", err)
	}

	l.Infow("transferring blobpool storage compose file...", "from", setup.BlobpoolComposePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, setup.BlobpoolComposePath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blobpool storage compose file: %w", err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remotePath)
	l.Infow("chowning blobpool storage compose file to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blobpool storage compose file on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageBlobhub handles the Blobhub deployment for a destination.
// Checks if the project directory exists and transfers updated directory if needed.
// Returns an error if any part of the management fails.
func manageBlobhub(l *zap.SugaredLogger, conn setup.System) error {
	remotePath := setup.RemoteBlobhubDir(conn.Destination)
	checkCmd := "test -d " + remotePath
	onlyShutdown := !conn.Options.Blobhub

	if err := execute.Remotely(l, conn.SSH, checkCmd); err != nil {
		if onlyShutdown {
			l.Infow("only shutdown, no need to transfer blobhub storage", "host", conn.Destination.Host)
			return nil
		}
		l.Infof("no blobhub storage found on %s - will transfer...", conn.Destination.Host)
		if err := transferBlobhub(l, conn); err != nil {
			return fmt.Errorf("failed to transfer blobhub storage: %w", err)
		}
	} else {
		// Shutdown existing blobhub containers and reset their volumes
		composeDir := setup.RemoteBlobhubComposeDir(conn.Destination)
		downCmd := fmt.Sprintf("docker compose -f=%s down -v", composeDir)
		l.Infow("shutting down and resetting blobhub storage...", "host", conn.Destination.Host, "cmd", downCmd)
		if err := execute.Remotely(l, conn.SSH, execute.WithSudo(downCmd, conn.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to shutdown and reset blobhub storage on %s: %w", conn.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to transfer", "host", conn.Destination.Host)
			return nil
		}

		// Transfer updated directory using rsync
		l.Infof("transferring updated blobhub storage to %s...", conn.Destination.Host)
		if err := transferBlobhub(l, conn); err != nil {
			return fmt.Errorf("failed to transfer blobhub storage: %w", err)
		}
	}
	return nil
}

func transferBlobhub(l *zap.SugaredLogger, conn setup.System) error {
	remoteDir := setup.RemoteBlobDir(conn.Destination)
	remotePath := setup.RemoteBlobhubDir(conn.Destination)

	// Ensure remote blob directory exists with proper ownership and permissions
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := execute.Remotely(l, conn.SSH, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote blob directory: %w", err)
	}

	// Set ownership and permissions BEFORE transfer
	setupCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDir)
	l.Infow("setting up ownership before transfer...", "host", conn.Destination.Host, "cmd", setupCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(setupCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to setup ownership on %s: %w", conn.Destination.Host, err)
	}

	chmodCmd := fmt.Sprintf("chmod -R g+rwx %s", remoteDir)
	l.Infow("setting up permissions before transfer...", "host", conn.Destination.Host, "cmd", chmodCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chmodCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to setup permissions on %s: %w", conn.Destination.Host, err)
	}

	// Now transfer will work because tharen has group write access
	l.Infow("transferring blobhub storage...", "from", setup.BlobhubDirPath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remotePath))
	if err := transfer(l, conn, setup.BlobhubDirPath(), remotePath); err != nil {
		return fmt.Errorf("failed to transfer blobhub storage: %w", err)
	}

	// Optional: Reset ownership after transfer in case local files had different ownership
	chownCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remotePath)
	l.Infow("ensuring correct ownership after transfer...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown blobhub storage on %s: %w", conn.Destination.Host, err)
	}

	return nil
}
