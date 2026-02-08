package runner

import (
	"fmt"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageBlobhub handles the Blobhub deployment for a destination.
// Rsyncs the full blobhub project and pre-built image tar to the remote, loads the image(s), and leaves deploy to compose up.
// Blobhub is deployed once per system (using instanceId 0 directory) and shared across all instances.
// blobhubImagePath is the local path to the blobhub image tar for this destination's platform; empty for local-only or shutdown-only.
// Returns an error if any part of the management fails.
func manageBlobhub(l *zap.SugaredLogger, conn setup.System, blobhubImagePath string) error {
	// Always use instance 0 directory for blobhub (shared across all instances on this system)
	var remoteComposeDir string
	if conn.IsLocal {
		remoteComposeDir = setup.LocalBlobhubComposeDir()
	} else {
		remoteComposeDir = setup.RemoteBlobhubComposeDir(conn.Destination, 0)
	}
	remoteBlobhubDir := filepath.Dir(remoteComposeDir)
	onlyShutdown := !conn.Options.Blobhub

	// Shutdown existing blobhub containers if present (check by compose dir existence or always try down)
	downCmd := fmt.Sprintf("docker compose -f=%s down -v 2>/dev/null || true", remoteComposeDir)
	l.Infow("ensuring blobhub is down...", "host", conn.Destination.Host, "cmd", downCmd)
	if err := execute.OnSystemWithSudo(l, conn, downCmd); err != nil {
		return fmt.Errorf("failed to shutdown blobhub on %s: %w", conn.Destination.Host, err)
	}

	if onlyShutdown {
		l.Infow("only shutdown, no need to transfer blobhub", "host", conn.Destination.Host)
		return nil
	}

	if conn.IsLocal {
		// Local: image already built and loaded; compose file is at LocalBlobhubDir
		l.Infow("local blobhub, image and compose already in place", "host", conn.Destination.Host)
		return nil
	}

	// Remote: ensure directory exists, transfer compose file and image tar, load image(s)
	remoteDir := setup.RemoteBlobDir(conn.Destination, 0)
	mkdirCmd := fmt.Sprintf("mkdir -p %s", remoteDir)
	if err := execute.OnSystem(l, conn, mkdirCmd); err != nil {
		return fmt.Errorf("failed to create blob directory: %w", err)
	}
	setupCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDir)
	l.Infow("setting up ownership...", "host", conn.Destination.Host, "cmd", setupCmd)
	if err := execute.OnSystemWithSudo(l, conn, setupCmd); err != nil {
		return fmt.Errorf("failed to setup ownership on %s: %w", conn.Destination.Host, err)
	}
	chmodCmd := fmt.Sprintf("chmod -R g+rwx %s", remoteDir)
	if err := execute.OnSystemWithSudo(l, conn, chmodCmd); err != nil {
		return fmt.Errorf("failed to setup permissions on %s: %w", conn.Destination.Host, err)
	}

	// Rsync entire blobhub project so compose (including nested) and all code are on the remote.
	l.Infow("transferring blobhub project...", "from", setup.BlobhubDirPath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteBlobhubDir))
	if err := transfer(l, conn, setup.BlobhubDirPath(), remoteBlobhubDir, ".git", "node_modules"); err != nil {
		return fmt.Errorf("failed to transfer blobhub project: %w", err)
	}

	if blobhubImagePath == "" {
		return fmt.Errorf("blobhub image path missing for platform %s", setup.DockerPlatform(conn.Destination))
	}
	remoteTarPath := filepath.Join(conn.Destination.Dir, filepath.Base(blobhubImagePath))
	l.Infow("transferring blobhub image tar...", "from", blobhubImagePath, "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteTarPath))
	if err := transfer(l, conn, blobhubImagePath, remoteTarPath); err != nil {
		return fmt.Errorf("failed to transfer blobhub image tar: %w", err)
	}
	loadCmd := fmt.Sprintf("docker load -i %s", remoteTarPath)
	l.Infow("loading blobhub image(s) on remote...", "host", conn.Destination.Host)
	if err := execute.OnSystemWithSudo(l, conn, loadCmd); err != nil {
		return fmt.Errorf("failed to load blobhub image(s) on %s: %w", conn.Destination.Host, err)
	}

	chownCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteBlobhubDir)
	if err := execute.OnSystemWithSudo(l, conn, chownCmd); err != nil {
		return fmt.Errorf("failed to chown blobhub dir on %s: %w", conn.Destination.Host, err)
	}
	return nil
}
