package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/mockbridgex"
	sidecarserver "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	host                = flag.String("host", "localhost", "host for the grpc-service to listen on")
	port                = flag.String("port", "8080", "port for the grpc-service to listen on")
	ethNodeRPC          = flag.String("eth_node_rpc", "http://127.0.0.1:8545/", "Ethereum node RPC endpoint")
	cosmosNodeRPC       = flag.String("cosmos_node_rpc", "127.0.0.1:9090", "Cosmos node RPC endpoint")
	tendermintNodeRPC   = flag.String("tendermint_node_rpc", "http://127.0.0.1:26657", "Tendermint node RPC endpoint")
	contractAddressHex  = flag.String("contract_address", "", "Contract address in hex format")
	unsafeEthereumBlock = flag.Int64("unsafe_ethereum_block", 0, "Ethereum start query block")
	ethMaxBlockRange    = flag.Int64("eth_max_block_range", 100, "max number of Ethereum blocks per query")
	development         = flag.Bool("development", false, "Start logger in development mode")
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

	// Validate required flags that have no default
	if *contractAddressHex == "" {
		log.Fatal("contract_address is a required flag")
	}

	// Validate flag values
	if *unsafeEthereumBlock < 0 {
		log.Fatalf("ethereum start block must be >= 0, got: %d", *unsafeEthereumBlock)
	}
	if *ethMaxBlockRange < 1 {
		log.Fatalf("ethereum max block range must be >= 1, got: %d", *ethMaxBlockRange)
	}

	// Check if the unsafeEthereumBlock is provided and use it instead of querying the genesis.
	var lastEthereumBlockSynced *big.Int
	if *unsafeEthereumBlock > 0 {
		lastEthereumBlockSynced = big.NewInt(*unsafeEthereumBlock)
	} else {
		// Otherwise query the genesis file for the up to date last ethereum block synced.

		// Check if the tendermintNodeRPC is not empty
		if *tendermintNodeRPC == "" {
			log.Fatal("tendermint node rpc url is required")
		}

		// Append `/genesis?` to the tendermintNodeRPC URL
		genesisURL := fmt.Sprintf("%s/genesis?", *tendermintNodeRPC)

		// Query the genesis file
		resp, err := http.Get(genesisURL)
		if err != nil {
			log.Fatalf("Failed to make request to the tendermint node rpc %s: %v", genesisURL, err)
		}
		defer resp.Body.Close()

		// Read the response body
		genbz, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read the response body of the genesis file from tendermint node rpc %s: %v", genesisURL, err)
		}

		// Get the lasts ethereum block synced from genesis or stop.
		lastEthereumBlockSyncedUint := sidecar.MustGetLastEthereumBlockSyncedFromGenesis(genbz)
		lastEthereumBlockSynced = big.NewInt(int64(lastEthereumBlockSyncedUint))
	}

	// Connect to the ethereum client
	ethClient, err := ethclient.Dial(*ethNodeRPC)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the address from string
	contractAddr := common.HexToAddress(*contractAddressHex)
	contractAbi, err := abi.JSON(strings.NewReader(mockbridgex.MockBridgeXABI))
	if err != nil {
		log.Fatal(err)
	}

	// Create a connection to the Cosmos gRPC server.
	grpcConn, err := grpc.Dial(
		*cosmosNodeRPC,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	)
	if err != nil {
		log.Fatal(err)
	}

	// This creates a gRPC client to query the x/bridge service.
	bridgeClient := bridgetypes.NewQueryClient(grpcConn)

	var logger *zap.Logger
	if *development {
		logger, err = zap.NewDevelopment()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create logger: %s\n", err.Error())
			return
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to create logger: %s\n", err.Error())
			return
		}
	}

	// Create the sidecar.
	sideCar := sidecar.NewSidecar(
		ethClient,
		bridgeClient,
		contractAddr,
		contractAbi,
		lastEthereumBlockSynced,
		big.NewInt(*ethMaxBlockRange),
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
