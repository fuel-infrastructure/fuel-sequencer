package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	cometutils "github.com/fuel-infrastructure/fuel-sequencer/sidecar/cometutils"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	scethclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/ethwrappedclient"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/mockbridgex"
	scsequencerclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/sequencerclient"
	sidecarserver "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service"
	scstore "github.com/fuel-infrastructure/fuel-sequencer/sidecar/store"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/sidecar"
)

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	_ codectypes.InterfaceRegistry,
	_ codec.Codec,
	basicManager module.BasicManager,
) {
	rootCmd.AddCommand(
		genutilcli.InitCmd(basicManager, app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addStartFlags)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		genesisCommand(txConfig, basicManager),
		queryCommand(),
		txCommand(),
		startSidecarServerCmd(),
		querySidecarServerCmd(),
		keys.Commands(),
	)
}

func addStartFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	sidecarconfig.AddStartCmdFlags(startCmd)
}

// genesisCommand builds genesis-related `fuelsequencerd genesis` command. Users may provide application specific commands as a parameter
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome)

	for _, subCmd := range cmds {
		cmd.AddCommand(subCmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.QueryEventForTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)
	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func startSidecarServerCmd() *cobra.Command {
	var (
		host                string
		port                string
		ethNodeRPC          string
		cosmosNodeRPC       string
		tendermintNodeRPC   string
		contractAddressHex  string
		unsafeEthereumBlock int64
		ethMaxBlockRange    int64
		development         bool
		acceptableDelay     uint64
	)

	cmd := &cobra.Command{
		Use:   "start-sidecar",
		Short: "Starts the Sidecar service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startSidecar(
				host,
				port,
				ethNodeRPC,
				cosmosNodeRPC,
				tendermintNodeRPC,
				contractAddressHex,
				unsafeEthereumBlock,
				ethMaxBlockRange,
				development,
				acceptableDelay,
			)
		},
	}

	cmd.Flags().StringVar(&host, "host", "localhost", "host for the grpc-service to listen on")
	cmd.Flags().StringVar(&port, "port", "8080", "port for the grpc-service to listen on")
	cmd.Flags().StringVar(&ethNodeRPC, "eth_node_rpc", "http://127.0.0.1:8545/", "Ethereum node RPC endpoint")
	cmd.Flags().StringVar(&cosmosNodeRPC, "cosmos_node_rpc", "127.0.0.1:9090", "Cosmos node RPC endpoint")
	cmd.Flags().StringVar(&tendermintNodeRPC, "tendermint_node_rpc", "http://127.0.0.1:26657", "Tendermint node RPC endpoint")
	cmd.Flags().StringVar(&contractAddressHex, "contract_address", "", "Contract address in hex format")
	cmd.Flags().Int64Var(&unsafeEthereumBlock, "unsafe_ethereum_block", 0, "Ethereum start query block")
	cmd.Flags().Int64Var(&ethMaxBlockRange, "eth_max_block_range", 100, "max number of Ethereum blocks per query")
	cmd.Flags().BoolVar(&development, "development", false, "Starts the sidecar in development mode")
	cmd.Flags().Uint64Var(&acceptableDelay, "unsafe_acceptable_delay", 1, "the amount of blocks the sidecar can be out-of-sync with Ethereum")

	return cmd
}

