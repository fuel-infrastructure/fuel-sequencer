// Package runner provides functionality for setting up and managing a distributed
// network of Fuel Sequencer validator nodes. It handles binary building, configuration,
// deployment and management of the network.
package runner

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// establishConnections creates SSH connections to all systems in parallel.
// Returns any error encountered.
func establishConnections() error {
	var wg sync.WaitGroup
	systems := Systems()
	errChan := make(chan error, len(systems))

	for i := range systems {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			client, err := connectSSH(systems[i].Destination)
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

// closeConnections closes all SSH connections in the provided slice.
func closeConnections() {
	systems := Systems()
	for _, sys := range systems {
		if sys.SSH != nil {
			sys.SSH.Close()
		}
	}
}

// connectSSH establishes an SSH connection to a single destination.
// Returns the SSH client and any error encountered.
func connectSSH(dest Destination) (*ssh.Client, error) {
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
