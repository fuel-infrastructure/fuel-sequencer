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
// service configuration, binary deployment, data management and blob file deployment.
// Returns an error if management of any system fails.
func manageSystems(systems []setup.System, localBinaryPath string) error {
	l := logging.Named("Manage")

	localServiceHash, err := execute.CalculateFileHash(setup.ServicePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local service hash: %w", err)
	}

	localBinaryHash, err := execute.CalculateFileHash(localBinaryPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get local binary hash: %w", err)
	}

	localBlobRedisHash, err := execute.CalculateFileHash(setup.BlobRedisPath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob redis hash: %w", err)
	}

	localBlobpoolComposeHash, err := execute.CalculateFileHash(setup.BlobpoolComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob compose hash: %w", err)
	}

	for i, sys := range systems {
		if err := manage(l, sys, i, localBinaryPath, localServiceHash, localBinaryHash,
			localBlobRedisHash, localBlobpoolComposeHash); err != nil {
			return fmt.Errorf("failed to manage destination %s: %w", sys.Destination.Host, err)
		}
	}
	return nil
}

// manage handles the management of a single destination including service,
// binary, data and blob file management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, sys setup.System, nodeId int,
	localBinaryPath string, localServiceHash, localBinaryHash,
	localBlobRedisHash, localBlobpoolComposeHash []byte,
) error {
	// Sequencer Related
	if err := manageService(l, sys, localServiceHash); err != nil {
		return fmt.Errorf("failed to manage service: %w", err)
	}

	if err := manageBinary(l, sys, localBinaryHash, localBinaryPath); err != nil {
		return fmt.Errorf("failed to manage binary: %w", err)
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
