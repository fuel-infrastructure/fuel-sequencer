// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"fmt"

	"go.uber.org/zap"
)

// deployNetwork deploys the fuelsequencerd service to all nodes in the cluster.
// Takes a slice of connections and returns an error if any node deployment fails.
func deployNetwork(connections []connection) error {
	l := logging.Named("Deploy")

	for i, conn := range connections {
		if err := deployNode(l, conn); err != nil {
			return fmt.Errorf("failed to deploy node %d: %w", i, err)
		}
	}
	return nil
}

// deployNode deploys the fuelsequencerd service to a single node.
// Enables and starts the systemd service on the remote host.
// Returns an error if service deployment fails.
func deployNode(l *zap.SugaredLogger, conn connection) error {
	cmd := "systemctl enable fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.pass)); err != nil {
		return fmt.Errorf("failed to enable service: %w", err)
	}

	// Start the node using the systemd service
	cmd = "systemctl start fuelsequencerd"
	if err := remotely(l, conn.SSH, withSudo(cmd, conn.pass)); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.host)
	return nil
}
