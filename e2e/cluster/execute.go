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

type executor interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
}

// execute can use exec.Cmd directly
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

func locally(l *zap.SugaredLogger, name string, args ...string) error {
	return execute(l.Named("CMD"), exec.Command(name, args...))
}

func remotely(l *zap.SugaredLogger, client *ssh.Client, cmd string) error {
	// Create new SSH client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return execute(l.Named("SSH"), &sshSessionExecutor{Session: session, cmd: cmd})
}

// Need to wrap ssh.Session now
type sshSessionExecutor struct {
	*ssh.Session
	cmd string
}

func (s *sshSessionExecutor) Start() error {
	return s.Session.Start(s.cmd)
}

// ioCloser adapts an io.Reader to io.ReadCloser
type ioCloser struct {
	io.Reader
}

func (r *ioCloser) Close() error {
	return nil // SSH session handles closing
}

func (s *sshSessionExecutor) StdoutPipe() (io.ReadCloser, error) {
	r, err := s.Session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	return &ioCloser{r}, nil
}

func (s *sshSessionExecutor) StderrPipe() (io.ReadCloser, error) {
	r, err := s.Session.StderrPipe()
	if err != nil {
		return nil, err
	}
	return &ioCloser{r}, nil
}
