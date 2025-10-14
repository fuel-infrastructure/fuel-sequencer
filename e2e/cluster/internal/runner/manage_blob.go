package runner

import (
	"fmt"

	"go.uber.org/zap"
)

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

// manageBlobhubCompose handles the docker-compose.blobhub.yml file deployment for a destination.
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
