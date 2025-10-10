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

// system represents the state of a remote host with configs and connection
type system struct {
	destination
	options
	SSH *ssh.Client
}

// destination represents a remote host configuration for SSH connections
type destination struct {
	peer_ip string // IP address used for P2P communication
	host    string // Hostname or IP for SSH connection
	user    string // SSH username
	pass    string // SSH password
	dir     string // Working directory on remote host
}

// options represents the options for the system that need to be configured
type options struct {
	sequencer bool
	blobpool  bool
}

// establishConnections creates SSH connections to all systems in parallel.
// Returns any error encountered.
func establishConnections() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(systems))

	for i := range systems {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			client, err := connectSSH(systems[i].destination)
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
	for _, sys := range systems {
		if sys.SSH != nil {
			sys.SSH.Close()
		}
	}
}

// connectSSH establishes an SSH connection to a single destination.
// Returns the SSH client and any error encountered.
func connectSSH(dest destination) (*ssh.Client, error) {
	l := logging.Named("Connection")

	l.Debugw("establishing SSH connection", "host", dest.host, "user", dest.user)

	config := &ssh.ClientConfig{
		User:            dest.user,
		Auth:            []ssh.AuthMethod{ssh.Password(dest.pass)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // Note: In production, use ssh.FixedHostKey() or ssh.KnownHosts()
		Timeout:         30 * time.Second,
	}

	// Establish connection
	client, err := ssh.Dial("tcp", net.JoinHostPort(dest.host, "22"), config)
	if err != nil {
		l.Errorw("ssh connection failed", "host", dest.host, "user", dest.user, "error", err)
		return nil, fmt.Errorf("failed to connect to %s: %w", dest.host, err)
	}

	l.Debugw("successfully established ssh connection", "host", dest.host, "user", dest.user)
	return client, nil
}
