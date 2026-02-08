// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes.
package runner

import (
	"fmt"
	"os"
	"strings"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/execute"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// resolveVersions resolves sequencer_version and blob_storage_version from config:
// empty or "current" => use makefile_dir / blobhub_dir_path as-is.
// branch or commit => copy repo to a temp dir, checkout there, build from it; cleanups remove temp dirs when Setup returns.
func resolveVersions(l *zap.SugaredLogger) (cleanups []func(), err error) {
	cfg := setup.LoadedConfig()
	if cfg == nil {
		return nil, nil
	}

	var sequencerDir, blobDir string
	defer func() {
		if err != nil {
			// On error, run any cleanups we already registered
			for _, f := range cleanups {
				f()
			}
		}
	}()

	if !useCurrentVersion(cfg.Binary.SequencerVersion) {
		dir, cleanup, err := copyRepoAndCheckout(l, cfg.Binary.MakefileDir, cfg.Binary.SequencerVersion, "fuel-sequencer", true)
		if err != nil {
			return nil, fmt.Errorf("sequencer version %q: %w", cfg.Binary.SequencerVersion, err)
		}
		sequencerDir = dir
		cleanups = append(cleanups, cleanup)
		l.Infow("building sequencer from version in temp dir", "version", cfg.Binary.SequencerVersion, "dir", dir)
	}

	if !useCurrentVersion(cfg.Blob.BlobStorageVersion) {
		dir, cleanup, err := copyRepoAndCheckout(l, cfg.Blob.BlobhubDirPath, cfg.Blob.BlobStorageVersion, "blob-storage", false)
		if err != nil {
			return nil, fmt.Errorf("blob-storage version %q: %w", cfg.Blob.BlobStorageVersion, err)
		}
		blobDir = dir
		cleanups = append(cleanups, cleanup)
		l.Infow("building blob-storage from version in temp dir", "version", cfg.Blob.BlobStorageVersion, "dir", dir)
	}

	setup.SetResolvedPaths(sequencerDir, blobDir)
	return cleanups, nil
}

func useCurrentVersion(v string) bool {
	return v == "" || strings.TrimSpace(strings.ToLower(v)) == "current"
}

// copyRepoAndCheckout copies the git repo to a temp dir, checks out ref there, and returns the temp path
// and a cleanup func that clears the resolved path and removes the temp dir.
// isSequencer: true => cleanup clears resolved sequencer path; false => clears resolved blob path.
func copyRepoAndCheckout(l *zap.SugaredLogger, repoDir, ref, namePrefix string, isSequencer bool) (tempDir string, cleanup func(), err error) {
	if err := execute.Locally(l, "git", "-C", repoDir, "fetch", "origin"); err != nil {
		return "", nil, fmt.Errorf("git fetch: %w", err)
	}

	dir, err := os.MkdirTemp("", namePrefix+"-e2e-*")
	if err != nil {
		return "", nil, fmt.Errorf("mkdir temp: %w", err)
	}

	// Clone from local repo (shares objects, fast) so we can checkout without touching the original
	if err := execute.Locally(l, "git", "clone", "--local", repoDir, dir); err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("git clone --local: %w", err)
	}

	if err := execute.Locally(l, "git", "-C", dir, "checkout", "--force", ref); err != nil {
		os.RemoveAll(dir)
		return "", nil, fmt.Errorf("git checkout %s: %w", ref, err)
	}

	cleanup = func() {
		if isSequencer {
			setup.ClearResolvedSequencerDir()
		} else {
			setup.ClearResolvedBlobhubDir()
		}
		_ = os.RemoveAll(dir)
	}
	return dir, cleanup, nil
}
