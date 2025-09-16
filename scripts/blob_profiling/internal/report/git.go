package report

import (
	"os/exec"
	"strings"
)

// CommitHash returns the current git commit hash
func CommitHash() string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// CommitHashShort returns the short version of the current git commit hash
func CommitHashShort() string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// Branch returns the current git branch
func Branch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// Status returns the git status (clean/dirty)
func Status() string {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	if len(strings.TrimSpace(string(output))) == 0 {
		return "clean"
	}
	return "dirty"
}
