package runner

import (
	"fmt"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageImage handles the Docker image for a destination.
// Builds the image locally for the specified platform if needed, then transfers and loads it on the remote host.
// Returns an error if Docker image management fails.
func manageImage(l *zap.SugaredLogger, sys setup.System, nodeId int, imagePath string) error {
	onlyShutdown := !sys.Options.Sequencer

	platform := setup.DockerPlatform(sys.Destination)
	fullImageName, _ := setup.DockerImageTagLatest(sys.Destination)

	// Stop and remove existing containers if they exist
	containerName := setup.DockerContainerName(nodeId)
	stopCmd := fmt.Sprintf("docker stop %s 2>/dev/null || true", containerName)
	l.Infow("stopping existing container...", "host", sys.Destination.Host, "container", containerName)
	if err := execute.Remotely(l, sys.SSH, execute.WithSudo(stopCmd, sys.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to stop container on %s: %w", sys.Destination.Host, err)
	}

	removeCmd := fmt.Sprintf("docker rm %s 2>/dev/null || true", containerName)
	l.Infow("removing existing container...", "host", sys.Destination.Host, "container", containerName)
	if err := execute.Remotely(l, sys.SSH, execute.WithSudo(removeCmd, sys.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to remove container on %s: %w", sys.Destination.Host, err)
	}

	if onlyShutdown {
		l.Infow("only shutdown, no need to transfer Docker image", "host", sys.Destination.Host)
		return nil
	}

	// Build Docker image locally for the target platform if it doesn't exist
	l.Infow("checking for Docker image locally...", "image", fullImageName, "platform", platform)
	// Check if image exists by trying to get its ID
	output, err := execute.LocallyWithOutput(l, "docker", "images", "-q", fullImageName)
	if err != nil || output == "" {
		l.Errorw("Docker image not found locally, should be already built", "image", fullImageName, "platform", platform, "error", err)
		return fmt.Errorf("docker image not found locally, should be already built: %w", err)
	}

	// Transfer image tar file to remote
	remoteTarPath := filepath.Join(sys.Destination.Dir, filepath.Base(imagePath))
	l.Infow("transferring Docker image tar...", "from", imagePath, "to", fmt.Sprintf("%s:%s", sys.Destination.Host, remoteTarPath))
	if err := transfer(l, sys, imagePath, remoteTarPath); err != nil {
		return fmt.Errorf("failed to transfer Docker image tar: %w", err)
	}

	// Load image on remote
	l.Infow("loading Docker image on remote...", "host", sys.Destination.Host)
	loadCmd := fmt.Sprintf("docker load -i %s", remoteTarPath)
	if err := execute.Remotely(l, sys.SSH, execute.WithSudo(loadCmd, sys.Destination.Pass)); err != nil {
		return fmt.Errorf("failed to load Docker image on %s: %w", sys.Destination.Host, err)
	}

	return nil
}
