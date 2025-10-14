// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/setup"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// transfer copies a file or directory from the local machine to a remote destination using rsync.
// Takes a logger, connection, local path, and remote path.
// Returns an error if the transfer fails.
func transfer(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string) error {
	if err := ensureDir(l, conn.SSH, filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("failed to ensure remote directory: %w", err)
	}

	if err := transferPath(l, conn, localPath, remotePath); err != nil {
		return fmt.Errorf("failed to transfer file: %w", err)
	}

	return nil
}

// ensureDir ensures that a directory exists on the remote machine.
// Creates the directory and any parent directories if they don't exist.
// Returns an error if directory creation fails.
func ensureDir(l *zap.SugaredLogger, client *ssh.Client, remotePath string) error {
	// First, ensure the remote directory exists
	remoteDir := filepath.Dir(remotePath)
	mkdirCmd := fmt.Sprintf("mkdir -vp %s", remoteDir)
	if err := remotely(l, client, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	return nil
}

// transferPath handles the actual rsync transfer of files or directories.
// Implements rsync with progress reporting and retry mechanism.
// Returns an error if the transfer fails after all retry attempts.
func transferPath(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		l.Debugf("transferring to %s (attempt %d/%d)", remotePath, attempt, maxRetries)

		startTime := time.Now()
		err := attemptTransfer(l, conn, localPath, remotePath)
		duration := time.Since(startTime)

		if err == nil {
			l.Infow("successfully transferred", "remotePath", remotePath, "duration", duration)
			return nil
		}

		l.Warnw("transfer attempt failed", "localPath", localPath, "remotePath", remotePath, "attempt", attempt, "duration", duration, "error", err)

		if attempt < maxRetries {
			l.Infof("retrying transfer in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("transfer failed after %d attempts", maxRetries)
}

// attemptTransfer performs a single transfer attempt using rsync.
func attemptTransfer(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string) error {
	// Get file info for size and type
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path: %w", err)
	}

	// Build rsync command with equivalent flags to -rvzhP
	// -r: recursive (for directories)
	// -v: verbose
	// -z: compress
	// -h: human-readable
	// -P: progress and partial (resume)
	flags := "-rvzhP"

	// Ensure trailing slash for directories to sync contents
	sourcePath := localPath
	if stat.IsDir() && !strings.HasSuffix(sourcePath, "/") {
		sourcePath += "/"
	}

	// Try rsync with sshpass first, fallback to SCP if not available
	remoteDest := fmt.Sprintf("%s@%s:%s", conn.Destination.User, conn.Destination.Host, remotePath)

	// Construct rsync command with sshpass
	rsyncCmd := fmt.Sprintf("sshpass -p '%s' rsync %s %s %s", conn.Destination.Pass, flags, sourcePath, remoteDest)

	l.Debugw("executing rsync command", "command", rsyncCmd)

	// Execute rsync command with sshpass
	args := []string{"-p", conn.Destination.Pass, "rsync"}
	args = append(args, strings.Fields(flags)...)
	args = append(args, sourcePath, remoteDest)
	if err := locallyWithCustomName(l, "sshpass", "rsync", args...); err != nil {
		return fmt.Errorf("rsync with sshpass failed: %w", err)
	}

	return nil
}
