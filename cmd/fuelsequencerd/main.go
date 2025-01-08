package main

import (
	"fmt"
	"os"
	"sync"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/grafana/pyroscope-go"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/fuel-infrastructure/fuel-sequencer/cmd/fuelsequencerd/cmd"
)

func main() {
	fmt.Println("Starting main application...")

	rootCmd := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	// Start profiler in a separate goroutine
	go func() {
		defer wg.Done()
		profiler, err := pyroscope.Start(pyroscope.Config{
			ApplicationName: "seqprofiler",
			ServerAddress:   "http://localhost:4040",
			Logger:          pyroscope.StandardLogger,
			ProfileTypes: []pyroscope.ProfileType{
				pyroscope.ProfileInuseSpace,   // Current memory in use
				pyroscope.ProfileAllocSpace,   // Cumulative memory allocations
				pyroscope.ProfileInuseObjects, // Current objects in use
				pyroscope.ProfileAllocObjects, // Cumulative object allocations
				pyroscope.ProfileGoroutines,   // Goroutine profiling
				pyroscope.ProfileMutexCount,   // Mutex contention count
			},
		})
		if err != nil {
			fmt.Printf("Profiler error: %v\n", err)
			return
		}
		fmt.Println("Profiler started successfully.")
		defer profiler.Stop()

		// Keep the profiler running
		select {}
	}()
}
