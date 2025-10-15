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

	systems, err := setup.CheckParameters(logging)
	if err != nil {
		return logAndWrapErr("loaded configuration has errors", err)
	}

	// TODO: build against existing git tag

	// Build binary
	binaryPath, err := buildBinary()
	if err != nil {
		return logAndWrapErr("binary build failed", err)
	}

	// Configure the network
	peerIPs := make([]string, len(systems))
	for i, val := range systems {
		peerIPs[i] = val.Destination.PeerIP
	}
	if err := sequencer.ConfigureNetwork(logging, setup.DataDir(), execute.Locally, peerIPs); err != nil {
		return logAndWrapErr("network configuration failed", err)
	}

	// Establish connection to all destinations
	if err := connect.EstablishConnections(logging, systems); err != nil {
		return logAndWrapErr("connection establishment failed", err)
	}
	defer connect.CloseConnections(systems)

	// Manage destinations (clean existing instances and transfer necessary files)
	if err := manageSystems(systems, binaryPath); err != nil {
		return logAndWrapErr("destination cleanup failed", err)
	}

	// Setup and run each node in the cluster
	if err := deployNetwork(systems); err != nil {
		return logAndWrapErr("cluster setup failed", err)
	}

	return nil
}
