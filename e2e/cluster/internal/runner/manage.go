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

// manageSystems handles the management of all systems including
// Docker image, data management and blob file deployment.
// Returns an error if management of any system fails.
func manageSystems(systems []setup.System, localImagePaths map[string]string) error {
	l := logging.Named("Manage")

	localBlobRedisHash, err := execute.CalculateFileHash(setup.BlobRedisPath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob redis hash: %w", err)
	}

	localBlobpoolComposeHash, err := execute.CalculateFileHash(setup.BlobpoolComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob compose hash: %w", err)
	}

	for i, sys := range systems {
		platform := setup.DockerPlatform(sys.Destination)
		localImagePath := localImagePaths[platform]
		if err := manage(l, sys, i, localImagePath, localBlobRedisHash, localBlobpoolComposeHash); err != nil {
			return fmt.Errorf("failed to manage destination %s: %w", sys.Destination.Host, err)
		}
	}
	return nil
}

// manage handles the management of a single destination including Docker image,
// data and blob file management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, sys setup.System, nodeId int,
	localImagePath string, localBlobRedisHash, localBlobpoolComposeHash []byte,
) error {
	// Sequencer Related - Docker image management
	if err := manageImage(l, sys, nodeId, localImagePath); err != nil {
		return fmt.Errorf("failed to manage Docker image: %w", err)
	}

	if err := manageData(l, sys, nodeId); err != nil {
		return fmt.Errorf("failed to manage data: %w", err)
	}

	// Blobpool & Blobhub Related
	if err := manageBlobStorageRedisConf(l, sys, localBlobRedisHash); err != nil {
		return fmt.Errorf("failed to manage blobpool storage redis file: %w", err)
	}

	if err := manageBlobpoolCompose(l, sys, localBlobpoolComposeHash); err != nil {
		return fmt.Errorf("failed to manage blobpool storage compose file: %w", err)
	}

	if err := manageBlobhub(l, sys); err != nil {
		return fmt.Errorf("failed to manage blobhub storage: %w", err)
	}

	return nil
}
