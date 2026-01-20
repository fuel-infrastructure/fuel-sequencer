package execute

import (
	"os/exec"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

// OnSystem executes a command on a system, automatically routing to local or remote execution
// based on whether the system has an SSH connection. If sys.SSH is nil, executes locally.
// Otherwise, executes remotely via SSH.
func OnSystem(l *zap.SugaredLogger, sys setup.System, cmd string) error {
	if sys.IsLocal {
		// Execute locally using sh -c to handle shell features properly
		return LocallyWithCustomName(l, "sh", "sh", "-c", cmd)
	}
	// Execute remotely via SSH
	return Remotely(l, sys.SSH, cmd)
}

// OnSystemWithOutput executes a command on a system and returns its output.
// Automatically routes to local or remote execution based on SSH presence.
func OnSystemWithOutput(l *zap.SugaredLogger, sys setup.System, cmd, cmdname string) (string, error) {
	if sys.IsLocal {
		// Execute locally using sh -c to handle shell features properly
		output, err := execute(l.Named("CMD"), exec.Command("sh", "-c", cmd), cmdname, false)
		return output, err
	}
	// Execute remotely via SSH
	return RemotelyWithOutput(l, sys.SSH, cmd, cmdname)
}

// OnSystemWithSudo executes a command with sudo privileges on a system.
// Automatically routes to local or remote execution based on SSH presence.
// sudo command is only required on remote systems.
func OnSystemWithSudo(l *zap.SugaredLogger, sys setup.System, cmd string) error {
	if !sys.IsLocal {
		cmd = WithSudo(cmd, sys.Destination.Pass)
	}
	return OnSystem(l, sys, cmd)
}
