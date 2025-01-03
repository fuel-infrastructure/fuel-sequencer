package cluster

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

func deployNetwork(connections []connection, binaryPath string) error {
	l := logging.Named("Deploy")
	binaryName := filepath.Base(binaryPath)

	for i, conn := range connections {
		if err := deployNode(l, conn, binaryName); err != nil {
			return fmt.Errorf("failed to deploy node %d: %w", i, err)
		}
	}
	return nil
}

func deployNode(l *zap.SugaredLogger, conn connection, binaryName string) error {
	remotePath := filepath.Join(conn.dir, binaryName)

	// Start the node with appropriate configuration
	cmd := fmt.Sprintf("%s --home %s start", remotePath, homeDir(conn.destination))
	if err := remotely(l, conn.SSH, cmd); err != nil {
		return fmt.Errorf("failed to start node: %w", err)
	}

	return nil
}
