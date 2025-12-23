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

// execute runs a command through the provided executor, capturing output and optionally logging it.
// It streams stdout and stderr through the logger (unless silent) and returns the combined stdout output with any error encountered.
func execute(l *zap.SugaredLogger, e executor, commandName string, silent bool) (string, error) {

	// Capture output for logging by creating pipes for stdout and stderr
	stdout, err := e.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	stderr, err := e.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create a wait group to ensure both goroutines complete
	var wg sync.WaitGroup
	wg.Add(2)

	// Capture stdout for return
	var stdoutBuf strings.Builder

	// Stream stdout with smart progress detection
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				stdoutBuf.WriteString(line)
				stdoutBuf.WriteByte('\n')
				if !silent {
					if isProgressLine(line) {
						// This is a progress update, log it as progress info
						l.Infow(fmt.Sprintf("running %s...", commandName), "progress", trimmed)
					} else {
						// Regular output line (skip empty lines)
						l.Info(line)
					}
				}
			}
		}
	}()

	// Stream stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if trimmed := strings.TrimSpace(line); trimmed != "" && !silent {
				l.Warn(line)
			}
		}
	}()

	// Start the command
	if err := e.Start(); err != nil {
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Wait for both output streams to complete
	wg.Wait()

	// Wait for the command to complete
	if err := e.Wait(); err != nil {
		if !silent {
			l.Errorw("command failed", "error", err)
		}
		return "", fmt.Errorf("command failed: %w", err)
	}

	return strings.TrimSpace(stdoutBuf.String()), nil
}
