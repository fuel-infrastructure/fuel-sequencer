package runner

import (
	"fmt"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageService handles the systemd service configuration for a destination.
// Checks if service exists, compares hashes and transfers updated service file if needed.
// Returns an error if service management fails.
func manageService(l *zap.SugaredLogger, sys setup.System, localHash []byte) error {
	onlyShutdown := !sys.Options.Sequencer

	// Check if systemd service exists
	checkCmd := "test -f /etc/systemd/system/fuelsequencerd.service"
	l.Infow("checking for fuelsequencerd service...", "host", sys.Destination.Host, "cmd", checkCmd)
	if err := execute.Remotely(l, sys.SSH, checkCmd); err != nil {
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
		if err := execute.Remotely(l, sys.SSH, execute.WithSudo(disableCmd, sys.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to disable service on %s: %w", sys.Destination.Host, err)
		}

		stopCmd := "systemctl stop fuelsequencerd"
		l.Infow("stopping fuelsequencerd...", "host", sys.Destination.Host, "cmd", stopCmd)
		if err := execute.Remotely(l, sys.SSH, execute.WithSudo(stopCmd, sys.Destination.Pass)); err != nil {
			return fmt.Errorf("failed to stop service on %s: %w", sys.Destination.Host, err)
		}

		if onlyShutdown {
			l.Infow("only shutdown, no need to compare hashes", "host", sys.Destination.Host)
			return nil // only shutdown, no need to compare hashes
		}

		// Compare local and remote service file hashes
		l.Infow("comparing service file hashes...", "host", sys.Destination.Host)

		// Get remote service file hash
		remoteHash, err := execute.CalculateFileHash(setup.BinaryConfig().SystemdPath, sys.SSH)
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
func transferService(l *zap.SugaredLogger, conn setup.System) error {
	// First transfer to temporary location
	tmpServicePath := filepath.Join(conn.Destination.Dir, "tmp-fuelsequencerd.service")
	l.Debugw("transferring service to tmp file...", "from", setup.ServicePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, tmpServicePath))

	if err := transfer(l, conn, setup.ServicePath(), tmpServicePath); err != nil {
		return fmt.Errorf("failed to transfer service file to temp location: %w", err)
	}

	// Then move to final location with sudo
	moveCmd := fmt.Sprintf("mv %s %s", tmpServicePath, setup.BinaryConfig().SystemdPath)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(moveCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to move service file to system directory: %w", err)
	}

	l.Infow("...transferred service file to system directory", "from", setup.ServicePath(), "to", fmt.Sprintf("%s:%s", conn.Destination.Host, setup.BinaryConfig().SystemdPath))
	return nil
}

// manageBinary handles the binary deployment for a destination.
// Checks if binary exists, compares hashes and transfers updated binary if needed.
// Returns an error if binary management fails.
func manageBinary(l *zap.SugaredLogger, conn setup.System, localBinaryHash []byte, localBinaryPath string) error {
	if !conn.Options.Sequencer {
		l.Infow("only shutdown, no need to compare hashes nor transfer binary", "host", conn.Destination.Host)
		return nil // only shutdown, no need to compare hashes nor transfer binary
	}

	// check if binary exists
	remoteBinaryPath := setup.RemoteBinaryPath(conn.Destination)
	checkCmd := "test -f " + remoteBinaryPath
	if err := execute.Remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infow("no binary found - will transfer binary...", "host", conn.Destination.Host)
		if err := transferBinary(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
			return fmt.Errorf("failed to transfer binary: %w", err)
		}
	} else {
		// if it does, compare local and remote binary hashes
		l.Debugw("comparing binary hashes...", "host", conn.Destination.Host)

		// Get remote binary hash
		remoteHash, err := execute.CalculateFileHash(setup.RemoteBinaryPath(conn.Destination), conn.SSH)
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
func transferBinary(l *zap.SugaredLogger, conn setup.System, localBinaryPath, remoteBinaryPath string) error {
	logging.Infow("transferring binary...", "from", localBinaryPath, "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteBinaryPath))
	if err := transfer(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
		return fmt.Errorf("failed to transfer binary to %s: %w", conn.Destination.Host, err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remoteBinaryPath)
	l.Infow("chowning binary to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown binary on %s: %w", conn.Destination.Host, err)
	}
	return nil
}

// manageData handles the chain data management for a destination.
// Cleans existing data and transfers new configuration.
// Returns an error if data management fails.
func manageData(l *zap.SugaredLogger, conn setup.System, nodeId int) error {
	onlyDelete := !conn.Options.Sequencer

	// if data on remote exists, remove it
	homeDir := setup.RemoteChainHomeDir(conn.Destination)
	checkCmd := "test -d " + homeDir
	if err := execute.Remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infof("no chain home directory found on %s", conn.Destination.Host)
		if onlyDelete {
			l.Infow("no chain home directory found on %s, but only delete, so no need to transfer config", conn.Destination.Host)
			return nil // only delete, found nothing, and no need to transfer config
		}
	} else {
		removeCmd := "rm -rf " + homeDir
		l.Infow("removing chain home directory...", "host", conn.Destination.Host, "cmd", removeCmd)
		if err := execute.Remotely(l, conn.SSH, execute.WithSudo(removeCmd, conn.Destination.Pass)); err != nil {
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
func transferConfig(l *zap.SugaredLogger, conn setup.System, nodeId int) error {
	instanceDir := filepath.Join(setup.DataDir(), sequencer.ChainName, fmt.Sprintf("fuelsequencer%d", nodeId))
	remoteDataDir := setup.RemoteChainHomeDir(conn.Destination)

	l.Infow("transferring config as chain home directory...", "from", instanceDir, "to", fmt.Sprintf("%s:%s", conn.Destination.Host, remoteDataDir))
	if err := transfer(l, conn, instanceDir, remoteDataDir); err != nil {
		return fmt.Errorf("failed to transfer config to %s: %w", conn.Destination.Host, err)
	}

	chownCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDataDir)
	l.Infow("chowning chain home directory to benchmarks user and group...", "host", conn.Destination.Host, "cmd", chownCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(chownCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to chown chain home directory on %s: %w", conn.Destination.Host, err)
	}
	return nil
}
