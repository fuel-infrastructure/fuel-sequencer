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
// Build is skipped when images already exist for the current blob-storage git state
// (same skip-if-exists approach as fuel-sequencer in build.go).
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

	// Git version from blob-storage repo for content-based image tag (skip build when unchanged)
	gitHash, err := execute.LocallyWithOutput(l, "git", "-C", blobhubDir, "describe", "--always", "--dirty", "--long")
	if err != nil {
		return nil, fmt.Errorf("failed to get blobhub git version at %s: %w", blobhubDir, err)
	}
	gitHash = strings.TrimSpace(string(gitHash))
	// Sanitize for use in docker tag (allow [a-zA-Z0-9_.-])
	gitHash = strings.ReplaceAll(gitHash, "/", "-")
	gitHash = strings.ReplaceAll(gitHash, ":", "-")

	builtImageBaseNames, err := getBlobhubBuiltImageNames(l, blobhubDir)
	if err != nil {
		return nil, err
	}
	if len(builtImageBaseNames) == 0 {
		return nil, fmt.Errorf("blobhub compose produced no built images")
	}

	imagePaths := make(map[string]string)
	seenPlatform := make(map[string]bool)

	for _, dest := range destinations {
		platform := setup.DockerPlatform(dest)
		if seenPlatform[platform] {
			continue
		}
		seenPlatform[platform] = true

		platformTag := strings.ReplaceAll(platform, "/", "-")
		versionTag := gitHash + "-" + platformTag
		versionedRefs := make([]string, len(builtImageBaseNames))
		for i, base := range builtImageBaseNames {
			versionedRefs[i] = base + ":" + versionTag
		}

		// Skip build if all versioned images already exist (same pattern as fuel-sequencer build.go)
		allExist := true
		for _, ref := range versionedRefs {
			if _, err := execute.LocallyWithOutput(l, "docker", "image", "inspect", ref); err != nil {
				allExist = false
				break
			}
		}
		if allExist {
			l.Infow("blobhub image(s) already exist, skipping build", "platform", platform, "version", versionTag)
		} else {
			// Build with docker compose from blobhub dir. Use DOCKER_DEFAULT_PLATFORM for cross-build when set.
			// DOCKER_BUILDKIT=1 and compose build with ssh: default allow private Go modules (e.g. fuel-sequencer) via SSH.
			buildCmd := fmt.Sprintf("cd %s && DOCKER_BUILDKIT=1 DOCKER_DEFAULT_PLATFORM=%s docker compose -f docker-compose.yml build --ssh default", quoteForShell(blobhubDir), platform)
			l.Infow("building blobhub Docker image(s)...", "platform", platform, "dir", blobhubDir)
			if err := execute.Locally(l, "sh", "-c", buildCmd); err != nil {
				return nil, fmt.Errorf("docker compose build failed for blobhub (platform %s): %w", platform, err)
			}
			// Tag built images with version so next run can skip build
			for _, base := range builtImageBaseNames {
				latestRef := base + ":latest"
				versionedRef := base + ":" + versionTag
				if err := execute.Locally(l, "docker", "tag", latestRef, versionedRef); err != nil {
					l.Warnw("failed to tag blobhub image with version", "error", err, "from", latestRef, "to", versionedRef)
				}
			}
		}

		// Ensure :latest tags exist so tar works with remote compose (expects :latest)
		for _, base := range builtImageBaseNames {
			versionedRef := base + ":" + versionTag
			latestRef := base + ":latest"
			_ = execute.Locally(l, "docker", "tag", versionedRef, latestRef)
		}
		// Save with :latest so remote docker compose up finds the images
		latestRefs := make([]string, len(builtImageBaseNames))
		for i, base := range builtImageBaseNames {
			latestRefs[i] = base + ":latest"
		}
		tarFilename := fmt.Sprintf("blobhub-image-%s.tar", platformTag)
		tarPath := filepath.Join(buildPath, tarFilename)
		saveArgs := append([]string{"save", "-o", tarPath}, latestRefs...)
		l.Infow("saving blobhub Docker image(s) to tar...", "tar", tarPath, "count", len(latestRefs))
		if err := execute.Locally(l, "docker", saveArgs...); err != nil {
			return nil, fmt.Errorf("failed to save blobhub image(s) to %s: %w", tarPath, err)
		}
		imagePaths[platform] = tarPath
	}

	return imagePaths, nil
}

// getBlobhubBuiltImageNames returns base image names (no tag) for services built by blobhub compose.
func getBlobhubBuiltImageNames(l *zap.SugaredLogger, blobhubDir string) ([]string, error) {
	configCmd := fmt.Sprintf("cd %s && docker compose -f docker-compose.yml config --images", quoteForShell(blobhubDir))
	output, err := execute.LocallyWithOutput(l, "sh", "-c", configCmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list blobhub compose image names: %w", err)
	}
	var baseNames []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		// Only include local-built images (no registry slash)
		if !strings.Contains(name, "/") {
			if idx := strings.Index(name, ":"); idx != -1 {
				name = name[:idx]
			}
			baseNames = append(baseNames, name)
		}
	}
	return baseNames, nil
}

// quoteForShell returns a path safe for use inside double-quoted shell strings.
// For simple paths (no single quote) we use single-quoted path so cd 'path' works.
func quoteForShell(path string) string {
	if strings.Contains(path, "'") {
		return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
	}
	return "'" + path + "'"
}
