// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

// manageDestinations handles the management of all destination nodes including
// service configuration, binary deployment and data management.
// Returns an error if management of any destination fails.
func manageDestinations(connections []connection, localBinaryPath string) error {
	l := logging.Named("Manage")

	localServiceHash, err := calculateFileHash(servicePath, nil)
	if err != nil {
		return fmt.Errorf("failed to get local service hash: %w", err)
	}

	localBinaryHash, err := calculateFileHash(localBinaryPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get local binary hash: %w", err)
	}

	for i, conn := range connections {
		if err := manage(l, conn, i, localBinaryPath, localServiceHash, localBinaryHash); err != nil {
			return fmt.Errorf("failed to manage destination %s: %w", conn.host, err)
		}
	}
	return nil
}

// manage handles the management of a single destination including service,
// binary and data management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, conn connection, nodeId int, localBinaryPath string, localServiceHash, localBinaryHash []byte) error {
	if err := manageService(l, conn, localServiceHash); err != nil {
		return fmt.Errorf("failed to manage service: %w", err)
	}

	if err := manageBinary(l, conn, localBinaryHash, localBinaryPath); err != nil {
		return fmt.Errorf("failed to manage binary: %w", err)
	}

	if err := manageData(l, conn, nodeId); err != nil {
		return fmt.Errorf("failed to manage data: %w", err)
	}

	return nil
}

// manageService handles the systemd service configuration for a destination.
// Checks if service exists, compares hashes and transfers updated service file if needed.
// Returns an error if service management fails.
func manageService(l *zap.SugaredLogger, conn connection, localHash []byte) error {
	// Check if systemd service exists
	checkCmd := "test -f /etc/systemd/system/fuelsequencerd.service"
	l.Infow("checking for fuelsequencerd service...", "host", conn.host, "cmd", checkCmd)
	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infof("no fuelsequencerd service found on %s - will transfer service file...", conn.host)
		if err := transferService(l, conn); err != nil {
			return fmt.Errorf("failed to transfer service file: %w", err)
		}
	} else {
		// Service exists, make sure it is disabled and stopped
		disableCmd := "systemctl disable fuelsequencerd"
		l.Infow("disabling fuelsequencerd...", "host", conn.host, "cmd", disableCmd)
		if err := remotely(l, conn.SSH, withSudo(disableCmd, conn.pass)); err != nil {
			return fmt.Errorf("failed to disable service on %s: %w", conn.host, err)
		}

		stopCmd := "systemctl stop fuelsequencerd"
		l.Infow("stopping fuelsequencerd...", "host", conn.host, "cmd", stopCmd)
		if err := remotely(l, conn.SSH, withSudo(stopCmd, conn.pass)); err != nil {
			return fmt.Errorf("failed to stop service on %s: %w", conn.host, err)
		}

		// Compare local and remote service file hashes
		l.Infow("comparing service file hashes...", "host", conn.host)

		// Get remote service file hash
		remoteHash, err := calculateFileHash(systemdPath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote service file hash: %w", err)
		}

		l.Debugw("hashes", "connection", conn.host, "local", localHash, "remote", remoteHash)

		// Compare and replace if different
		if string(localHash) != string(remoteHash) {
			l.Infow("service file differs! replacing...", "host", conn.host)
			if err := transferService(l, conn); err != nil {
				return fmt.Errorf("failed to transfer service file: %w", err)
			}
		}
	}
	return nil
}

// transferService transfers the systemd service file to a destination.
// Returns an error if transfer fails.
func transferService(l *zap.SugaredLogger, conn connection) error {
	// First transfer to temporary location
	tmpServicePath := filepath.Join(conn.dir, "tmp-fuelsequencerd.service")
	l.Debugw("transferring service to tmp file...", "from", servicePath, "to", fmt.Sprintf("%s:%s", conn.host, tmpServicePath))

	if err := transfer(l, conn.SSH, servicePath, tmpServicePath); err != nil {
		return fmt.Errorf("failed to transfer service file to temp location: %w", err)
	}

	// Then move to final location with sudo
	moveCmd := fmt.Sprintf("mv %s %s", tmpServicePath, systemdPath)
	if err := remotely(l, conn.SSH, withSudo(moveCmd, conn.pass)); err != nil {
		return fmt.Errorf("failed to move service file to system directory: %w", err)
	}

	l.Infow("...transferred service file to system directory", "from", servicePath, "to", fmt.Sprintf("%s:%s", conn.host, systemdPath))
	return nil
}

