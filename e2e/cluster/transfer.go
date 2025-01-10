// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// transfer copies a file or directory from the local machine to a remote destination using SCP.
// Takes a logger, SSH client, local path, and remote path.
// Returns an error if the transfer fails.
func transfer(l *zap.SugaredLogger, client *ssh.Client, localPath, remotePath string) error {
	if err := ensureDir(l, client, filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("failed to ensure remote directory: %w", err)
	}

	if err := transferPath(l, client, localPath, remotePath); err != nil {
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

// transferPath handles the actual SCP transfer of files or directories.
// Implements the SCP protocol for secure file transfer.
// Returns an error if the transfer fails.
func transferPath(l *zap.SugaredLogger, client *ssh.Client, localPath, remotePath string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Get file info for size and type
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat local path: %w", err)
	}

	// Set up SCP flags
	flags := "-qt"
	if stat.IsDir() {
		flags += "r"
	}

	// Start scp command on remote
	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Create error channel for receiving copy errors
	errCh := make(chan error, 1)

	// Start scp receiver on the remote side
	go func() {
		defer w.Close()
		if stat.IsDir() {
			err = sendDirectory(w, localPath, filepath.Base(remotePath))
		} else {
			err = sendFile(w, localPath, filepath.Base(remotePath), stat)
		}
		errCh <- err
	}()

	// Start scp command on remote
	if err := session.Run("scp " + flags + " " + filepath.Dir(remotePath)); err != nil {
		l.Errorw("transfer failed", "localPath", localPath, "remotePath", remotePath, "error", err)
		return fmt.Errorf("scp command failed: %w", err)
	}

	// Check for any errors during transfer
	if err := <-errCh; err != nil {
		return err
	}

	l.Debugf("successfully transferred to %s", remotePath)
	return nil
}

// sendDirectory recursively transfers a directory and its contents using SCP.
// Implements the SCP protocol's directory transfer functionality.
// Returns an error if the directory transfer fails.
func sendDirectory(w io.Writer, localPath, remoteName string) error {
	stat, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}

	// Send directory header
	fmt.Fprintf(w, "D%04o 0 %s\n", stat.Mode().Perm(), remoteName)

	entries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Process all directory entries
	for _, entry := range entries {
		fullPath := filepath.Join(localPath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("failed to get entry info: %w", err)
		}

		if info.IsDir() {
			if err := sendDirectory(w, fullPath, entry.Name()); err != nil {
				return err
			}
		} else {
			if err := sendFile(w, fullPath, entry.Name(), info); err != nil {
				return err
			}
		}
	}

	// Send directory footer
	fmt.Fprint(w, "E\n")
	return nil
}

// sendFile transfers a single file using SCP.
// Implements the SCP protocol's file transfer functionality.
// Returns an error if the file transfer fails.
func sendFile(w io.Writer, localPath, remoteName string, stat os.FileInfo) error {
	// Send file header
	fmt.Fprintf(w, "C%04o %d %s\n", stat.Mode().Perm(), stat.Size(), remoteName)

	// Open and send file contents
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Send file terminator
	fmt.Fprint(w, "\x00")
	return nil
}
