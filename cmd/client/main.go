package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

var (
	host        = flag.String("host", "localhost", "host of the gRPC service to query")
	port        = flag.String("port", "8080", "port of the gRPC service to query")
	blockNumber = flag.String("blocknumber", "", "block number to query events for")
)

func main() {
	// Channel with width for termination signal.
	sigs := make(chan os.Signal, 1)

	// Gracefully trigger close on interrupt or terminate signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Parse flags.
	flag.Parse()

	// Check for block number input
	if *blockNumber == "" {
		log.Fatalf("You must provide a block number using the -blocknumber flag.")
	}

	// Set up a connection to the server.
	url := fmt.Sprintf("%s:%s", *host, *port)
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a new client
	client := sidecartypes.NewSidecarClient(conn)

	log.Printf("Calling GetBlockEvents RPC for block number %s...\n", *blockNumber)
	resp, err := client.GetBlockEvents(
		context.Background(),
		&sidecartypes.QueryBlockEventsRequest{BlockNumber: *blockNumber},
	)
	if err != nil {
		log.Fatalf("could not get block events: %v", err) //nolint
	}

	events := resp.GetEvents()

	// Log the response
	for _, event := range events {
		log.Printf("Block Event: %s\n", event)
	}
}
