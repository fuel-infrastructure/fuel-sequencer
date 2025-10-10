// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// executor defines an interface for executing commands either locally or remotely
type executor interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

// execute runs a command through the provided executor, capturing and logging output.
// It streams stdout and stderr through the logger and returns any error encountered.
func execute(l *zap.SugaredLogger, e executor, commandName string) error {

	// Capture output for logging by creating pipes for stdout and stderr
	stdout, err := e.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := e.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create a wait group to ensure both goroutines complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Stream stdout with smart progress detection
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if isProgressLine(line) {
				// This is a progress update, log it as progress info
				l.Infow(fmt.Sprintf("running %s...", commandName), "progress", strings.TrimSpace(line))
			} else if strings.TrimSpace(line) != "" {
				// Regular output line (skip empty lines)
				l.Info(line)
			}
		}
	}()

	// Stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				l.Warn(line)
			}
		}
	}()

	// Start the command
	if err := e.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for both output streams to complete
	wg.Wait()

	// Wait for the command to complete
	if err := e.Wait(); err != nil {
		l.Errorw("command failed", "error", err)
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// locally executes a command on the local machine with the given name and arguments.
// Returns an error if the command fails.
func locally(l *zap.SugaredLogger, name string, args ...string) error {
	return execute(l.Named("CMD"), exec.Command(name, args...), name)
}

func locallyWithCustomName(l *zap.SugaredLogger, name, cmdname string, args ...string) error {
	return execute(l.Named("CMD"), exec.Command(name, args...), cmdname)
}

// remotely executes a command on a remote machine through an SSH connection.
// Returns an error if the command fails.
func remotely(l *zap.SugaredLogger, client *ssh.Client, cmd string) error {
	cmdname := strings.Split(cmd, " ")[0]
	return remotelyWithCustomName(l, client, cmd, cmdname)
}

func remotelyWithCustomName(l *zap.SugaredLogger, client *ssh.Client, cmd, cmdname string) error {
	// Create new SSH client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return execute(l.Named("SSH"), &sshSessionExecutor{Session: session, cmd: cmd}, cmdname)
}

// sshSessionExecutor wraps an SSH session to implement the executor interface
type sshSessionExecutor struct {
	*ssh.Session
	cmd string
}

// Start begins execution of the SSH command
func (s *sshSessionExecutor) Start() error {
	return s.Session.Start(s.cmd)
}

// ioCloser adapts an io.Reader to io.ReadCloser
type ioCloser struct {
	io.Reader
}

// Close implements io.Closer for the ioCloser type
func (r *ioCloser) Close() error {
	return nil // SSH session handles closing
}

// StdoutPipe returns a pipe that will be connected to the command's standard output
func (s *sshSessionExecutor) StdoutPipe() (io.ReadCloser, error) {
	r, err := s.Session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &ioCloser{r}, nil
}

// StderrPipe returns a pipe that will be connected to the command's standard error
func (s *sshSessionExecutor) StderrPipe() (io.ReadCloser, error) {
	r, err := s.Session.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &ioCloser{r}, nil
}

// withSudo wraps a command to be executed with sudo privileges using the provided password
func withSudo(cmd string, password string) string {
	return fmt.Sprintf("echo '%s' | sudo -S %s", password, cmd)
}

// calculateFileHash computes the SHA256 hash of a file, either locally or remotely.
// Returns the hash as a byte slice and any error encountered.
func calculateFileHash(filepath string, client *ssh.Client) ([]byte, error) {
	cmd := "sha256sum " + filepath + " | cut -d' ' -f1"

	// Local file
	if client == nil {
		output, err := exec.Command("sh", "-c", cmd).Output()
		if err != nil {
			return nil, fmt.Errorf("failed to get file hash: %w", err)
		}
		return output, nil
	}

	// Remote file
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()
	session.Stdout = nil // Disable stdout to capture output
	return session.Output(cmd)
}

// isProgressLine determines if a command output line represents a progress update.
// This function identifies lines that contain progress information typically generated
// by tools like rsync, wget, curl, and other transfer/download utilities.
//
// Progress lines are characterized by:
//   - Percentage indicators (%)
//   - Transfer speed indicators (MB/s, KB/s, etc.)
//   - Transfer completion indicators (xfer#, to-check, etc.)
//
// Examples of progress lines:
//   - "5046272   5%    4.75MB/s   00:00:19"
//   - "100204544 100%    1.64MB/s   00:00:00 (xfer#1, to-check=0/1)"
//   - "Downloading: 45% [2.3MB/s] [00:15<00:12]"
//
// This detection allows for structured logging of progress updates while
// maintaining regular logging for other command output.
func isProgressLine(line string) bool {
	// Must contain a percentage indicator
	if !strings.Contains(line, "%") {
		return false
	}

	// Must contain at least one of the common progress indicators
	progressIndicators := []string{
		"MB/s", "KB/s", "GB/s", // Transfer speeds
		"xfer#", "to-check", // rsync completion indicators
		"ETA", "ETA:", // Estimated time remaining
		"Downloading:", // Download progress
		"Uploading:",   // Upload progress
		"Progress:",    // Generic progress indicator
	}

	for _, indicator := range progressIndicators {
		if strings.Contains(line, indicator) {
			return true
		}
	}

	return false
}
