// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// deployNetwork deploys the fuelsequencerd service to all nodes in the cluster.
// Takes a slice of connections and returns an error if any node deployment fails.
func deployNetwork(connections []setup.System) error {
	l := logging.Named("Deploy")

	for _, operation := range []func(l *zap.SugaredLogger, id int, conn setup.System) error{
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

func blobpoolStore(l *zap.SugaredLogger, i int, conn setup.System) error {
	if !conn.Options.Blobpool {
		return nil
	}
	if err := deployBlobpoolStore(l, conn); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

func blobhub(l *zap.SugaredLogger, i int, conn setup.System) error {
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
func deployBlobhub(l *zap.SugaredLogger, conn setup.System) error {
	composeDir := setup.RemoteBlobhubComposeDir(conn.Destination)

	// Build the containers first
	buildCmd := fmt.Sprintf("docker compose -f=%s build", composeDir)
	l.Infow("building blobhub containers...", "host", conn.Destination.Host, "cmd", buildCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(buildCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to build blobhub containers: %w", err)
	}

	// Start the containers
	upCmd := fmt.Sprintf("docker compose -f=%s up -d", composeDir)
	l.Infow("starting blobhub containers...", "host", conn.Destination.Host, "cmd", upCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(upCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobhub: %w", err)
	}
	return nil
}

// deployBlobpoolStore starts the blobpool storage using docker compose on the remote host.
// It executes the docker compose up command (in detached mode) for the blobpool compose file on the target node.
// Returns an error if the blobpool store fails to start.
func deployBlobpoolStore(l *zap.SugaredLogger, conn setup.System) error {
	cmd := "docker compose -f=" + setup.RemoteBlobpoolComposePath(conn.Destination) + " up -d"
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(cmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobpool store: %w", err)
	}
	return nil
}

func node(l *zap.SugaredLogger, i int, conn setup.System) error {
	if !conn.Options.Sequencer {
		return nil
	}
	if err := deployNode(l, conn, i); err != nil {
		return fmt.Errorf("failed to deploy node %d: %w", i, err)
	}
	return nil
}

// deployNode deploys the fuelsequencerd service to a single node using Docker.
// Uses host networking for P2P communication across nodes.
// Returns an error if deployment fails.
func deployNode(l *zap.SugaredLogger, conn setup.System, nodeId int) error {

	// Prepare container configuration
	containerName := setup.DockerContainerName(nodeId)
	imageName, _ := setup.DockerImageTagLatest(conn.Destination)
	homeDir := setup.RemoteChainHomeDir(conn.Destination)
	containerHomeDir := "/home/fuelsequencer/.fuelsequencer"

	// Get user ID for container user mapping
	userIdCmd := "id -u benchmarks"
	uid, err := execute.RemotelyWithOutput(l, conn.SSH, userIdCmd, "id")
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}
	// Map host user to container user (1000:1000 is fuelsequencer user in container)
	userMap := fmt.Sprintf("%s:1000", uid)

	// Build docker run command
	// Mount data directory, use host network for P2P communication across nodes,
	// use Docker network for local container communication, set user mapping
	dockerRunCmd := fmt.Sprintf(
		"docker run -d --name %s --network host --restart unless-stopped "+
			"-v %s:%s "+
			"--user %s "+
			"%s fuelsequencerd start",
		containerName,
		homeDir, containerHomeDir,
		userMap,
		imageName,
	)

	l.Infow("starting Docker container...", "host", conn.Destination.Host, "container", containerName, "image", imageName)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(dockerRunCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start Docker container on %s: %w", conn.Destination.Host, err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.Destination.Host, "container", containerName)
	return nil
}
