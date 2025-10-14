// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"go.uber.org/zap"
)

// manageSystems handles the management of all systems including
// service configuration, binary deployment, data management and blob file deployment.
// Returns an error if management of any system fails.
func manageSystems(systems []System, localBinaryPath string) error {
	l := logging.Named("Manage")

	localServiceHash, err := calculateFileHash(ServicePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local service hash: %w", err)
	}

	localBinaryHash, err := calculateFileHash(localBinaryPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get local binary hash: %w", err)
	}

	localBlobRedisHash, err := calculateFileHash(BlobRedisPath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob redis hash: %w", err)
	}

	localBlobpoolComposeHash, err := calculateFileHash(BlobpoolComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blob compose hash: %w", err)
	}

	localBlobhubComposeHash, err := calculateFileHash(BlobhubComposePath(), nil)
	if err != nil {
		return fmt.Errorf("failed to get local blobhub compose hash: %w", err)
	}

	for i, sys := range systems {
		if err := manage(l, sys, i, localBinaryPath, localServiceHash, localBinaryHash,
			localBlobRedisHash, localBlobpoolComposeHash, localBlobhubComposeHash); err != nil {
			return fmt.Errorf("failed to manage destination %s: %w", sys.Destination.Host, err)
		}
	}
	return nil
}

// manage handles the management of a single destination including service,
// binary, data and blob file management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, sys System, nodeId int,
	localBinaryPath string, localServiceHash, localBinaryHash,
	localBlobRedisHash, localBlobpoolComposeHash, localBlobhubComposeHash []byte,
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

	if err := manageBlobhubCompose(l, sys, localBlobhubComposeHash); err != nil {
		return fmt.Errorf("failed to manage blobhub storage compose file: %w", err)
	}

	return nil
}
