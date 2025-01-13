// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

// buildBinary builds or finds an existing fuelsequencerd binary.
// Returns the path to the binary and any error encountered.
// If an existing binary is found with matching architecture, it will be used.
// Otherwise, a new binary will be built using make.
func buildBinary() (string, error) {
	l := logging.Named("Build")

	// Remove any existing build directory
	err := locally(l, "rm", "-rf", buildPath)
	if err != nil {
		l.Errorw("failed to remove build directory", "error", err)
		return "", fmt.Errorf("failed to remove build directory: %w", err)
	}

	// Run make target
	l.Info("running make build...")
	if err := locally(l, "make", "--directory", makefileDir, "build-fuelsequencerd"); err != nil {
		l.Errorw("make build failed", "error", err)
		return "", fmt.Errorf("make build failed: %w", err)
	}

	// Verify binary exists
	binaryPath, err := findBuild(l)
	if err != nil {
		return "", fmt.Errorf("failed to verify build: %w", err)
	}

	l.Infow("build successful", "path", binaryPath)
	return binaryPath, nil
}

// findBuild looks for an existing fuelsequencerd binary in the build directory
// matching the desired architecture. Returns the path to the binary if found
// or an error if not found or multiple matches exist.
func findBuild(l *zap.SugaredLogger) (string, error) {
	// Verify binary exists
	files, err := filepath.Glob(filepath.Join(buildPath, "fuelsequencerd-*-"+wantArch))
	if err != nil {
		l.Errorw("failed to find binary", "error", err)
		return "", fmt.Errorf("failed to find binary: %w", err)
	}
	if len(files) != 1 {
		l.Errorw("desired binary not found", "path", buildPath, "found", files)
		return "", fmt.Errorf("desired binary not found: path: %s; found: %+v", buildPath, files)
	}
	binaryPath := files[0]

	return binaryPath, nil
}
