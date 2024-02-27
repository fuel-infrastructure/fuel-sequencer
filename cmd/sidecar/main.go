package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/mockbridgex"
	sidecarserver "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service"
)

var (
	host               = flag.String("host", "localhost", "host for the grpc-service to listen on")
	port               = flag.String("port", "8080", "port for the grpc-service to listen on")
	ethNodeAPI         = flag.String("eth_node_api", "http://127.0.0.1:8545/", "Ethereum node API endpoint")
	contractAddressHex = flag.String("contract_address", "", "Contract address in hex format")
	ethStartBlockStr   = flag.String("eth_start_block", "0", "Ethereum start query block")
)

// start the sidecar-grpc server + sidecar process, cancel on interrupt or terminate.
func main() {
	// channel with width for either signal
	sigs := make(chan os.Signal, 1)

	// gracefully trigger close on interrupt or terminate signals
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// parse flags
	flag.Parse()

	// Validate required flags
	if *ethNodeAPI == "" || *contractAddressHex == "" {
		log.Fatal("ethNodeAPI and contractAddress are required flags")
	}

	// Convert the ethStartBlock to big.Int
	ethStartBlock := new(big.Int)
	_, ok := ethStartBlock.SetString(*ethStartBlockStr, 10)
	if !ok {
		log.Fatalf("Invalid ethStartBlock value: %s", *ethStartBlockStr)
	}

	// Connect to the ethereum client
	client, err := ethclient.Dial(*ethNodeAPI)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the address from string
	contractAddr := common.HexToAddress(*contractAddressHex)
	contractAbi, err := abi.JSON(strings.NewReader(mockbridgex.MockBridgeXABI))
	if err != nil {
		log.Fatal(err)
	}

	var logger *zap.Logger
	logger, err = zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %s\n", err.Error())
		return
	}

	// Create the sidecar.
	sideCar := sidecar.NewSidecar(
		client,
		contractAddr,
		contractAbi,
		ethStartBlock,
		logger,
	)
	if err != nil {
		logger.Error("failed to create sidecar", zap.Error(err))
		return
	}

	// create server
	srv := sidecarserver.NewSidecarServer(sideCar, logger)

	// cancel sidecar on interrupt or terminate
	go func() {
		<-sigs
		logger.Info(
			"received interrupt or terminate signal, closing sidecar",
		)

		cancel()
	}()

	// start sidecar + server, and wait for either to finish
	if err := srv.StartServer(ctx, *host, *port); err != nil {
		logger.Error("stopping server", zap.Error(err))
	}
}
