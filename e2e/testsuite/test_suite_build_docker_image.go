package testsuite

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ory/dockertest/v3/docker"
)

// ensureDockerImageExists checks if the required Docker image exists locally,
// and if not, builds it from the current repository
func (s *E2ETestSuite) ensureDockerImageExists() error {
	imageTag := fmt.Sprintf("%s:%s", s.FuelSequencerDockerImageRepo, s.FuelSequencerDockerImageTag)

	s.T().Logf("checking if Docker image exists: %s", imageTag)

	// Check if image already exists locally
	if s.dockerImageExists(imageTag) {
		s.T().Logf("✅ Docker image %s already exists locally", imageTag)
		return nil
	}

	s.T().Logf("🔨 Docker image %s not found locally, building it...", imageTag)

	// Verify we're in the correct repository
	if err := s.verifyCurrentRepository(); err != nil {
		return fmt.Errorf("repository verification failed: %w", err)
	}

	// Get current commit hash
	commitHash, err := s.getCurrentCommitHash()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}

	s.T().Logf("building Docker image for commit: %s", commitHash)

	// Build the Docker image
	if err := s.buildDockerImageInTempRepo(commitHash); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	// Verify the image was built successfully
	if !s.dockerImageExists(imageTag) {
		return fmt.Errorf("Docker image %s was not found after build", imageTag)
	}

	s.T().Logf("✅ Successfully built Docker image: %s", imageTag)
	return nil
}

// dockerImageExists checks if a Docker image exists locally
func (s *E2ETestSuite) dockerImageExists(imageTag string) bool {
	images, err := s.dockerPool.Client.ListImages(docker.ListImagesOptions{})
	if err != nil {
		s.T().Logf("error listing Docker images: %v", err)
		return false
	}

	for _, image := range images {
		for _, tag := range image.RepoTags {
			if tag == imageTag {
				return true
			}
		}
	}
	return false
}

// verifyCurrentRepository checks if we're in the correct repository
func (s *E2ETestSuite) verifyCurrentRepository() error {
	// Get the remote origin URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = s.ProjectRoot
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get git remote origin: %w", err)
	}

	remoteURL := strings.TrimSpace(string(output))
	s.T().Logf("git remote origin: %s", remoteURL)

	// Check if the remote URL matches expected patterns for fuel-sequencer
	expectedPatterns := []string{
		"fuel-sequencer",
		"fuel-infrastructure/fuel-sequencer",
	}

	for _, pattern := range expectedPatterns {
		if strings.Contains(remoteURL, pattern) {
			s.T().Logf("✅ Repository verification passed: %s", remoteURL)
			return nil
		}
	}

	return fmt.Errorf("current repository (%s) does not match expected fuel-sequencer repository", remoteURL)
}

// getCurrentCommitHash gets the current commit hash
func (s *E2ETestSuite) getCurrentCommitHash() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = s.ProjectRoot
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	commitHash := strings.TrimSpace(string(output))
	if commitHash == "" {
		return "", fmt.Errorf("commit hash is empty")
	}

	return commitHash, nil
}

// buildDockerImageInTempRepo creates a temporary copy of the repo, checks out the specific commit, and builds the Docker image
func (s *E2ETestSuite) buildDockerImageInTempRepo(commitHash string) error {
	// If the requested tag is a short commit hash (8 chars), we need to checkout the full commit hash to that specific commit
	var targetCommit string
	if len(s.FuelSequencerDockerImageTag) == 8 {
		// This looks like a short commit hash, use it as the target commit
		targetCommit = s.FuelSequencerDockerImageTag
	} else {
		// Use the current commit hash
		targetCommit = commitHash
	}

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "fuel-sequencer-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			s.T().Logf("warning: failed to clean up temp directory %s: %v", tempDir, removeErr)
		}
	}()

	s.T().Logf("using temporary directory: %s", tempDir)

	// Clone the current repository to temp directory
	cloneCmd := exec.Command("git", "clone", s.ProjectRoot, tempDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout specific commit
	checkoutCmd := exec.Command("git", "checkout", targetCommit)
	checkoutCmd.Dir = tempDir
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout commit %s: %w", targetCommit, err)
	}

	s.T().Logf("checked out commit %s in temp directory", targetCommit)

	// Set a timeout for the build process
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Build Docker image using make, ensuring the tag matches what we want
	buildCmd := exec.CommandContext(ctx, "make", "build-docker-image")
	buildCmd.Dir = tempDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Env = append(os.Environ(),
		fmt.Sprintf("DOCKER_IMAGE_NAME=%s", s.FuelSequencerDockerImageRepo),
		fmt.Sprintf("DOCKER_IMAGE_TAG=%s", s.FuelSequencerDockerImageTag),
	)

	s.T().Logf("building Docker image with: make build-docker-image (target tag: %s)", s.FuelSequencerDockerImageTag)

	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("failed to build Docker image: %w", err)
	}

	s.T().Logf("✅ Docker image build completed successfully")
	return nil
}
