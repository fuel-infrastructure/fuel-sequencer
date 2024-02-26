package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/mockbridgex"
	sidecarserver "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/servers/sidecar"
)

var (
	host = flag.String("host", "localhost", "host for the grpc-service to listen on")
	port = flag.String("port", "8080", "port for the grpc-service to listen on")
)

// start the oracle-grpc server + oracle process, cancel on interrupt or terminate.
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

	ethNodeAPI, contractAddressHex, ethStartBlock := config.GetEnvironmentalVariables()

	client, err := ethclient.Dial(ethNodeAPI)
	if err != nil {
		log.Fatal(err)
	}

	contractAddr := common.HexToAddress(contractAddressHex)
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

	// Create the oracle.
	sideCar := sidecar.NewSidecar(
		client,
		contractAddr,
		contractAbi,
		ethStartBlock,
		logger,
	)
	if err != nil {
		logger.Error("failed to create oracle", zap.Error(err))
		return
	}

	// create server
	srv := sidecarserver.NewSidecarServer(sideCar, logger)

	// cancel oracle on interrupt or terminate
	go func() {
		<-sigs
		logger.Info(
			"received interrupt or terminate signal, closing oracle",
		)

		cancel()
	}()

	// // start prometheus metrics
	// if cfg.Metrics.Enabled {
	// 	logger.Info("starting prometheus metrics", zap.String("address", cfg.Metrics.PrometheusServerAddress))
	// 	ps, err := promserver.NewPrometheusServer(cfg.Metrics.PrometheusServerAddress, logger)
	// 	if err != nil {
	// 		logger.Error("failed to start prometheus metrics", zap.Error(err))
	// 		return
	// 	}

	// 	go ps.Start()

	// 	// close server on shut-down
	// 	go func() {
	// 		<-ctx.Done()
	// 		logger.Info("stopping prometheus metrics")
	// 		ps.Close()
	// 	}()
	// }

	// start oracle + server, and wait for either to finish
	if err := srv.StartServer(ctx, *host, *port); err != nil {
		logger.Error("stopping server", zap.Error(err))
	}
}
