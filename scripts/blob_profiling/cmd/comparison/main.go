package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/visualizer"
)

func main() {
	var (
		decoupledDir = flag.String("decoupled", "", "Path to MsgPostBlobMetadata system output directory")
		fullpostDir  = flag.String("fullpost", "", "Path to MsgPostBlob system output directory")
	)
	flag.Parse()

	if *decoupledDir == "" || *fullpostDir == "" {
		fmt.Println("Usage: go run comparison.go -decoupled <path> -fullpost <path>")
		fmt.Println("Example: go run comparison.go -decoupled scripts/blob_profiling/benchnet_eu_decoupled_output -fullpost scripts/blob_profiling/benchnet_eu_fullpost_output")
		os.Exit(1)
	}

	// Check if directories exist
	if _, err := os.Stat(*decoupledDir); os.IsNotExist(err) {
		log.Fatalf("Decoupled directory does not exist: %s", *decoupledDir)
	}
	if _, err := os.Stat(*fullpostDir); os.IsNotExist(err) {
		log.Fatalf("Fullpost directory does not exist: %s", *fullpostDir)
	}

	fmt.Printf("Generating comparison plot...\n")
	fmt.Printf("Decoupled system: %s\n", *decoupledDir)
	fmt.Printf("Fullpost system: %s\n", *fullpostDir)

	if err := visualizer.GenerateComparisonPlot(*decoupledDir, *fullpostDir); err != nil {
		log.Fatalf("Failed to generate comparison plot: %v", err)
	}

	fmt.Println("Comparison plot generated successfully!")
}
