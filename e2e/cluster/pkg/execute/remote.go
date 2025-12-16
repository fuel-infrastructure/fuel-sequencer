package execute

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

// Remotely executes a command on a remote machine through an SSH connection.
// Returns an error if the command fails.
func Remotely(l *zap.SugaredLogger, client *ssh.Client, cmd string) error {
	cmdname := strings.Split(cmd, " ")[0]
	return RemotelyWithCustomName(l, client, cmd, cmdname)
}

func RemotelyWithCustomName(l *zap.SugaredLogger, client *ssh.Client, cmd, cmdname string) error {
	// Create new SSH client
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	_, err = execute(l.Named("SSH"), &sshSessionExecutor{Session: session, cmd: cmd}, cmdname, false)
	return err
}

func RemotelyWithOutput(l *zap.SugaredLogger, client *ssh.Client, cmd, cmdname string) (string, error) {
	// Create new SSH client
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	return execute(l.Named("SSH"), &sshSessionExecutor{Session: session, cmd: cmd}, cmdname, false)
}
