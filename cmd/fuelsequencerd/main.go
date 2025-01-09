package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/fuel-infrastructure/fuel-sequencer/cmd/fuelsequencerd/cmd"
	"github.com/grafana/pyroscope-go"
	"github.com/spf13/viper"
)

// TO REMOVE IF CHAIN DIRECTORY NEEDS TO BE INITIALIZED
func initConfig() {
	viper.SetDefault("SERVICE_TYPE", "sequencer")
	viper.SetDefault("PYROSCOPE_SERVER", "http://localhost:4040")
	viper.AutomaticEnv()
}

func main() {
	// TO REMOVE IF CHAIN DIRECTORY NEEDS TO BE INITIALIZED
	initConfig()

	// TO REMOVE IF CHAIN DIRECTORY NEEDS TO BE INITIALIZED
	// Check for environment variables.
	serviceType := viper.GetString("SERVICE_TYPE")
	pyroscopeServer := viper.GetString("PYROSCOPE_SERVER")

	// TO REMOVE IF CHAIN DIRECTORY NEEDS TO BE INITIALIZED
	// Start profiler
	profiler, err := pyroscope.Start(pyroscope.Config{
		// based on service type to run profiling for both node and sidecar
		ApplicationName: fmt.Sprintf("profiler-%s", serviceType),
		ServerAddress:   pyroscopeServer,
		Logger:          pyroscope.StandardLogger,
	})
	if err != nil {
		fmt.Printf("Profiler error: %v\n", err)
		return
	}
	defer profiler.Stop()

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
