package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	_ "cosmossdk.io/api/cosmos/tx/config/v1" // import for side-effects
	"cosmossdk.io/core/address"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	_ "cosmossdk.io/x/evidence" // import for side-effects
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	_ "cosmossdk.io/x/upgrade" // import for side-effects
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/types/module"
	_ "github.com/cosmos/cosmos-sdk/x/auth" // import for side-effects
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"   // import for side-effects
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/authz/module" // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/bank"         // import for side-effects
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/consensus" // import for side-effects
	consensuskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/distribution" // import for side-effects
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/slashing" // import for side-effects
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	_ "github.com/cosmos/cosmos-sdk/x/staking" // import for side-effects
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	appcodec "github.com/fuel-infrastructure/fuel-sequencer/app/codec"
	"github.com/fuel-infrastructure/fuel-sequencer/app/upgrades/multi_vesting_accounts"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	blobconfig "github.com/fuel-infrastructure/fuel-sequencer/x/blob/config"
	commitmentsconfig "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/config"
	commitmentsservice "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/service"
	_ "github.com/fuel-infrastructure/fuel-sequencer/x/mint" // import for side-effects

	blobmodulekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/blob/keeper"
	bridgemodulekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	sequencingmodulekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"

	// this line is used by starport scaffolding # stargate/app/moduleImport

	"github.com/fuel-infrastructure/fuel-sequencer/client/docs"
)

const (
	AccountAddressPrefix = "fuelsequencer"
	Name                 = "fuelsequencer"
	PrettyName           = "FuelSequencer"
)

// DefaultNodeHome default home directories for the application daemon
var DefaultNodeHome string

var (
	_ runtime.AppI            = (*FuelSequencerApp)(nil)
	_ servertypes.Application = (*FuelSequencerApp)(nil)
)

// FuelSequencerApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object capabilities aren't needed for testing.
type FuelSequencerApp struct {
	*runtime.App
	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper

	SlashingKeeper slashingkeeper.Keeper
	MintKeeper     mintkeeper.Keeper
	GovKeeper      *govkeeper.Keeper
	UpgradeKeeper  *upgradekeeper.Keeper
	AuthzKeeper    authzkeeper.Keeper
	EvidenceKeeper evidencekeeper.Keeper

	BridgeKeeper     bridgemodulekeeper.Keeper
	SequencingKeeper sequencingmodulekeeper.Keeper
	BlobKeeper       blobmodulekeeper.Keeper
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

	// simulation manager
	sm *module.SimulationManager

	// sidecar
	sidecar sidecarclient.AppSidecarClient

	// blob
	blobConfig blobconfig.Config

	// commitments
	commitmentsConfig commitmentsconfig.Config
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)

	// DefaultPowerReduction is the amount of staking tokens required for 1 unit of consensus-engine power.
	// We change this to 1e9 to match the number of decimal places used in the bridged token denomination.
	// There are plans to change this to an on-chain param: https://github.com/cosmos/cosmos-sdk/issues/8365
	// Note: This should be revised if the number of decimals for the bridged token changes.
	sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000000)
}

// getGovProposalHandlers return the chain proposal handlers.
// Deprecated: This function is added just in case we need to register any module param handlers with the gov module.
func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler
	// this line is used by starport scaffolding # stargate/app/govProposalHandlers

	//nolint:staticcheck
	govProposalHandlers = append(govProposalHandlers) // this line is used by starport scaffolding # stargate/app/govProposalHandler

	return govProposalHandlers
}

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appConfig,
		// Loads the app config from a YAML file.
		// appconfig.LoadYAML(AppConfigYAML),
		depinject.Supply(
			// supply custom module basics
			map[string]module.AppModuleBasic{
				genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
				govtypes.ModuleName:     gov.NewAppModuleBasic(getGovProposalHandlers()),
				// this line is used by starport scaffolding # stargate/appConfig/moduleBasic
			},
		),
	)
}

