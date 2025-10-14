// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/setup"
	"go.uber.org/zap"
)

// buildBinary builds or finds an existing fuelsequencerd binary.
// Returns the path to the binary and any error encountered.
// If an existing binary is found with matching architecture, it will be used.
// Otherwise, a new binary will be built using make.
func buildBinary() (string, error) {
	l := logging.Named("Build")

	binaryConfig := setup.BinaryConfig()
	makefileDir := binaryConfig.MakefileDir
	wantArch := binaryConfig.WantArch

	// Remove any existing build directory
	err := locally(l, "rm", "-rf", setup.BuildPath())
	if err != nil {
		l.Errorw("failed to remove build directory", "error", err)
		return "", fmt.Errorf("failed to remove build directory: %w", err)
	}

	// Run make target
	l.Info("running make build...")
	if err := locally(l, "make", "--directory", makefileDir, fmt.Sprintf("build-fuelsequencerd-%s", wantArch)); err != nil {
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
	binaryConfig := setup.BinaryConfig()
	buildPath := setup.BuildPath()

	// Verify binary exists
	files, err := filepath.Glob(filepath.Join(buildPath, "fuelsequencerd-*-"+binaryConfig.WantArch))
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
