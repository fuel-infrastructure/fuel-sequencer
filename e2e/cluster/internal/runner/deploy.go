// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"go.uber.org/zap"
)

// deployNetwork deploys the fuelsequencerd service to all nodes in the cluster.
// Takes a slice of connections and returns an error if any node deployment fails.
func deployNetwork(connections []System) error {
	l := logging.Named("Deploy")

	for _, operation := range []func(l *zap.SugaredLogger, id int, conn System) error{
		blobpoolStore, blobhub, node,
	} {
		for i, conn := range connections {
			if err := operation(l, i, conn); err != nil {
				return err
			}
		}
	}

	return nil
}

func blobpoolStore(l *zap.SugaredLogger, i int, conn System) error {
	if !conn.Options.Blobpool {
		return nil
	}
	if err := deployBlobpoolStore(l, conn); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

func blobhub(l *zap.SugaredLogger, i int, conn System) error {
	if !conn.Options.Blobhub {
		return nil
	}
	if err := deployBlobhub(l, conn); err != nil {
		return fmt.Errorf("failed to deploy blobhub %d: %w", i, err)
	}
	return nil
}

// deployBlobhub starts the blobhub storage using docker compose on the remote host.
// It executes the docker compose build and up commands (in detached mode) for the blobhub directory on the target node.
// Returns an error if the blobhub store fails to start.
func deployBlobhub(l *zap.SugaredLogger, conn System) error {
	composeDir := remoteBlobhubComposeDir(conn.Destination)

	// Build the containers first
	buildCmd := fmt.Sprintf("docker compose -f=%s build", composeDir)
	l.Infow("building blobhub containers...", "host", conn.Destination.Host, "cmd", buildCmd)
	if err := remotely(l, conn.SSH, withSudo(buildCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to build blobhub containers: %w", err)
	}

	// Start the containers
	upCmd := fmt.Sprintf("docker compose -f=%s up -d", composeDir)
	l.Infow("starting blobhub containers...", "host", conn.Destination.Host, "cmd", upCmd)
	if err := remotely(l, conn.SSH, withSudo(upCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobhub: %w", err)
	}
	return nil
}

// deployBlobpoolStore starts the blobpool storage using docker compose on the remote host.
// It executes the docker compose up command (in detached mode) for the blobpool compose file on the target node.
// Returns an error if the blobpool store fails to start.
func deployBlobpoolStore(l *zap.SugaredLogger, conn System) error {
	cmd := "docker compose -f=" + remoteBlobpoolComposePath(conn.Destination) + " up -d"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobpool store: %w", err)
	}
	return nil
}

func node(l *zap.SugaredLogger, i int, conn System) error {
	if !conn.Options.Sequencer {
		return nil
	}
	if err := deployNode(l, conn); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

// deployNode deploys the fuelsequencerd service to a single node.
// Enables and starts the systemd service on the remote host.
// Returns an error if service deployment fails.
func deployNode(l *zap.SugaredLogger, conn System) error {
	cmd := "systemctl enable fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start the node using the systemd service
	cmd = "systemctl start fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.Destination.Host)
	return nil
}
