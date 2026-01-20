package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageBlobhub handles the Blobhub deployment for a destination.
// Checks if the project directory exists and transfers updated directory if needed.
// Blobhub is deployed once per system (using instanceId 0 directory) and shared across all instances.
// Returns an error if any part of the management fails.
func manageBlobhub(l *zap.SugaredLogger, conn setup.System) error {
	// Always use instance 0 directory for blobhub (shared across all instances on this system)
	var remotePath string
	if conn.IsLocal {
		remotePath = setup.LocalBlobhubDir()
	} else {
		remotePath = setup.RemoteBlobhubDir(conn.Destination, 0)
	}
	checkCmd := "test -d " + remotePath
	onlyShutdown := !conn.Options.Blobhub

	if err := execute.OnSystem(l, conn, checkCmd); err != nil {
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
		var composeDir string
		if conn.IsLocal {
			composeDir = setup.LocalBlobhubComposeDir()
		} else {
			composeDir = setup.RemoteBlobhubComposeDir(conn.Destination, 0)
		}
		downCmd := fmt.Sprintf("docker compose -f=%s down -v", composeDir)
		l.Infow("shutting down and resetting blobhub storage...", "host", conn.Destination.Host, "cmd", downCmd)
		if err := execute.OnSystemWithSudo(l, conn, downCmd); err != nil {
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
	if conn.IsLocal {
		l.Infow("skipping blobhub transfer for local destination", "host", conn.Destination.Host)
		return nil
	}

	// Always use instance 0 directory for blobhub (shared across all instances)
	remoteDir := setup.RemoteBlobDir(conn.Destination, 0)
	remotePath := setup.RemoteBlobhubDir(conn.Destination, 0)

	// Ensure blob directory exists with proper ownership and permissions
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := execute.OnSystem(l, conn, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create blob directory: %w", err)
	}

	// Set ownership and permissions BEFORE transfer
	setupCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDir)
	l.Infow("setting up ownership before transfer...", "host", conn.Destination.Host, "cmd", setupCmd)
	if err := execute.OnSystemWithSudo(l, conn, setupCmd); err != nil {
		return fmt.Errorf("failed to setup ownership on %s: %w", conn.Destination.Host, err)
	}

	chmodCmd := fmt.Sprintf("chmod -R g+rwx %s", remoteDir)
	l.Infow("setting up permissions before transfer...", "host", conn.Destination.Host, "cmd", chmodCmd)
	if err := execute.OnSystemWithSudo(l, conn, chmodCmd); err != nil {
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
	if err := execute.OnSystemWithSudo(l, conn, chownCmd); err != nil {
		return fmt.Errorf("failed to chown blobhub storage on %s: %w", conn.Destination.Host, err)
	}

	return nil
}
