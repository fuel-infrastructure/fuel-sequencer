package main

import (
	"flag"
	"log"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/cluster/pkg/fullcluster"
)

func main() {
	var configPath = flag.String("config", "", "Path to cluster configuration file")
	flag.Parse()

	if *configPath == "" {
		log.Fatalf("no config path defined")
	}

	if err := fullcluster.SetupWithConfig(*configPath); err != nil {
		log.Fatalf("Setup failed: %v", err)
	}
}