// NewFuelSequencerApp returns a reference to an initialized Fuel Sequencer App.
func NewFuelSequencerApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	initialiseBlobhub bool,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) (*FuelSequencerApp, error) {
	var (
		app        = &FuelSequencerApp{}
		appBuilder *runtime.AppBuilder

		// merge the AppConfig and other configuration in one config
		appConfig = depinject.Configs(
			AppConfig(),
			depinject.Supply(
				// Supply the application options
				appOpts,

				// Supply the logger
				logger,

				// ADVANCED CONFIGURATION
				//
				// AUTH
				//
				// For providing a custom function required in auth to generate custom account types
				// add it below. By default the auth module uses simulation.RandomGenesisAccounts.
				//
				// authtypes.RandomGenesisAccountsFn(simulation.RandomGenesisAccounts),
				//
				// For providing a custom a base account type add it below.
				// By default the auth module uses authtypes.ProtoBaseAccount().
				//
				// func() sdk.AccountI { return authtypes.ProtoBaseAccount() },
				//
				// For providing a different address codec, add it below.
				// By default the auth module uses a Bech32 address codec,
				// with the prefix defined in the auth module configuration.
				//
				func() address.Codec {
					return appcodec.NewFuelSequencerAddressCodec(sdkAddressCodec.NewBech32Codec(AccountAddressPrefix))
				},

				//
				// STAKING
				//
				// For providing a different validator and consensus address codec, add it below.
				// By default, the staking module uses the bech32 prefix provided in the auth config,
				// and appends "valoper" and "valcons" for validator and consensus addresses respectively.
				// When providing a custom address codec in auth, custom address codecs must be provided here as well.
				//
				func() runtime.ValidatorAddressCodec {
					return appcodec.NewFuelSequencerAddressCodec(
						sdkAddressCodec.NewBech32Codec(AccountAddressPrefix + "valoper"),
					)
				},
				func() runtime.ConsensusAddressCodec {
					return appcodec.NewFuelSequencerAddressCodec(
						sdkAddressCodec.NewBech32Codec(AccountAddressPrefix + "valcons"),
					)
				},

				//
				// MINT
				//
				// For providing a custom inflation function for x/mint add here your
				// custom function that implements the minttypes.InflationCalculationFn
				// interface.
			),
		)
	)

	if err := depinject.Inject(appConfig,
		&appBuilder,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AccountKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.DistrKeeper,
		&app.ConsensusParamsKeeper,
		&app.SlashingKeeper,
		&app.MintKeeper,
		&app.GovKeeper,
		&app.UpgradeKeeper,
		&app.AuthzKeeper,
		&app.EvidenceKeeper,
		&app.BridgeKeeper,
		&app.SequencingKeeper,
		&app.BlobKeeper,
		// this line is used by starport scaffolding # stargate/app/keeperDefinition
	); err != nil {
		panic(err)
	}

	// Below we could construct and set an application specific mempool and
	// ABCI 1.0 PrepareProposal and ProcessProposal handlers. These defaults are
	// already set in the SDK's BaseApp, this shows an example of how to override
	// them.
	//
	// Example:
	//
	// app.App = appBuilder.Build(...)
	// nonceMempool := mempool.NewSenderNonceMempool()
	// abciPropHandler := NewDefaultProposalHandler(nonceMempool, app.App.BaseApp)
	//
	// app.App.BaseApp.SetMempool(nonceMempool)
	// app.App.BaseApp.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// app.App.BaseApp.SetProcessProposal(abciPropHandler.ProcessProposalHandler())
	//
	// Alternatively, you can construct BaseApp options, append those to
	// baseAppOptions and pass them to the appBuilder.
	//
	// Example:
	//
	// prepareOpt = func(app *baseapp.BaseApp) {
	// 	abciPropHandler := baseapp.NewDefaultProposalHandler(nonceMempool, app)
	// 	app.SetPrepareProposal(abciPropHandler.PrepareProposalHandler())
	// }
	// baseAppOptions = append(baseAppOptions, prepareOpt)
	//
	// create and set vote extension handler
	// voteExtOp := func(bApp *baseapp.BaseApp) {
	// 	voteExtHandler := NewVoteExtensionHandler()
	// 	voteExtHandler.SetHandlers(bApp)
	// }

	app.App = appBuilder.Build(db, traceStore, baseAppOptions...)

	// COMMITMENTS :: Configure
	commitmentsCfg, err := commitmentsconfig.NewConfigFromAppOptions(appOpts)
	if err != nil {
		panic(err)
	}
	err = commitmentsCfg.ValidateBasic()
	if err != nil {
		panic(err)
	}
	app.commitmentsConfig = commitmentsCfg

	// BLOB :: Configure
	blobCfg, err := blobconfig.NewConfigFromAppOptions(appOpts)
	if err != nil {
		panic(err)
	}
	err = blobCfg.ValidateBasic()
	if err != nil {
		panic(err)
	}
	app.blobConfig = blobCfg

	// Set the blobhub address on the keeper
	app.BlobKeeper.SetBlobhubAddress(blobCfg.BlobhubAddress)
	app.BlobKeeper.SetBlobpoolRedisAddress(blobCfg.BlobpoolRedisAddress)

	if initialiseBlobhub {
		// BLOB :: Initialize blobhub and blobpool connections
		if err := app.BlobKeeper.Initialize(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to initialize blob keeper: %w", err)
		}
	}

	// SIDECAR :: Configure
	sidecarCfg, err := sidecarconfig.NewConfigFromAppOptions(appOpts)
	if err != nil {
		panic(err)
	}

	// SIDECAR :: Create client
	app.sidecar, err = sidecarclient.NewClientFromConfig(
		sidecarCfg,
		app.Logger().With("client", "sidecar"),
	)
	if err != nil {
		panic(err)
	}

	// SIDECAR :: Connect to the client if the Sidecar is enabled
	if sidecarCfg.Enabled {
		go func() {
			if err := app.sidecar.Start(context.Background()); err != nil {
				app.Logger().Error("failed to start Sidecar client", "err", err)
				panic(err)
			}

			app.Logger().Info("started Sidecar client", "sidecar server address", sidecarCfg.Address)
		}()
	}

	app.UpgradeKeeper.SetUpgradeHandler(
		multi_vesting_accounts.UpgradeName,
		multi_vesting_accounts.CreateUpgradeHandler(app.ModuleManager, app.Configurator()),
	)

	// PREPARE AND PROCESS PROPOSAL HANDLERS
	proposalHandler := abci.NewFuelSequencerProposalHandler(
		app.appCodec, app.StakingKeeper, app, app.sidecar, app.BridgeKeeper, app.BlobKeeper,
	)
	app.SetPrepareProposal(proposalHandler.PrepareProposalHandler())
	app.SetProcessProposal(proposalHandler.ProcessProposalHandler())

	// ANTEHANDLER
	anteHandler, err := NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: app.txConfig.SignModeHandler(),
			FeegrantKeeper:  nil,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		},
		app.BridgeKeeper,
		app.SequencingKeeper,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ante handler: %w", err)
	}
	app.SetAnteHandler(anteHandler)

	// SET mempool to NoOp. This is required for PrepareProposal and ProcessProposal to work as expected.
	app.SetMempool(mempool.NoOpMempool{})

	// Register legacy modules

	// register streaming services
	if err := app.RegisterStreamingServices(appOpts, app.kvStoreKeys()); err != nil {
		return nil, err
	}

	/****  Module Options ****/

	// CrisisKeeper is not wired, therefore, no invariants are registered.
	// Justification: https://github.com/cosmos/cosmos-sdk/issues/15706
	//app.ModuleManager.RegisterInvariants(app.CrisisKeeper)

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required for apps that don't use the simulator for fuzz testing transactions
	// overrideModules := map[string]module.AppModuleSimulation{
	// 	authtypes.ModuleName: auth.NewAppModule(
	// 		app.appCodec,
	// 		app.AccountKeeper,
	// 		authsims.RandomGenesisAccounts,
	// 		app.GetSubspace(authtypes.ModuleName),
	// 	),
	// }
	// app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	// app.sm.RegisterStoreDecoders()

	// A custom InitChainer can be set if extra pre-init-genesis logic is required.
	// By default, when using app wiring enabled module, this is not required.
	// For instance, the upgrade module will set automatically the module version map
	// in its init genesis thanks to app wiring.
	// However, when registering a module manually (i.e. that does not support app wiring), the module version map
	// must be set manually as follow. The upgrade module will de-duplicate the module version map.
	//
	// app.SetInitChainer(func(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	// 	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	// 	return app.App.InitChainer(ctx, req)
	// })

	if err := app.Load(loadLatest); err != nil {
		return nil, err
	}

	return app, nil
}

