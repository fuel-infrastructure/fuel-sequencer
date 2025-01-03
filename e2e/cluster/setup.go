package cluster

import (
	"fmt"

	"go.uber.org/zap"
)

var logging *zap.SugaredLogger

func init() {
	logger, _ := zap.NewDevelopment()
	logging = logger.Sugar().Named("Setup")
	defer logger.Sync()
}

func logAndWrapErr(msg string, err error) error {
	logging.Errorw(msg, "error", err)
	return fmt.Errorf("%s: %w", msg, err)
}

func Setup() error {

	ld, lm := len(destinations), len(mnemonics)
	if ld != lm {
		logging.Fatalw("check config: number of destinations (%d) does not match number of mnemonics (%d)", ld, lm)
	}

	// // Ensure git repository tag exists
	// if err := checkGitTag(); err != nil {
	// 	return logAndWrapErr("git tag check failed", err)
	// }

	// Build binary
	binaryPath, err := buildBinary()
	if err != nil {
		return logAndWrapErr("binary build failed", err)
	}

	// Configure the network
	if err := configureNetwork(); err != nil {
		return logAndWrapErr("network configuration failed", err)
	}

	// Establish connection to all destinations
	connections, err := establishConnections()
	if err != nil {
		return logAndWrapErr("connection establishment failed", err)
	}
	defer closeConnections(connections)

	// // Ensure destinations are clean (no existing instance running)
	// if err := cleanDestinations(connections); err != nil {
	// 	return logAndWrapErr("destination cleanup failed", err)
	// }

	// Transfer binary to all destinations
	if err := transferFiles(connections, binaryPath); err != nil {
		return logAndWrapErr("binary transfer failed", err)
	}

	// Setup each node in the cluster, using E2ETestSuite
	if err := deployNetwork(connections, binaryPath); err != nil {
		return logAndWrapErr("cluster setup failed", err)
	}

	return nil
}
