package custom

import (
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

type Service interface {
	Name() string
	Setup(l *zap.SugaredLogger, sys setup.System) error
	Deploy(l *zap.SugaredLogger, sys setup.System) error
	Shutdown(l *zap.SugaredLogger, sys setup.System) error
}