// LegacyAmino returns App's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *FuelSequencerApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns App's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *FuelSequencerApp) AppCodec() codec.Codec {
	return app.appCodec
}

// GetKey returns the KVStoreKey for the provided store key.
func (app *FuelSequencerApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	kvStoreKey, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.KVStoreKey)
	if !ok {
		return nil
	}
	return kvStoreKey
}

// GetMemKey returns the MemoryStoreKey for the provided store key.
func (app *FuelSequencerApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	key, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.MemoryStoreKey)
	if !ok {
		return nil
	}

	return key
}

// kvStoreKeys returns all the kv store keys registered inside App.
func (app *FuelSequencerApp) kvStoreKeys() map[string]*storetypes.KVStoreKey {
	keys := make(map[string]*storetypes.KVStoreKey)
	for _, k := range app.GetStoreKeys() {
		if kv, ok := k.(*storetypes.KVStoreKey); ok {
			keys[kv.Name()] = kv
		}
	}

	return keys
}

// SimulationManager implements the SimulationApp interface.
func (app *FuelSequencerApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *FuelSequencerApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	app.App.RegisterAPIRoutes(apiSvr, apiConfig)
	// register swagger API in app.go so that other applications can override easily
	if err := docs.RegisterSwaggerAPI(apiSvr.ClientCtx, apiSvr.Router, apiConfig.Swagger); err != nil {
		panic(err)
	}

	if app.commitmentsConfig.ApiEnabled {
		// Register bridge commitments routes.
		commitmentsservice.RegisterGRPCGatewayRoutes(apiSvr.ClientCtx, apiSvr.GRPCGatewayRouter)
	}
}