func startSidecar(
	host,
	port,
	ethNodeRPC,
	cosmosNodeRPC,
	tendermintNodeRPC,
	contractAddressHex string,
	unsafeEthereumBlock,
	ethMaxBlockRange int64,
	development bool,
	acceptableDelay uint64,
) error {
	sigs := make(chan os.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var logger *zap.Logger
	var err error
	if development {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return fmt.Errorf("failed to create logger: %s", err)
	}

	if unsafeEthereumBlock < 0 {
		return fmt.Errorf("ethereum start block must be >= 0, got: %d", unsafeEthereumBlock)
	}
	if ethMaxBlockRange < 1 {
		return fmt.Errorf("ethereum max block range must be >= 1, got: %d", ethMaxBlockRange)
	}
	if acceptableDelay > 10 {
		return fmt.Errorf("acceptable delay is too large, must be <= 10, got: %d", acceptableDelay)
	}

	// Check if the unsafeEthereumBlock is provided and use it instead of querying the genesis.
	startBlock := big.NewInt(0)
	if unsafeEthereumBlock > 0 {
		startBlock = big.NewInt(unsafeEthereumBlock)
		logger.Info(
			"ethereum start block set to unsafe-ethereum-block",
			zap.String("startBlock", startBlock.String()),
		)
	}

	// Create a connection to the Cosmos gRPC server.
	grpcConn, err := grpc.Dial(cosmosNodeRPC, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	// Create the sequencer client
	scsequencerclient := scsequencerclient.NewClient(grpcConn)

	// If the unsafeEthereumBlock is not set then we attempt to query the 
	// last Ethereum block synced from the genesis file and the Sequencer.
	if unsafeEthereumBlock == 0 {

		// Only query the genesis file if the tendermintNodeRPC was specified.
		if tendermintNodeRPC != "" {
			lastEthereumBlockSyncedUint, err := cometutils.QueryCometGenesisForLastEthereumBlockSynced(ctx, tendermintNodeRPC)
			if err != nil {
				logger.Error("Failed to read the response body of the genesis file",
					zap.String("url", tendermintNodeRPC),
					zap.Error(err))
			} else {
				startBlock = big.NewInt(int64(lastEthereumBlockSyncedUint + 1))
				logger.Info("Last ethereum block synced used from genesis",
					zap.String("startBlock", startBlock.String()))
			}
		}

		lastSyncedSequencerEthereumBlock, err := scsequencerclient.FetchLastSyncedEthereumBlock(ctx)
		if err != nil {
			logger.Error("Failed to query the last synced block from the sequencer",
				zap.Error(err))
		} else {
			nextBlockToStartQueryAt := big.NewInt(lastSyncedSequencerEthereumBlock.Int64() + 1)
			if nextBlockToStartQueryAt.Cmp(startBlock) > 0 {
				startBlock = nextBlockToStartQueryAt
				logger.Info("start block overwritten by last synced ethereum block query now using",
					zap.String("startBlock", startBlock.String()))
			}
		}
	}

	// If the startblock is 0 that means we've failed to set it through genesis or querying the sequencer therefore we don't start.
	if startBlock.Cmp(big.NewInt(0)) == 0 {
		panic("start block cannot be 0; ensure you have a connection to the Sequencer")
	}

	ethClient, err := ethclient.Dial(ethNodeRPC)
	if err != nil {
		return err
	}

	// TODO: replace MockBridgeXABI with actual contract once it's available.
	contractAddr := common.HexToAddress(contractAddressHex)
	contractAbi, err := abi.JSON(strings.NewReader(mockbridgex.MockBridgeXABI))
	if err != nil {
		return err
	}

	// Create the sidecar ethereum client
	scEthClient := scethclient.NewClient(ethClient, contractAddr, contractAbi)

	// Create the store
	eventStore := scstore.NewEventStore(startBlock, nil, big.NewInt(ethMaxBlockRange))

	sideCar := sidecar.NewSidecar(
		logger,
		scEthClient,
		scsequencerclient,
		eventStore,
		development,
		acceptableDelay,
	)
	srv := sidecarserver.NewSidecarServer(sideCar, logger)

	go func() {
		<-sigs
		logger.Info("Received interrupt or terminate signal, closing sidecar")
		cancel()
	}()

	if err := srv.InitializeServer(host, port); err != nil {
		logger.Error("Failed to initialize the server", zap.Error(err))
	}

	if err := srv.StartServer(ctx); err != nil {
		logger.Error("Stopping server", zap.Error(err))
	}

	return nil
}

func querySidecarServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query-sidecar-server-events [block-number]",
		Short:   "Queries block events from the Sidecar service by block number",
		Args:    cobra.ExactArgs(1),
		RunE:    queryBlockEvents,
		Aliases: []string{"qse"},
	}

	cmd.Flags().String("host", "localhost", "host of the gRPC service to query")
	cmd.Flags().String("port", "8080", "port of the gRPC service to query")

	return cmd
}

func queryBlockEvents(cmd *cobra.Command, args []string) error {
	host, err := cmd.Flags().GetString("host")
	if err != nil {
		return err
	}
	port, err := cmd.Flags().GetString("port")
	if err != nil {
		return err
	}

	blockNumber := args[0]
	_, err = strconv.Atoi(blockNumber) // try parse
	if err != nil {
		return fmt.Errorf("could not parse block number: %s", err.Error())
	}

	url := fmt.Sprintf("%s:%s", host, port)
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to connect to Sidecar service: %v", err)
	}
	defer conn.Close()

	client := types.NewSidecarClient(conn)

	resp, err := client.GetBlockEvents(context.Background(), &types.QueryBlockEventsRequest{BlockNumber: blockNumber})
	if err != nil {
		return fmt.Errorf("could not get block events: %v", err)
	}

	events := resp.GetEvents()

	if len(events) == 0 {
		fmt.Printf("No events emitted for block: %s\n", blockNumber)
		return nil
	}

	for _, event := range events {
		fmt.Printf("Block Event: %s\n", event)
	}

	return nil
}

// newApp creates the application
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	fuelSequencerApp, err := app.NewFuelSequencerApp(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
	if err != nil {
		panic(err)
	}
	return fuelSequencerApp
}

// appExport creates a new app (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var (
		bApp *app.FuelSequencerApp
		err  error
	)

	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	// overwrite the FlagSidecarEnabled since we don't need it for app export
	viperAppOpts.Set(sidecarconfig.FlagSidecarEnabled, false)
	appOpts = viperAppOpts

	if height != -1 {
		bApp, err = app.NewFuelSequencerApp(logger, db, traceStore, false, appOpts)
		if err != nil {
			return servertypes.ExportedApp{}, err
		}

		if err := bApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		bApp, err = app.NewFuelSequencerApp(logger, db, traceStore, true, appOpts)
		if err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return bApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
