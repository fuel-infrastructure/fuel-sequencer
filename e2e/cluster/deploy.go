package cluster

import (
	"fmt"

	"go.uber.org/zap"
)

func deployNetwork(connections []connection) error {
	l := logging.Named("Deploy")

	for i, conn := range connections {
		if err := deployNode(l, conn); err != nil {
			return fmt.Errorf("failed to deploy node %d: %w", i, err)
		}
	}
	return nil
}

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
	return nil
}
