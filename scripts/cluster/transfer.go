package cluster

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func transferFiles(connections []connection, binaryPath string) error {
	l := logging.Named("Transfer")
	binaryName := filepath.Base(binaryPath)
	for _, conn := range connections {
		remoteBinaryPath := filepath.Join(conn.dir, binaryName)
		logging.Infow("transferring binary...", "from", binaryPath, "to", fmt.Sprintf("%s@%s:%s", conn.user, conn.host, remoteBinaryPath))
		if err := transfer(l, conn.SSH, binaryPath, remoteBinaryPath); err != nil {
			return fmt.Errorf("failed to transfer binary to %s: %w", conn.destination.host, err)
		}

		remoteDataDir := filepath.Join(conn.dir, ".fuelsequencer")
		logging.Infow("transferring data...", "from", dataDir, "to", fmt.Sprintf("%s@%s:%s", conn.user, conn.host, remoteDataDir))
		if err := transfer(l, conn.SSH, dataDir, remoteDataDir); err != nil {
			return fmt.Errorf("failed to transfer data to %s: %w", conn.destination.host, err)
		}
	}

	return nil
}

// transfer a file over ssh using scp
func transfer(l *zap.SugaredLogger, client *ssh.Client, localPath, remotePath string) error {
	if err := ensureDir(l, client, filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("failed to ensure remote directory: %w", err)
	}

	if err := transferPath(l, client, localPath, remotePath); err != nil {
		return fmt.Errorf("failed to transfer file: %w", err)
	}

	// if err := verifyTransfer(conn.SSH, remotePath, 0); err != nil {
	// 	return fmt.Errorf("failed to verify binary transfer to %s: %w", conn.destination.host, err)
	// }

	return nil
}

func ensureDir(l *zap.SugaredLogger, client *ssh.Client, remotePath string) error {
	// Create new SFTP client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// First, ensure the remote directory exists
	remoteDir := filepath.Dir(remotePath)
	mkdirCmd := fmt.Sprintf("mkdir -vp %s", remoteDir)
	if err := remotely(l, session, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	return nil
}

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

// // Additional helper function to verify file transfer
// func verifyTransfer(client *ssh.Client, remotePath string, expectedSize int64) error {
// 	log.Debugf("Verifying file transfer: %s", remotePath)

// 	cmd := fmt.Sprintf("stat -f '%%z' %s", remotePath)
// 	session, err := client.NewSession()
// 	if err != nil {
// 		return fmt.Errorf("failed to create session: %w", err)
// 	}
// 	defer session.Close()

// 	var output bytes.Buffer
// 	session.Stdout = &output

// 	if err := session.Run(cmd); err != nil {
// 		return fmt.Errorf("failed to verify file: %w", err)
// 	}

// 	size, err := strconv.ParseInt(strings.TrimSpace(output.String()), 10, 64)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse file size: %w", err)
// 	}

// 	if size != expectedSize {
// 		return fmt.Errorf("size mismatch: expected %d, got %d", expectedSize, size)
// 	}

// 	log.Debug("File transfer verified successfully")
// 	return nil
// }
