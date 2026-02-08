// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"strings"

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
			// Deploy sequencer node for this instance
			if err := node(l, nodeId, instanceId, conn); err != nil {
				return err
			}
		}
	}

	return nil
}

func blobhub(l *zap.SugaredLogger, nodeId int, instanceId int, conn setup.System) error {
	if !conn.Options.Blobhub {
		return nil
	}

	var composeDir string
	if conn.IsLocal {
		composeDir = setup.LocalBlobhubComposeDir()
	} else {
		// Always use instance 0 directory for blobhub (shared across all instances on this system)
		composeDir = setup.RemoteBlobhubComposeDir(conn.Destination, 0)
	}

	if err := deployBlobhub(l, conn, composeDir, instanceId); err != nil {
		return fmt.Errorf("failed to deploy blobhub for node %d instance %d: %w", nodeId, instanceId, err)
	}
	return nil
}

// deployBlobhub starts the blobhub storage using docker compose on the target node.
// Images are already built locally and loaded on the host during manage; this only runs compose up.
// Blobhub is deployed once per system (using instanceId 0 directory) and shared across all instances.
// Returns an error if the blobhub store fails to start.
func deployBlobhub(l *zap.SugaredLogger, conn setup.System, composeDir string, instanceId int) error {
	upCmd := fmt.Sprintf("docker compose -f=%s up -d", composeDir)
	l.Infow("starting blobhub containers...", "host", conn.Destination.Host, "instance", instanceId, "cmd", upCmd)
	if err := execute.OnSystemWithSudo(l, conn, upCmd); err != nil {
		return fmt.Errorf("failed to start blobhub: %w", err)
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
// Uses host networking for remote deployments (Linux) and port mapping for local deployments
// (macOS Docker Desktop doesn't support host networking).
// Returns an error if deployment fails.
func deployNode(l *zap.SugaredLogger, conn setup.System, nodeId int, instanceId int) error {

	// Prepare container configuration
	containerName := setup.DockerContainerName(nodeId, instanceId)
	imageName, _ := setup.DockerImageTagLatest(conn.Destination)
	homeDir := setup.RemoteChainHomeDir(conn.Destination, instanceId)
	containerHomeDir := "/home/fuelsequencer/.fuelsequencer"

	dockerArgs := []string{
		"--name", containerName,
		"--restart", "unless-stopped",
		"-v", fmt.Sprintf("%s:%s", homeDir, containerHomeDir),
		"-w", containerHomeDir,
	}

	// For local deployments (macOS), use port mapping instead of host networking
	// Docker Desktop on macOS doesn't support --network host
	if conn.IsLocal {
		ports := setup.InstancePorts(instanceId)
		// Map all required ports: host:container
		dockerArgs = append(dockerArgs,
			"-p", fmt.Sprintf("%d:%d", ports.P2P, ports.P2P),
			"-p", fmt.Sprintf("%d:%d", ports.RPC, ports.RPC),
			"-p", fmt.Sprintf("%d:%d", ports.API, ports.API),
			"-p", fmt.Sprintf("%d:%d", ports.GRPC, ports.GRPC),
			"-p", fmt.Sprintf("%d:%d", ports.Prometheus, ports.Prometheus),
			"-p", fmt.Sprintf("%d:%d", ports.BlobpoolServer, ports.BlobpoolServer),
		)
	} else {
		// For remote deployments (Linux), use host networking
		dockerArgs = append(dockerArgs, "--network", "host")
	}

	// Get user ID for container user mapping
	if !conn.IsLocal {
		userIdCmd := "id -u benchmarks"
		uid, err := execute.OnSystemWithOutput(l, conn, userIdCmd, "id")
		if err != nil {
			return fmt.Errorf("failed to get user ID: %w", err)
		}
		// Map host user to container user (1000:1000 is fuelsequencer user in container)
		userMap := fmt.Sprintf("%s:1000", uid)

		dockerArgs = append(dockerArgs, "--user", userMap)
	}

	dockerRunCmd := fmt.Sprintf(
		"docker run -d %s %s fuelsequencerd start",
		strings.Join(dockerArgs, " "),
		imageName,
	)

	l.Infow("starting Docker container...", "host", conn.Destination.Host, "instance", instanceId, "container", containerName, "image", imageName)
	if err := execute.OnSystemWithSudo(l, conn, dockerRunCmd); err != nil {
		return fmt.Errorf("failed to start Docker container on %s: %w", conn.Destination.Host, err)
	}

	// TODO: Check if node is ready
	l.Infow("deployed node", "host", conn.Destination.Host, "instance", instanceId, "container", containerName)
	return nil
}
