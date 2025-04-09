// Package cluster provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package cluster

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
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
func execute(l *zap.SugaredLogger, e executor) error {

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

	// Stream stdout
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			l.Info(scanner.Text())
		}
	}()

	// Stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			l.Warn(scanner.Text())
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
	return execute(l.Named("CMD"), exec.Command(name, args...))
}

// remotely executes a command on a remote machine through an SSH connection.
// Returns an error if the command fails.
func remotely(l *zap.SugaredLogger, client *ssh.Client, cmd string) error {
	// Create new SSH client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return execute(l.Named("SSH"), &sshSessionExecutor{Session: session, cmd: cmd})
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
