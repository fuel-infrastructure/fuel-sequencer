package cluster

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func transferBinary(connections []connection, binaryPath string) error {
	l := logging.Named("Transfer")
	binaryName := filepath.Base(binaryPath)
	for _, conn := range connections {
		remotePath := filepath.Join(conn.dir, binaryName)

		logging.Debugw("transferring file...", "from", binaryPath, "to", fmt.Sprintf("%s@%s:%s", conn.user, conn.host, remotePath))
		if err := transfer(l, conn.SSH, binaryPath, remotePath); err != nil {
			return fmt.Errorf("failed to transfer binary to %s: %w", conn.destination.host, err)
		}
	}

	return nil
}

// transfer a file over ssh using scp
func transfer(l *zap.SugaredLogger, client *ssh.Client, localPath, remotePath string) error {
	if err := ensureDir(l, client, filepath.Dir(remotePath)); err != nil {
		return fmt.Errorf("failed to ensure remote directory: %w", err)
	}

	if err := transferFile(l, client, localPath, remotePath); err != nil {
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

func transferFile(l *zap.SugaredLogger, client *ssh.Client, localPath, remotePath string) error {
	// Create new SFTP client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Create SFTP pipe
	w, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Read local file
	localFile, err := os.Open(localPath)
	if err != nil {
		l.Errorw("failed to open local file", "path", localPath, "error", err)
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Get file info for size
	stat, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file: %w", err)
	}

	// Create error channel for receiving copy errors
	errCh := make(chan error, 1)

	// Start scp receiver on the remote side
	go func() {
		defer w.Close()

		// Send file metadata
		fmt.Fprintf(w, "C0775 %d %s\n", stat.Size(), filepath.Base(remotePath))

		// Copy file content
		if _, err := io.Copy(w, localFile); err != nil {
			errCh <- fmt.Errorf("failed to copy file content: %w", err)
			return
		}

		// Send transfer completion signal
		fmt.Fprint(w, "\x00")
		errCh <- nil
	}()

	// Start scp command on remote
	if err := session.Run("scp -qt " + filepath.Dir(remotePath)); err != nil {
		l.Errorw("file transfer failed", "localPath", localPath, "remotePath", remotePath, "error", err)
		return fmt.Errorf("scp command failed: %w", err)
	}

	// Check for any errors during transfer
	if err := <-errCh; err != nil {
		return err
	}

	l.Debugf("successfully transferred file to %s", remotePath)
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
