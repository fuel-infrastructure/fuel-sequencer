// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
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
func Setup() error {
	sequencer.CheckParameters(logging)

	ld, lm := len(destinations), len(sequencer.Mnemonics)
	if ld != lm {
		logging.Fatalw("check config: number of destinations (%d) does not match number of mnemonics (%d)", ld, lm)
	}

	// TODO: build against existing git tag

	// Build binary
	binaryPath, err := buildBinary()
	if err != nil {
		return logAndWrapErr("binary build failed", err)
	}

	// Configure the network
	peerIPs := make([]string, len(destinations))
	for i, val := range destinations {
		peerIPs[i] = val.peer_ip
	}
	if err := sequencer.ConfigureNetwork(logging, dataDir, locally, peerIPs); err != nil {
		return logAndWrapErr("network configuration failed", err)
	}

	// Establish connection to all destinations
	connections, err := establishConnections()
	if err != nil {
		return logAndWrapErr("connection establishment failed", err)
	}
	defer closeConnections(connections)

	// Manage destinations (clean existing instances and transfer necessary files)
	if err := manageDestinations(connections, binaryPath); err != nil {
		return logAndWrapErr("destination cleanup failed", err)
	}

	// Setup and run each node in the cluster
	if err := deployNetwork(connections); err != nil {
		return logAndWrapErr("cluster setup failed", err)
	}

	return nil
}
