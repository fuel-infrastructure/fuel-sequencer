package execute

import (
	"io"

	"golang.org/x/crypto/ssh"
)

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
