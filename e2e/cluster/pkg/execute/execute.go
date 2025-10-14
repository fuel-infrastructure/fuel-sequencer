package execute

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"sync"

	"go.uber.org/zap"
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
