package execute

import (
	"os/exec"

	"go.uber.org/zap"
)

type LocalExec = func(l *zap.SugaredLogger, name string, args ...string) error

// Locally executes a command on the local machine with the given name and arguments.
// Returns an error if the command fails.
var Locally LocalExec = func(l *zap.SugaredLogger, name string, args ...string) error {
	// _, err := execute(l.Named("CMD"), exec.Command(name, args...), name, false)
	_, err := LocallyWithOutput(l, name, args...)
	return err
}

func LocallyWithCustomName(l *zap.SugaredLogger, name, cmdname string, args ...string) error {
	_, err := execute(l.Named("CMD"), exec.Command(name, args...), cmdname, false)
	return err
}

// LocallyWithOutput executes a command locally and returns its output.
// Returns the trimmed output and any error encountered.
func LocallyWithOutput(l *zap.SugaredLogger, name string, args ...string) (string, error) {
	return execute(l.Named("CMD"), exec.Command(name, args...), name, false)
}
