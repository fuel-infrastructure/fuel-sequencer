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
func deployNetwork(connections []connection) error {
	l := logging.Named("Deploy")

	for _, operation := range []func(l *zap.SugaredLogger, id int, conn connection) error{
		blobpoolStore, node,
	} {
		for i, conn := range connections {
			if err := operation(l, i, conn); err != nil {
				return err
			}
		}
	}

	return nil
}

func blobpoolStore(l *zap.SugaredLogger, i int, conn connection) error {
	if err := deployBlobpoolStore(l, conn); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

// deployBlobpoolStore starts the blobpool storage using docker compose on the remote host.
// It executes the docker compose up command (in detached mode) for the blobpool compose file on the target node.
// Returns an error if the blobpool store fails to start.
func deployBlobpoolStore(l *zap.SugaredLogger, conn connection) error {
	cmd := "docker compose -f=" + remoteBlobpoolComposePath(conn.destination) + " up -d"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.destination.pass)); err != nil {
		return fmt.Errorf("failed to start blobpool store: %w", err)
	}
	return nil
}

func node(l *zap.SugaredLogger, i int, conn connection) error {
	if err := deployNode(l, conn); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

// deployNode deploys the fuelsequencerd service to a single node.
// Enables and starts the systemd service on the remote host.
// Returns an error if service deployment fails.
func deployNode(l *zap.SugaredLogger, conn connection) error {
	cmd := "systemctl enable fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.destination.pass)); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start the node using the systemd service
	cmd = "systemctl start fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.destination.pass)); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.destination.host)
	return nil
}
