// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes.
package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// buildBlobhubImages builds blobhub Docker image(s) locally via docker compose,
// saves them to a single tar per platform, and returns platform -> tar path.
// Blobhub is built once per destination platform (same approach as fuel-sequencer).
func buildBlobhubImages(l *zap.SugaredLogger, destinations []setup.Destination) (map[string]string, error) {
	blobhubDir := setup.BlobhubDirPath()
	composeFile := filepath.Join(blobhubDir, "docker-compose.yml")
	if err := execute.Locally(l, "test", "-f", composeFile); err != nil {
		return nil, fmt.Errorf("blobhub docker-compose.yml not found at %s: %w", composeFile, err)
	}

	buildPath := setup.BuildPath()
	if err := execute.Locally(l, "mkdir", "-p", buildPath); err != nil {
		return nil, fmt.Errorf("failed to create build directory: %w", err)
	}

	imagePaths := make(map[string]string)
	seenPlatform := make(map[string]bool)

	for _, dest := range destinations {
		platform := setup.DockerPlatform(dest)
		if seenPlatform[platform] {
			continue
		}
		seenPlatform[platform] = true

		// Build with docker compose from blobhub dir. Use DOCKER_DEFAULT_PLATFORM for cross-build when set.
		buildCmd := fmt.Sprintf("cd %s && DOCKER_DEFAULT_PLATFORM=%s docker compose -f docker-compose.yml build", quoteForShell(blobhubDir), platform)
		l.Infow("building blobhub Docker image(s)...", "platform", platform, "dir", blobhubDir)
		if err := execute.Locally(l, "sh", "-c", buildCmd); err != nil {
			return nil, fmt.Errorf("docker compose build failed for blobhub (platform %s): %w", platform, err)
		}

		// Get image IDs produced by this compose project (run from same dir).
		imagesCmd := fmt.Sprintf("cd %s && docker compose -f docker-compose.yml images -q", quoteForShell(blobhubDir))
		output, err := execute.LocallyWithOutput(l, "sh", "-c", imagesCmd)
		if err != nil {
			return nil, fmt.Errorf("failed to list blobhub compose images: %w", err)
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		var imageRefs []string
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				imageRefs = append(imageRefs, line)
			}
		}
		if len(imageRefs) == 0 {
			return nil, fmt.Errorf("blobhub compose produced no images (platform %s)", platform)
		}

		// Save all images to one tar (docker save accepts multiple refs).
		platformTag := strings.ReplaceAll(platform, "/", "-")
		tarFilename := fmt.Sprintf("blobhub-image-%s.tar", platformTag)
		tarPath := filepath.Join(buildPath, tarFilename)
		saveArgs := append([]string{"save", "-o", tarPath}, imageRefs...)
		l.Infow("saving blobhub Docker image(s) to tar...", "tar", tarPath, "count", len(imageRefs))
		if err := execute.Locally(l, "docker", saveArgs...); err != nil {
			return nil, fmt.Errorf("failed to save blobhub image(s) to %s: %w", tarPath, err)
		}
		imagePaths[platform] = tarPath
	}

	return imagePaths, nil
}

// quoteForShell returns a path safe for use inside double-quoted shell strings.
// For simple paths (no single quote) we use single-quoted path so cd 'path' works.
func quoteForShell(path string) string {
	if strings.Contains(path, "'") {
		return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
	}
	return "'" + path + "'"
}
