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

	for nodeId, conn := range connections {
		// Deploy blobhub only once per system (for instance 0 only)
		if conn.Options.Blobhub {
			if err := blobhub(l, nodeId, 0, conn); err != nil {
				return err
			}
		}
		for instanceId := 0; instanceId < conn.Options.Instances; instanceId++ {
			// Deploy blobpool for this instance
			if err := blobpoolStore(l, nodeId, instanceId, conn); err != nil {
				return err
			}
			// Deploy sequencer node for this instance
			if err := node(l, nodeId, instanceId, conn); err != nil {
				return err
			}
		}
	}

	return nil
}

func blobpoolStore(l *zap.SugaredLogger, nodeId int, instanceId int, conn setup.System) error {
	if !conn.Options.Blobpool {
		return nil
	}
	if err := deployBlobpoolStore(l, conn, instanceId); err != nil {
		return fmt.Errorf("failed to deploy blobpool for node %d instance %d: %w", nodeId, instanceId, err)
	}
	return nil
}

func blobhub(l *zap.SugaredLogger, nodeId int, instanceId int, conn setup.System) error {
	if !conn.Options.Blobhub {
		return nil
	}
	if err := deployBlobhub(l, conn, instanceId); err != nil {
		return fmt.Errorf("failed to deploy blobhub for node %d instance %d: %w", nodeId, instanceId, err)
	}
	return nil
}

// deployBlobhub starts the blobhub storage using docker compose on the remote host.
// It executes the docker compose build and up commands (in detached mode) for the blobhub directory on the target node.
// Blobhub is deployed once per system (using instanceId 0 directory) and shared across all instances.
// Returns an error if the blobhub store fails to start.
func deployBlobhub(l *zap.SugaredLogger, conn setup.System, instanceId int) error {
	// Always use instance 0 directory for blobhub (shared across all instances on this system)
	composeDir := setup.RemoteBlobhubComposeDir(conn.Destination, 0)

	// Build the containers first
	buildCmd := fmt.Sprintf("docker compose -f=%s build", composeDir)
	l.Infow("building blobhub containers...", "host", conn.Destination.Host, "instance", instanceId, "cmd", buildCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(buildCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to build blobhub containers: %w", err)
	}

	// Start the containers
	upCmd := fmt.Sprintf("docker compose -f=%s up -d", composeDir)
	l.Infow("starting blobhub containers...", "host", conn.Destination.Host, "instance", instanceId, "cmd", upCmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(upCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobhub: %w", err)
	}
	return nil
}

// deployBlobpoolStore starts the blobpool storage using docker compose on the remote host.
// It executes the docker compose up command (in detached mode) for the blobpool compose file on the target node.
// Returns an error if the blobpool store fails to start.
func deployBlobpoolStore(l *zap.SugaredLogger, conn setup.System, instanceId int) error {
	composeDir := setup.RemoteBlobDir(conn.Destination, instanceId)
	composeFile := setup.RemoteBlobpoolComposePath(conn.Destination, instanceId)
	// Use --project-directory to ensure relative paths (like ./redis.conf) in the compose file are resolved correctly
	cmd := fmt.Sprintf("docker compose --project-directory %s -f %s up -d", composeDir, composeFile)
	l.Infow("starting blobpool store...", "host", conn.Destination.Host, "instance", instanceId, "cmd", cmd)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(cmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start blobpool store: %w", err)
	}
	return nil
}

func node(l *zap.SugaredLogger, nodeId int, instanceId int, conn setup.System) error {
	if !conn.Options.Sequencer {
		return nil
	}
	if err := deployNode(l, conn, nodeId, instanceId); err != nil {
		return fmt.Errorf("failed to deploy node %d instance %d: %w", nodeId, instanceId, err)
	}
	return nil
}

// deployNode deploys the fuelsequencerd service to a single node instance using Docker.
// Uses host networking for P2P communication across nodes.
// Returns an error if deployment fails.
func deployNode(l *zap.SugaredLogger, conn setup.System, nodeId int, instanceId int) error {

	// Prepare container configuration
	containerName := setup.DockerContainerName(nodeId, instanceId)
	imageName, _ := setup.DockerImageTagLatest(conn.Destination)
	homeDir := setup.RemoteChainHomeDir(conn.Destination, instanceId)
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

	l.Infow("starting Docker container...", "host", conn.Destination.Host, "instance", instanceId, "container", containerName, "image", imageName)
	if err := execute.Remotely(l, conn.SSH, execute.WithSudo(dockerRunCmd, conn.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to start Docker container on %s: %w", conn.Destination.Host, err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.Destination.Host, "instance", instanceId, "container", containerName)
	return nil
}
