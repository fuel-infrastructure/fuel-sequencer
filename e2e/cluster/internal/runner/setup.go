// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/connect"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

var logging *zap.SugaredLogger

// init initializes the package-level logger
func init() {
	logger, _ := zap.NewDevelopment()
	logging = logger.Sugar().Named("Setup")
	defer logger.Sync()
}

// logAndWrapErr logs an error message and wraps the original error with additional context
func logAndWrapErr(msg string, err error) error {
	logging.Errorw(msg, "error", err)
	return fmt.Errorf("%s: %w", msg, err)
}

// Setup orchestrates the entire cluster setup process including building binaries,
// configuring the network, establishing connections, managing destinations and deploying
// the network. Returns an error if any step fails.
func Setup(configPath string) error {
	// Load configuration first
	if err := setup.LoadConfig(configPath); err != nil {
		return logAndWrapErr("failed to load configuration", err)
	}

	sequencer.CheckParameters(logging)

	systems, err := sequencer.CheckNetworkMnemonics(logging)
	if err != nil {
		return logAndWrapErr("loaded configuration has errors", err)
	}

	// TODO: build against existing git tag

	// // Build binary
	// binaryPath, err := buildBinary()
	// if err != nil {
	// 	return logAndWrapErr("binary build failed", err)
	// }

	// build docker image and configure the network, if at least one system has sequencer enabled
	var imagePaths map[string]string
	if func() bool {
		for _, val := range systems {
			if val.Options.Sequencer {
				return true
			}
		}
		return false
	}() {
		destinations := make([]setup.Destination, len(systems))
		for i, val := range systems {
			destinations[i] = val.Destination
		}
		imagePaths, err = buildImage(logging, setup.DockerImageName(), destinations...)
		if err != nil {
			return logAndWrapErr("docker image build failed", err)
		}

		// Configure the network - generate peer IPs and instance IDs for all instances
		totalInstances := setup.TotalInstances(systems)
		peerIPs := make([]string, 0, totalInstances)
		instanceIds := make([]int, 0, totalInstances)
		var blobhubAddress string
		for _, sys := range systems {
			// Find the first system with blobhub enabled to get its address
			if sys.Options.Blobhub && blobhubAddress == "" {
				// Blobhub typically uses port 31035
				// For local deployments, use host.docker.internal so containers can reach the host
				host := sys.Destination.Host
				if setup.IsLocal(sys.Destination) {
					host = "host.docker.internal"
				}
				blobhubAddress = fmt.Sprintf("%s:31035", host)
			}
			for instanceId := 0; instanceId < sys.Options.Instances; instanceId++ {
				// For local deployments, use host.docker.internal so containers can reach each other via the host
				peerIP := sys.Destination.PeerIP
				if setup.IsLocal(sys.Destination) {
					// Replace with host.docker.internal for container-to-container communication
					peerIP = "host.docker.internal"
				}
				peerIPs = append(peerIPs, peerIP)
				instanceIds = append(instanceIds, instanceId)
			}
		}
		if err := sequencer.ConfigureNetwork(logging, setup.DataDir(), execute.Locally, peerIPs, instanceIds, blobhubAddress); err != nil {
			return logAndWrapErr("network configuration failed", err)
		}
	}

	// Build blobhub image(s) locally when at least one system has blobhub enabled
	var blobhubImagePaths map[string]string
	if func() bool {
		for _, val := range systems {
			if val.Options.Blobhub {
				return true
			}
		}
		return false
	}() {
		destinations := make([]setup.Destination, 0, len(systems))
		for _, val := range systems {
			if val.Options.Blobhub {
				destinations = append(destinations, val.Destination)
			}
		}
		var errBuild error
		blobhubImagePaths, errBuild = buildBlobhubImages(logging, destinations)
		if errBuild != nil {
			return logAndWrapErr("blobhub image build failed", errBuild)
		}
	}

	// Establish connection to all destinations
	if err := connect.EstablishConnections(logging, systems); err != nil {
		return logAndWrapErr("connection establishment failed", err)
	}
	defer connect.CloseConnections(systems)

	// Manage destinations (clean existing instances and transfer necessary files)
	if err := manageSystems(systems, imagePaths, blobhubImagePaths); err != nil {
		return logAndWrapErr("destination cleanup failed", err)
	}

	// Setup and run each node in the cluster
	if err := deployNetwork(systems); err != nil {
		return logAndWrapErr("cluster setup failed", err)
	}

	return nil
}
