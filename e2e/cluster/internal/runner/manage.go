// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// manageSystems handles the management of all systems including
// Docker image, data management and blob file deployment.
// localImagePaths is platform -> sequencer image tar path; blobhubImagePaths is platform -> blobhub image tar path.
// Returns an error if management of any system fails.
func manageSystems(systems []setup.System, localImagePaths map[string]string, blobhubImagePaths map[string]string) error {
	l := logging.Named("Manage")

	for nodeId, sys := range systems {
		// Manage blobhub only once per system (for instance 0 only)
		var blobhubImagePath string
		if blobhubImagePaths != nil {
			blobhubImagePath = blobhubImagePaths[setup.DockerPlatform(sys.Destination)]
		}
		if err := manageBlobhub(l, sys, blobhubImagePath); err != nil {
			return fmt.Errorf("failed to manage blobhub on %s: %w", sys.Destination.Host, err)
		}
		for instanceId := 0; instanceId < sys.Options.Instances; instanceId++ {
			platform := setup.DockerPlatform(sys.Destination)
			localImagePath := localImagePaths[platform]
			if err := manage(l, sys, nodeId, instanceId, localImagePath); err != nil {
				return fmt.Errorf("failed to manage destination %s instance %d: %w", sys.Destination.Host, instanceId, err)
			}
		}
	}
	return nil
}

// manage handles the management of a single destination instance including Docker image,
// data and blob file management.
// Returns an error if any management step fails.
func manage(l *zap.SugaredLogger, sys setup.System, nodeId int, instanceId int,
	localImagePath string,
) error {
	// Sequencer Related - Docker image management
	if err := manageImage(l, sys, nodeId, instanceId, localImagePath); err != nil {
		return fmt.Errorf("failed to manage Docker image: %w", err)
	}

	if err := manageData(l, sys, nodeId, instanceId); err != nil {
		return fmt.Errorf("failed to manage data: %w", err)
	}

	// Blobhub is managed separately (once per system), not per instance

	return nil
}
