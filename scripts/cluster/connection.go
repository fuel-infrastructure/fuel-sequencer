package cluster

import (
	"fmt"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

type destination struct {
	host string
	user string
	auth ssh.AuthMethod
	dir  string
}

type connection struct {
	destination
	SSH *ssh.Client
}

func establishConnections() ([]connection, error) {
	var connections []connection
	var mu sync.Mutex
	var wg sync.WaitGroup
	errChan := make(chan error, len(destinations))

	for _, dest := range destinations {
		wg.Add(1)
		go func(dest destination) {
			defer wg.Done()

			client, err := connectSSH(dest)
			if err != nil {
				errChan <- err
				return
			}

			mu.Lock()
			connections = append(connections, connection{
				destination: dest,
				SSH:         client,
			})
			mu.Unlock()
		}(dest)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	return connections, nil
}

func closeConnections(connections []connection) {
	for _, conn := range connections {
		if conn.SSH != nil {
			conn.SSH.Close()
		}
	}
}

// connectSSH establishes an SSH connection to the specified destination
func connectSSH(dest destination) (*ssh.Client, error) {
	l := logging.Named("Connection")

	l.Debugw("establishing SSH connection", "host", dest.host, "user", dest.user)

	config := &ssh.ClientConfig{
		User:            dest.user,
		Auth:            []ssh.AuthMethod{dest.auth},
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
