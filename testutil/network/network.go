package network

import (
	"encoding/json"
	"fmt"
	"testing"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/cosmos/cosmos-sdk/types/mempool"

	fuelsequencerapp "github.com/fuel-infrastructure/fuel-sequencer/app"
	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	sequencingtypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	"github.com/stretchr/testify/require"
)

type (
	Network = network.Network
	Config  = network.Config
)

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...Config) *Network {
	t.Helper()
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = DefaultConfig()

		// Enable logging.
		cfg.EnableLogging = true

		// Set custom number of validators.
		cfg.NumValidators = 5

		// Update x/bridge genesis to ensure chain runs.
		bridgeGenesis := bridgetypes.DefaultGenesis()
		bridgeGenesis.Params.BridgeDenom = "stake"

		// Update x/sequencing genesis to ensure chain runs.
		sequencingGenesis := sequencingtypes.DefaultGenesis()

		bz, err := json.Marshal(bridgeGenesis)
		require.NoError(t, err)
		cfg.GenesisState["bridge"] = bz

		bz, err = json.Marshal(sequencingGenesis)
		require.NoError(t, err)
		cfg.GenesisState["sequencing"] = bz

	} else {
		cfg = configs[0]
	}
	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	_, err = net.WaitForHeight(1)
	require.NoError(t, err)
	t.Cleanup(net.Cleanup)
	return net
}

// DefaultConfig will initialize config for the network with custom application,
// genesis and single validator. All other parameters are inherited from cosmos-sdk/testutil/network.DefaultConfig
func DefaultConfig() network.Config {
	cfg, err := DefaultConfigWithAppConfig(fuelsequencerapp.AppConfig())
	if err != nil {
		panic(err)
	}
	ports, err := freePorts(3)
	if err != nil {
		panic(err)
	}
	if cfg.APIAddress == "" {
		cfg.APIAddress = fmt.Sprintf("tcp://0.0.0.0:%s", ports[0])
	}
	if cfg.RPCAddress == "" {
		cfg.RPCAddress = fmt.Sprintf("tcp://0.0.0.0:%s", ports[1])
	}
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = fmt.Sprintf("0.0.0.0:%s", ports[2])
	}
	return cfg
}

// freePorts return the available ports based on the number of requested ports.
func freePorts(n int) ([]string, error) {
	closeFns := make([]func() error, n)
	ports := make([]string, n)
	for i := 0; i < n; i++ {
		_, port, closeFn, err := network.FreeTCPAddr()
		if err != nil {
			return nil, err
		}
		ports[i] = port
		closeFns[i] = closeFn
	}
	for _, closeFn := range closeFns {
		if err := closeFn(); err != nil {
			return nil, err
		}
	}
	return ports, nil
}

func DefaultConfigWithAppConfig(appConfig depinject.Config) (Config, error) {
	var (
		appBuilder        *runtime.AppBuilder
		txConfig          client.TxConfig
		legacyAmino       *codec.LegacyAmino
		cdc               codec.Codec
		interfaceRegistry codectypes.InterfaceRegistry
	)

	if err := depinject.Inject(
		depinject.Configs(
			appConfig,
			depinject.Supply(log.NewNopLogger()),
		),
		&appBuilder,
		&txConfig,
		&cdc,
		&legacyAmino,
		&interfaceRegistry,
	); err != nil {
		return Config{}, err
	}

	cfg := network.DefaultConfig(func() network.TestFixture {
		return network.TestFixture{}
	})
	cfg.Codec = cdc
	cfg.TxConfig = txConfig
	cfg.LegacyAmino = legacyAmino
	cfg.InterfaceRegistry = interfaceRegistry
	cfg.GenesisState = appBuilder.DefaultGenesis()
	cfg.AppConstructor = func(val network.ValidatorI) servertypes.Application {
		app := &fuelsequencerapp.FuelSequencerApp{}

		// we build a unique app instance for every validator here
		var appBuilder *runtime.AppBuilder
		if err := depinject.Inject(
			depinject.Configs(
				appConfig,
				depinject.Supply(val.GetCtx().Logger),
			),
			&appBuilder); err != nil {
			panic(err)
		}
		app.App = appBuilder.Build(
			dbm.NewMemDB(),
			nil,
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.GetAppConfig().Pruning)),
			baseapp.SetMinGasPrices(val.GetAppConfig().MinGasPrices),
			baseapp.SetChainID(cfg.ChainID),
		)

		testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

		// PREPARE AND PROCESS PROPOSAL HANDLERS
		proposalHandler := abci.NewFuelSequencerProposalHandler(app.Logger(), app.StakingKeeper, app)
		app.SetPrepareProposal(proposalHandler.PrepareProposalHandler())
		app.SetProcessProposal(proposalHandler.ProcessProposalHandler())

		// SET mempool to NoOp. This is required for PrepareProposal and ProcessProposal to work as expected.
		app.SetMempool(mempool.NoOpMempool{})

		if err := app.Load(true); err != nil {
			panic(err)
		}

		return app
	}

	return cfg, nil
}