func (app *FuelSequencerApp) RegisterTendermintService(clientCtx client.Context) {
	app.App.RegisterTendermintService(clientCtx)

	if app.commitmentsConfig.ApiEnabled {
		commitmentsservice.RegisterCommitmentsService(
			clientCtx,
			app.GRPCQueryRouter(),
			app.interfaceRegistry,
			app.commitmentsConfig.MaxQueryRange,
		)
	}
}

// Close closes the underlying baseapp and the Sidecar service.
// This function blocks on the closure of the Sidecar service.
func (app *FuelSequencerApp) Close() error {
	if err := app.App.Close(); err != nil {
		return err
	}

	// close the Sidecar service
	if app.sidecar != nil {
		app.Logger().Info("stopping Sidecar")
		if err := app.sidecar.Stop(); err != nil {
			app.Logger().Error("error when stopping sidecar", "err", err)
		}
	}

	return nil
}

// NewTxBuilder returns a new instance of TxBuilder. This was added for testing purposes.
func (app *FuelSequencerApp) NewTxBuilder() client.TxBuilder {
	return app.txConfig.NewTxBuilder()
}

// GetTxEncoder returns the underlying TxEncoder. This was added for testing purposes.
func (app *FuelSequencerApp) GetTxEncoder() sdk.TxEncoder {
	return app.txConfig.TxEncoder()
}

// GetMaccPerms returns a copy of the module account permissions
//
// NOTE: This is solely to be used for testing purposes.
func GetMaccPerms() map[string][]string {
	dup := make(map[string][]string)
	for _, perms := range moduleAccPerms {
		dup[perms.Account] = perms.Permissions
	}
	return dup
}

// BlockedAddresses returns all the app's blocked account addresses.
func BlockedAddresses() map[string]bool {
	result := make(map[string]bool)
	if len(blockAccAddrs) > 0 {
		for _, addr := range blockAccAddrs {
			result[addr] = true
		}
	} else {
		for addr := range GetMaccPerms() {
			result[addr] = true
		}
	}
	return result
}
