package report

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// BlobStorageVersion returns the blob-storage version from go.mod
func BlobStorageVersion() string {
	// Try to read go.mod file
	file, err := os.Open("go.mod")
	if err != nil {
		return "unknown"
	}
	defer file.Close()

	// Look for blob-storage version in go.mod
	scanner := bufio.NewScanner(file)
	blobStorageRegex := regexp.MustCompile(`github\.com/fuel-infrastructure/blob-storage\s+v([^\s]+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := blobStorageRegex.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}

	return "unknown"
}

// BlobStorageCommitHash extracts the commit hash from the blob-storage version
func BlobStorageCommitHash() string {
	version := BlobStorageVersion()
	if version == "unknown" {
		return "unknown"
	}

	// Extract commit hash from version string like "v0.0.0-20250919082051-2fc207358bc3"
	// The commit hash is the last part after the second dash
	parts := strings.Split(version, "-")
	if len(parts) >= 3 {
		return parts[len(parts)-1]
	}

	return "unknown"
}

// BlobStorageCommitHashShort returns the short version of the blob-storage commit hash
func BlobStorageCommitHashShort() string {
	commitHash := BlobStorageCommitHash()
	if commitHash == "unknown" || len(commitHash) < 8 {
		return "unknown"
	}
	return commitHash[:8]
}
