// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// transfer copies a file or directory to a destination (local or remote).
// For local destinations, uses local file operations. For remote, uses rsync.
// Takes a logger, connection, local path, and remote path.
// Returns an error if the transfer fails.
func transfer(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string, excludes ...string) error {
	if err := ensureDir(l, conn, filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("failed to ensure directory: %w", err)
	}

	if err := transferPath(l, conn, localPath, remotePath, excludes); err != nil {
		return fmt.Errorf("failed to transfer file: %w", err)
	}

	return nil
}

// ensureDir ensures that a directory exists on the system (local or remote).
// Creates the directory and any parent directories if they don't exist.
// Returns an error if directory creation fails.
func ensureDir(l *zap.SugaredLogger, conn setup.System, remotePath string) error {
	remoteDir := filepath.Dir(remotePath)
	if conn.IsLocal {
		// Local: use os.MkdirAll
		if err := os.MkdirAll(remoteDir, 0755); err != nil {
			return fmt.Errorf("failed to create local directory: %w", err)
		}
		return nil
	}
	// Remote: use mkdir command via SSH
	mkdirCmd := fmt.Sprintf("mkdir -vp %s", remoteDir)
	if err := execute.OnSystem(l, conn, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}
	return nil
}

// transferPath handles the actual rsync transfer of files or directories.
// Implements rsync with progress reporting and retry mechanism.
// Returns an error if the transfer fails after all retry attempts.
func transferPath(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string, excludes []string) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		l.Debugf("transferring to %s (attempt %d/%d)", remotePath, attempt, maxRetries)

		startTime := time.Now()
		err := attemptTransfer(l, conn, localPath, remotePath, excludes)
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

// attemptTransfer performs a single transfer attempt.
// For local destinations, uses local file operations. For remote, uses rsync with sshpass.
func attemptTransfer(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string, excludes []string) error {
	// Get file info for size and type
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path: %w", err)
	}

	// Local transfer: use local file operations
	if conn.IsLocal {
		return transferLocal(l, localPath, remotePath, stat.IsDir(), excludes)
	}

	// Remote transfer: use rsync with sshpass
	return transferRemote(l, conn, localPath, remotePath, stat.IsDir(), excludes)
}

// transferLocal performs a local file or directory copy.
func transferLocal(l *zap.SugaredLogger, localPath, remotePath string, isDir bool, excludes []string) error {
	if isDir {
		// For directories, use rsync locally (without SSH) for efficient syncing
		flags := "-rvzhP"
		for _, exclude := range excludes {
			flags += fmt.Sprintf(" --exclude=%s", exclude)
		}
		sourcePath := localPath
		if !strings.HasSuffix(sourcePath, "/") {
			sourcePath += "/"
		}
		// Build rsync command and execute via sh -c
		rsyncCmd := fmt.Sprintf("rsync %s %s %s", flags, sourcePath, remotePath)
		l.Debugw("executing local rsync command", "command", rsyncCmd)
		// Use a temporary System with nil SSH to route to local execution
		tempSys := setup.System{IsLocal: true}
		if err := execute.OnSystem(l, tempSys, rsyncCmd); err != nil {
			// Fallback to cp -r if rsync is not available
			l.Debugw("rsync failed, falling back to cp", "error", err)
			cpCmd := fmt.Sprintf("cp -r %s %s", sourcePath, remotePath)
			return execute.OnSystem(l, tempSys, cpCmd)
		}
		return nil
	}

	// For files, use os operations to copy a single file from source to destination
	sourceFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Preserve source file permissions
	sourceInfo, err := os.Stat(localPath)
	if err == nil {
		os.Chmod(remotePath, sourceInfo.Mode())
	}

	return nil
}

// transferRemote performs a remote file or directory transfer using rsync with sshpass.
func transferRemote(l *zap.SugaredLogger, conn setup.System, localPath, remotePath string, isDir bool, excludes []string) error {
	// Build rsync command with equivalent flags to -rvzhP
	// -r: recursive (for directories)
	// -v: verbose
	// -z: compress
	// -h: human-readable
	// -P: progress and partial (resume)
	flags := "-rvzhP"

	// Add exclude patterns if provided
	for _, exclude := range excludes {
		flags += fmt.Sprintf(" --exclude=%s", exclude)
	}

	// Ensure trailing slash for directories to sync contents
	sourcePath := localPath
	if isDir && !strings.HasSuffix(sourcePath, "/") {
		sourcePath += "/"
	}

	// Remote destination with user@host:path format
	remoteDest := fmt.Sprintf("%s@%s:%s", conn.Destination.User, conn.Destination.Host, remotePath)

	// Construct rsync command with sshpass
	rsyncCmd := fmt.Sprintf("sshpass -p '%s' rsync %s %s %s", conn.Destination.Pass, flags, sourcePath, remoteDest)

	l.Debugw("executing rsync command", "command", rsyncCmd)

	// Execute rsync command with sshpass
	args := []string{"-p", conn.Destination.Pass, "rsync"}
	args = append(args, strings.Fields(flags)...)
	args = append(args, sourcePath, remoteDest)
	if err := execute.LocallyWithCustomName(l, "sshpass", "rsync", args...); err != nil {
		return fmt.Errorf("rsync with sshpass failed: %w", err)
	}

	return nil
}
