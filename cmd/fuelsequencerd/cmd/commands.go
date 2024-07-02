package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"os"
	"strconv"
	"time"

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
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/fuel-infrastructure/fuel-sequencer/app"
	cometutils "github.com/fuel-infrastructure/fuel-sequencer/sidecar/cometutils"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	scethclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/ethwrappedclient"
	scsequencerclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/sequencerclient"
	sidecarserver "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service"
	scstore "github.com/fuel-infrastructure/fuel-sequencer/sidecar/store"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
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

	scrCfg := sidecarConfig{}
	seqCfg := sequencerConfig{}
	ethCfg := ethereumConfig{}

	cmd := &cobra.Command{
		Use:   "start-sidecar",
		Short: "Starts the Sidecar service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return startSidecar(scrCfg, seqCfg, ethCfg)
		},
	}

	// Sidecar
	cmd.Flags().StringVar(&scrCfg.host, FlagSidecarHost, "localhost", "host for the gRPC server to listen on")
	cmd.Flags().StringVar(&scrCfg.port, FlagSidecarPort, "8080", "port for the gRPC server to listen on")
	cmd.Flags().BoolVar(&scrCfg.development, FlagSidecarDevelopment, false, "starts the sidecar in development mode")

	// Ethereum
	cmd.Flags().StringVar(&ethCfg.webSocketUrl, FlagEthereumWebSocketUrl, "ws://127.0.0.1:8545", "the ethereum node WebSocket endpoint")
	cmd.Flags().StringVar(&ethCfg.contractAddrHex, FlagEthereumContractAddr, "", "address in hex format of the contract to monitor for logs")
	cmd.Flags().Int64Var(&ethCfg.maxBlockRange, FlagEthereumMaxBlockRange, 100, "max number of ethereum blocks queried at one go")
	cmd.Flags().DurationVar(&ethCfg.minLogsQueryInterval, FlagEthereumMinLogsQueryInterval, time.Second*5, "minimum wait between successive queries for logs")
	cmd.Flags().Int64Var(&ethCfg.unsafeStartBlock, FlagEthereumUnsafeStartBlock, 0, "the ethereum block to start querying from")
	cmd.Flags().Int64Var(&ethCfg.unsafeEndBlock, FlagEthereumUnsafeEndBlock, 0, "the last ethereum block to sync - incorrect use can cause the validator to propose empty blocks, leading to slashing!")

	// Sequencer
	cmd.Flags().StringVar(&seqCfg.grpcUrl, FlagSequencerGrpcUrl, "127.0.0.1:9090", "the sequencer's gRPC endpoint")
	cmd.Flags().StringVar(&seqCfg.rpcUrl, FlagSequencerRpcUrl, "http://127.0.0.1:26657", "the sequencer's CometBFT RPC endpoint")

	return cmd
}

