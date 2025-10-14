package connect

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/setup"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// EstablishConnections creates SSH connections to all systems in parallel.
// Returns any error encountered.
func EstablishConnections(logging *zap.SugaredLogger, systems []setup.System) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(systems))

	for i := range systems {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			client, err := ConnectSSH(logging, systems[i].Destination)
			if err != nil {
				errChan <- err
				return
			}
			systems[i].SSH = client
		}(i)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

// CloseConnections closes all SSH connections in the provided slice.
func CloseConnections(systems []setup.System) {
	for _, sys := range systems {
		if sys.SSH != nil {
			sys.SSH.Close()
		}
	}
}

// ConnectSSH establishes an SSH connection to a single destination.
// Returns the SSH client and any error encountered.
func ConnectSSH(logging *zap.SugaredLogger, dest setup.Destination) (*ssh.Client, error) {
	l := logging.Named("Connection")

	l.Debugw("establishing SSH connection", "host", dest.Host, "user", dest.User)

	config := &ssh.ClientConfig{
		User:            dest.User,
		Auth:            []ssh.AuthMethod{ssh.Password(dest.Pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use ssh.FixedHostKey() or ssh.KnownHosts()
		Timeout:         30 * time.Second,
	}

	// Establish connection
	client, err := ssh.Dial("tcp", net.JoinHostPort(dest.Host, "22"), config)
	if err != nil {
		l.Errorw("ssh connection failed", "host", dest.Host, "user", dest.User, "error", err)
		return nil, fmt.Errorf("failed to connect to %s: %w", dest.Host, err)
	}

	l.Debugw("successfully established ssh connection", "host", dest.Host, "user", dest.User)
	return client, nil
}
