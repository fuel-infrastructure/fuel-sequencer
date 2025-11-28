package execute

import (
	"os/exec"

	"go.uber.org/zap"
)

type LocalExec = func(l *zap.SugaredLogger, name string, args ...string) error

// Locally executes a command on the local machine with the given name and arguments.
// Returns an error if the command fails.
var Locally LocalExec = func(l *zap.SugaredLogger, name string, args ...string) error {
	return execute(l.Named("CMD"), exec.Command(name, args...), name)
}

func LocallyWithCustomName(l *zap.SugaredLogger, name, cmdname string, args ...string) error {
	return execute(l.Named("CMD"), exec.Command(name, args...), cmdname)
}
