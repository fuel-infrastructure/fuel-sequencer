// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

/*
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
	err := execute.Locally(l, "rm", "-rf", setup.BuildPath())
	if err != nil {
		l.Errorw("failed to remove build directory", "error", err)
		return "", fmt.Errorf("failed to remove build directory: %w", err)
	}

	// Run make target
	l.Info("running make build...")
	if err := execute.Locally(l, "make", "--directory", makefileDir, fmt.Sprintf("build-fuelsequencerd-%s", wantArch)); err != nil {
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
*/

// buildImage builds a Docker image for a specific platform using buildx.
func buildImage(l *zap.SugaredLogger, imageName string, destinations ...setup.Destination) (map[string]string, error) {
	// Ensure buildx is available and set up
	if err := ensureBuildx(l); err != nil {
		return nil, fmt.Errorf("failed to ensure buildx: %w", err)
	}

	binaryConfig := setup.BinaryConfig()
	makefileDir := binaryConfig.MakefileDir

	// Build using buildx with platform specification
	// We need to build directly with docker buildx since make build-docker-image doesn't support platform
	l.Infow("building Docker image with buildx...", "image", imageName, "context", makefileDir)

	// Build the image using buildx
	// makefileDir should be the absolute path to the repository root (where Makefile and Dockerfile are)
	dockerfilePath := filepath.Join(makefileDir, "Dockerfile")

	// Verify Dockerfile exists at the expected location
	if err := execute.Locally(l, "test", "-f", dockerfilePath); err != nil {
		return nil, fmt.Errorf("dockerfile not found at %s: %w", dockerfilePath, err)
	}

	// Ensure makefileDir is an absolute path
	absMakefileDir, err := filepath.Abs(makefileDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for makefileDir %s: %w", makefileDir, err)
	}

	// Ensure build directory exists
	buildPath := setup.BuildPath()
	if err := execute.Locally(l, "mkdir", "-p", buildPath); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	// Get git hash for the image tag, and update the global variable
	gitHash, err := execute.LocallyWithOutput(l, "git", "describe", "--always", "--dirty", "--long")
	if err != nil {
		return nil, fmt.Errorf("failed to get git hash: %w", err)
	}
	gitHash = strings.TrimSpace(string(gitHash))
	setup.DockerImageTag(gitHash)

	absDockerfilePath, err := filepath.Abs(dockerfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for Dockerfile path %s: %w", dockerfilePath, err)
	}

	imagePaths := make(map[string]string)

	// Build with buildx
	for _, destination := range destinations {
		platform := setup.DockerPlatform(destination)
		fullImageName, platformTag := setup.DockerImageTagLatest(destination)

		if _, exists := imagePaths[platform]; exists {
			continue
		}

		if err := execute.Locally(l, "docker", "buildx", "build",
			"--platform", platform,
			"--tag", fullImageName,
			"--ssh", "default", // Use default SSH agent to access private repositories
			"--file", absDockerfilePath, // Path relative to build context
			"--load",       // Load into local Docker daemon
			absMakefileDir, // Build context - absolute path to repo root where Dockerfile is
		); err != nil {
			return nil, fmt.Errorf("docker buildx build failed: %w", err)
		}

		// Also tag as latest for this platform (extract base image name)
		imageParts := strings.Split(fullImageName, ":")
		if len(imageParts) == 2 {
			latestTag := fmt.Sprintf("%s:latest-%s", imageParts[0], strings.ReplaceAll(platform, "/", "-"))
			if err := execute.Locally(l, "docker", "tag", fullImageName, latestTag); err != nil {
				l.Warnw("failed to tag image as latest", "error", err,
					"image", fullImageName, "platform", platform, "latest_tag", latestTag)
			}
		}

		// Save image as tar file (use platform in filename to avoid conflicts)
		tarFilename := fmt.Sprintf("fuel-sequencer-image-%s.tar", platformTag)
		tarPath := filepath.Join(buildPath, tarFilename)
		l.Infow("saving Docker image to tar...", "image", fullImageName, "tar", tarPath)
		if err := execute.Locally(l, "docker", "save", fullImageName, "-o", tarPath); err != nil {
			return nil, fmt.Errorf("failed to save Docker %s image: %w", platform, err)
		}
		imagePaths[platform] = tarPath
	}

	return imagePaths, nil
}

// // Build Docker image locally for the target platform if it doesn't exist
// l.Infow("checking for Docker image locally...", "image", fullImageName, "platform", platform)
// // Check if image exists by trying to get its ID
// output, err := execute.LocallyWithOutput(l, "docker", "images", "-q", fullImageName)
// if err != nil || output == "" {
// 	l.Infow("Docker image not found locally, should be built...", "image", fullImageName, "platform", platform)
// 	if _, err := buildImage(l, fullImageName, platform); err != nil {
// 		return fmt.Errorf("failed to build Docker image for platform %s: %w", platform, err)
// 	}
// }

// ensureBuildx ensures that Docker buildx is available and a builder instance exists.
func ensureBuildx(l *zap.SugaredLogger) error {
	// Check if buildx is available
	if err := execute.Locally(l, "docker", "buildx", "version"); err != nil {
		return fmt.Errorf("docker buildx is not available: %w", err)
	}

	// Check if a builder instance exists, create one if not
	builderName := "fuel-sequencer-builder"
	output, err := execute.LocallyWithOutput(l, "docker", "buildx", "ls")
	if err != nil {
		return fmt.Errorf("failed to list buildx builders: %w", err)
	}

	// Check if builder exists (look for builderName in the output)
	builderExists := false
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, builderName) {
			builderExists = true
			break
		}
	}

	if !builderExists {
		l.Infow("creating buildx builder instance...", "builder", builderName)
		// Create a new builder instance with docker-container driver for cross-platform support
		// Use --driver docker-container for better cross-platform support
		if err := execute.Locally(l, "docker", "buildx", "create", "--name", builderName, "--driver", "docker-container", "--use"); err != nil {
			// If creation fails, try without --driver (uses default)
			if err := execute.Locally(l, "docker", "buildx", "create", "--name", builderName, "--use"); err != nil {
				return fmt.Errorf("failed to create buildx builder: %w", err)
			}
		}
		// Bootstrap the builder
		if err := execute.Locally(l, "docker", "buildx", "inspect", "--bootstrap", builderName); err != nil {
			l.Warnw("failed to bootstrap buildx builder, continuing anyway", "error", err)
		}
	} else {
		// Use the existing builder
		if err := execute.Locally(l, "docker", "buildx", "use", builderName); err != nil {
			return fmt.Errorf("failed to use buildx builder: %w", err)
		}
	}

	return nil
}
