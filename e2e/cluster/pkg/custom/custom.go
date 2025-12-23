package custom

import (
	"fmt"
	"sync"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/connect"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/internal/sequencer"
	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/setup"
	"go.uber.org/zap"
)

var logging *zap.SugaredLogger

// init initializes the package-level logger
func init() {
	logger, _ := zap.NewDevelopment()
	logging = logger.Sugar().Named("Custom Setup")
	defer logger.Sync()
}

// logAndWrapErr logs an error message and wraps the original error with additional context
func logAndWrapErr(msg string, err error) error {
	logging.Errorw(msg, "error", err)
	return fmt.Errorf("%s: %w", msg, err)
}

// Run orchestrates an entire custom cluster setup process, to run a Service on the cluster.
// Returns an error if any step fails.
func Run(service Service, clusterConfigPath string) error {
	// Load configuration first
	if err := setup.LoadConfig(clusterConfigPath); err != nil {
		return logAndWrapErr("failed to load configuration", err)
	}

	systems, err := sequencer.CheckNetworkMnemonics(logging)
	if err != nil {
		return logAndWrapErr("loaded configuration has errors", err)
	}

	// Establish connection to all destinations
	if err := connect.EstablishConnections(logging, systems); err != nil {
		return logAndWrapErr("connection establishment failed", err)
	}
	defer connect.CloseConnections(systems)

	var operations = map[string]func(*zap.SugaredLogger, setup.System) error{
		"setup":    service.Setup,
		"deploy":   service.Deploy,
		"shutdown": service.Shutdown,
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, len(systems))
	defer close(errs)

	for opstring, operand := range operations {
		for _, sys := range systems {
			wg.Add(1)
			go func(sys setup.System) {
				defer wg.Done()
				if err := operand(logging, sys); err != nil {
					errs <- logAndWrapErr(service.Name()+" "+opstring+" failed", err)
				}
			}(sys)
		}

		wg.Wait()
		if err := <-errs; err != nil {
			return err
		}
	}

	return nil
}