func startSidecar(
	scrCfg sidecarConfig,
	seqCfg sequencerConfig,
	ethCfg ethereumConfig,
) error {
	sigs := make(chan os.Signal, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Configure logger.
	var loggerCfg zap.Config
	if scrCfg.development {
		loggerCfg = zap.NewDevelopmentConfig()
	} else {
		loggerCfg = zap.NewProductionConfig()
		loggerCfg.EncoderConfig.CallerKey = zapcore.OmitKey // do not output file and line number of caller
	}
	loggerCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	loggerCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	loggerCfg.Encoding = "console" // more readable compared to JSON

	// Build logger based on config.
	logger, err := loggerCfg.Build()
	if err != nil {
		return fmt.Errorf("failed to create logger: %s", err)
	}

	if ethCfg.unsafeStartBlock < 0 {
		return fmt.Errorf("ethereum unsafe start block must be >= 0, got: %d", ethCfg.unsafeStartBlock)
	}
	if ethCfg.unsafeEndBlock < 0 {
		return fmt.Errorf("ethereum unsafe end block must be >= 0, got: %d", ethCfg.unsafeEndBlock)
	}
	if ethCfg.maxBlockRange < 1 {
		return fmt.Errorf("ethereum max block range must be >= 1, got: %d", ethCfg.maxBlockRange)
	}

	// Check if the unsafe start block is provided and use it instead of querying the genesis.
	startBlock := big.NewInt(0)
	if ethCfg.unsafeStartBlock > 0 {
		startBlock = big.NewInt(ethCfg.unsafeStartBlock)
		logger.Warn(
			fmt.Sprintf("ethereum start block set to %s flag value", FlagEthereumUnsafeStartBlock),
			zap.String("start_block", startBlock.String()),
		)
	}

	// Check if the unsafe end block is provided and use it.
	var endBlock *big.Int
	if ethCfg.unsafeEndBlock > 0 {
		endBlock = big.NewInt(ethCfg.unsafeEndBlock)
		logger.Warn(
			fmt.Sprintf("ethereum end block set to %s flag value", FlagEthereumUnsafeEndBlock),
			zap.String("end_block", endBlock.String()),
		)
	} else {
		endBlock = nil
	}

	// Create a connection to the Cosmos gRPC server.
	logger.Info("dialling Sequencer node", zap.String("grpc_url", seqCfg.grpcUrl))
	grpcConn, err := grpc.Dial(seqCfg.grpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	// Create the sequencer client
	scSequencerClient := scsequencerclient.NewClient(grpcConn)

	// If the unsafe start block is not set then we attempt to query the
	// last Ethereum block synced from the genesis file and the Sequencer.
	if ethCfg.unsafeStartBlock == 0 {

		// Only try quering the genesis file if the sequencer RPC URL was specified.
		if seqCfg.rpcUrl != "" {
			lastEthereumBlockSynced, err := cometutils.QuerySequencerGenesisForLastEthereumBlockSynced(
				ctx, seqCfg.rpcUrl,
			)
			if err != nil {
				logger.Error(
					"failed to read the response body of the genesis file",
					zap.String("rpc_url", seqCfg.rpcUrl),
					zap.Error(err),
				)
			} else {
				startBlock = big.NewInt(int64(lastEthereumBlockSynced + 1))
				logger.Info(
					"ethereum start block set to LastEthereumBlockSynced+1 from Sequencer genesis",
					zap.String("start_block", startBlock.String()),
				)
			}
		}

		// Try querying the last Ethereum block synced from the Sequencer.
		lastEthereumBlockSynced, err := scSequencerClient.FetchLastEthereumBlockSynced(ctx)
		if err != nil {
			logger.Warn(
				"failed to query LastEthereumBlockSynced from Sequencer, but maybe Sequencer hasn't started",
				zap.String("grpc_url", seqCfg.grpcUrl),
				zap.Error(err),
			)
		} else {
			newStartBlock := new(big.Int).Add(lastEthereumBlockSynced, big.NewInt(1))
			if newStartBlock.Cmp(startBlock) > 0 {
				startBlock = newStartBlock
				logger.Info(
					"ethereum start block set to LastEthereumBlockSynced+1 from Sequencer state",
					zap.String("start_block", startBlock.String()),
				)
			} else {
				logger.Warn(
					"ignoring LastEthereumBlockSynced from Sequencer because it is too small",
					zap.String("last_ethereum_block_synced", lastEthereumBlockSynced.String()),
					zap.String("start_block", startBlock.String()),
				)
			}
		}
	}

	// If the startBlock is 0, we've failed to set it through the various attempts (unsafe flag / genesis / node).
	if startBlock.Cmp(big.NewInt(0)) == 0 {
		panic(fmt.Sprintf(
			"did not find a start block; ensure Sequencer is available at grpc=%s, rpc=%s",
			seqCfg.grpcUrl, seqCfg.rpcUrl,
		))
	}

	logger.Info("dialling Ethereum node", zap.String("ws_url", ethCfg.webSocketUrl))
	ethClient, err := ethclient.Dial(ethCfg.webSocketUrl)
	if err != nil {
		return err
	}

	// Contract ABI
	var contractAbi abi.ABI
	err = contractAbi.UnmarshalJSON([]byte(sidecartypes.MockSequencerProxyContractABI))
	if err != nil {
		return err
	}

	// Create the sidecar ethereum client
	contractAddr := common.HexToAddress(ethCfg.contractAddrHex)
	scEthClient := scethclient.NewClient(logger, ethClient, contractAddr, contractAbi, ethCfg.minLogsQueryInterval)

	// Create the store
	eventStore := scstore.NewEventStore(startBlock, endBlock, big.NewInt(ethCfg.maxBlockRange))

	sideCar := sidecar.NewSidecar(
		logger,
		scEthClient,
		scSequencerClient,
		eventStore,
		scrCfg.development,
	)
	srv := sidecarserver.NewSidecarServer(sideCar, logger)

	go func() {
		<-sigs
		logger.Info("received interrupt or terminate signal, closing sidecar")
		cancel()
	}()

	if err := srv.InitializeServer(scrCfg.host, scrCfg.port); err != nil {
		logger.Error("failed to initialize the server", zap.Error(err))
	}

	if err := srv.StartServer(ctx); err != nil {
		logger.Error("stopping server", zap.Error(err))
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

	cmd.Flags().StringP(FlagSidecarGrpcUrl, "s", "localhost:8080", "Sidecar's gRPC URL")
	cmd.Flags().DurationP(FlagQueryTimeout, "t", time.Second*5, "how long to wait before timing out")

	return cmd
}

func queryBlockEvents(cmd *cobra.Command, args []string) error {
	sidecarGrpcUrl, err := cmd.Flags().GetString(FlagSidecarGrpcUrl)
	if err != nil {
		return err
	}
	queryTimeout, err := cmd.Flags().GetDuration(FlagQueryTimeout)
	if err != nil {
		return err
	}

	blockNumber := args[0]
	_, err = strconv.Atoi(blockNumber) // try parse
	if err != nil {
		return fmt.Errorf("could not parse block number: %s", err.Error())
	}

	// Prepending a "//" allows addresses without a scheme.
	if _, err := url.Parse("//" + sidecarGrpcUrl); err != nil {
		return fmt.Errorf("invalid Sidecar address: %w", err)
	}

	conn, err := grpc.Dial(sidecarGrpcUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Sidecar service: %v", err)
	}
	defer conn.Close()

	sidecarClient := sidecartypes.NewSidecarClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	resp, err := sidecarClient.GetBlockEvents(ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: blockNumber})
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