// manageBinary handles the binary deployment for a destination.
// Checks if binary exists, compares hashes and transfers updated binary if needed.
// Returns an error if binary management fails.
func manageBinary(l *zap.SugaredLogger, conn connection, localBinaryHash []byte, localBinaryPath string) error {
	// check if binary exists
	remoteBinaryPath := remoteBinaryPath(conn.destination)
	checkCmd := "test -f " + remoteBinaryPath
	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infow("no binary found - will transfer binary...", "host", conn.host)
		if err := transferBinary(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
			return fmt.Errorf("failed to transfer binary: %w", err)
		}
	} else {
		// if it does, compare local and remote binary hashes
		l.Debugw("comparing binary hashes...", "host", conn.host)

		// Get remote binary hash
		remoteHash, err := calculateFileHash(remoteBinaryPath, conn.SSH)
		if err != nil {
			return fmt.Errorf("failed to get remote binary hash: %w", err)
		}

		l.Debugw("hashes", "connection", conn.host, "local", localBinaryHash, "remote", remoteHash)

		// if they are different, replace the remote binary with the local one
		if string(localBinaryHash) != string(remoteHash) {
			l.Infof("binary differs on %s, replacing...", conn.host)
			if err := transferBinary(l, conn, localBinaryPath, remoteBinaryPath); err != nil {
				return fmt.Errorf("failed to transfer binary: %w", err)
			}
		}
	}
	return nil
}

// transferBinary transfers the fuelsequencerd binary to a destination.
// Returns an error if transfer fails.
func transferBinary(l *zap.SugaredLogger, conn connection, localBinaryPath, remoteBinaryPath string) error {
	logging.Infow("transferring binary...", "from", localBinaryPath, "to", fmt.Sprintf("%s:%s", conn.host, remoteBinaryPath))
	if err := transfer(l, conn.SSH, localBinaryPath, remoteBinaryPath); err != nil {
		return fmt.Errorf("failed to transfer binary to %s: %w", conn.host, err)
	}

	chownCmd := fmt.Sprintf("chown benchmarks:benchmarks_group %s", remoteBinaryPath)
	l.Infow("chowning binary to benchmarks user and group...", "host", conn.host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.pass)); err != nil {
		return fmt.Errorf("failed to chown binary on %s: %w", conn.host, err)
	}
	return nil
}

// manageData handles the chain data management for a destination.
// Cleans existing data and transfers new configuration.
// Returns an error if data management fails.
func manageData(l *zap.SugaredLogger, conn connection, nodeId int) error {
	// if data on remote exists, remove it
	homeDir := chainHomeDir(conn.destination)
	checkCmd := "test -d " + homeDir
	if err := remotely(l, conn.SSH, checkCmd); err != nil {
		l.Infof("no chain home directory found on %s", conn.host)
	} else {
		removeCmd := "rm -rf " + homeDir
		l.Infow("removing chain home directory...", "host", conn.host, "cmd", removeCmd)
		if err := remotely(l, conn.SSH, withSudo(removeCmd, conn.pass)); err != nil {
			return fmt.Errorf("failed to remove home directory on %s: %w", conn.host, err)
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
func transferConfig(l *zap.SugaredLogger, conn connection, nodeId int) error {
	instanceDir := filepath.Join(dataDir, chainName, fmt.Sprintf("fuelsequencer%d", nodeId))
	remoteDataDir := chainHomeDir(conn.destination)

	l.Infow("transferring config as chain home directory...", "from", instanceDir, "to", fmt.Sprintf("%s:%s", conn.host, remoteDataDir))
	if err := transfer(l, conn.SSH, instanceDir, remoteDataDir); err != nil {
		return fmt.Errorf("failed to transfer config to %s: %w", conn.host, err)
	}

	chownCmd := fmt.Sprintf("chown -R benchmarks:benchmarks_group %s", remoteDataDir)
	l.Infow("chowning chain home directory to benchmarks user and group...", "host", conn.host, "cmd", chownCmd)
	if err := remotely(l, conn.SSH, withSudo(chownCmd, conn.pass)); err != nil {
		return fmt.Errorf("failed to chown chain home directory on %s: %w", conn.host, err)
	}
	return nil
}
