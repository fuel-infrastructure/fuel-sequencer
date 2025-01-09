package cluster

import (
	"fmt"
	"path/filepath"

	"go.uber.org/zap"
)

// func checkGitTag() error {
// 	cmd := exec.Command("git", "rev-parse", repoTag)
// 	if err := cmd.Run(); err != nil {
// 		return fmt.Errorf("tag %s not found: %w", repoTag, err)
// 	}
// 	return nil
// }

func buildBinary() (string, error) {
	l := logging.Named("Build")

	// // Checkout the specific tag
	// log.Debugf("Checking out tag: %s", repoTag)
	// cmd := exec.Command("git", "checkout", repoTag)
	// if err := cmd.Run(); err != nil {
	// 	log.Errorw("Failed to checkout tag",
	// 		"tag", repoTag,
	// 		"error", err)
	// 	return "", fmt.Errorf("failed to checkout tag: %w", err)
	// }

	binaryPath, err := findBuild(l)
	if err == nil {
		l.Infow("existing build found", "path", binaryPath)
		return binaryPath, nil
	}

	// Run make target
	l.Info("build not found, running make build...")
	if err := locally(l, "make", "--directory", makefileDir, "build-fuelsequencerd"); err != nil {
		l.Errorw("make build failed", "error", err)
		return "", fmt.Errorf("make build failed: %w", err)
	}

	// Verify binary exists
	binaryPath, err = findBuild(l)
	if err != nil {
		return "", fmt.Errorf("failed to verify build: %w", err)
	}

	l.Infow("build successful", "path", binaryPath)
	return binaryPath, nil
}

func findBuild(l *zap.SugaredLogger) (string, error) {
	// Verify binary exists
	files, err := filepath.Glob(filepath.Join(buildPath, "fuelsequencerd-*-"+wantArch))
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
