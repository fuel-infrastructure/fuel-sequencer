package app

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"cosmossdk.io/core/appmodule"
	govkeeper "cosmossdk.io/x/gov/keeper"
	mintkeeper "cosmossdk.io/x/mint/keeper"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	commitmentsservice "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/service"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint"
	reportskeeper "github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	sequencingkeeper "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	_ "github.com/jackc/pgx/v5/stdlib" // Import and register pgx driver

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/registry"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	_ "cosmossdk.io/indexer/postgres" // register the postgres indexer
	"cosmossdk.io/log"
	"cosmossdk.io/x/accounts"
	basedepinject "cosmossdk.io/x/accounts/defaults/base/depinject"
	lockupdepinject "cosmossdk.io/x/accounts/defaults/lockup/depinject"
	multisigdepinject "cosmossdk.io/x/accounts/defaults/multisig/depinject"
	authzkeeper "cosmossdk.io/x/authz/keeper"
	bankkeeper "cosmossdk.io/x/bank/keeper"
	consensuskeeper "cosmossdk.io/x/consensus/keeper"
	distrkeeper "cosmossdk.io/x/distribution/keeper"
	evidencekeeper "cosmossdk.io/x/evidence/keeper"
	_ "cosmossdk.io/x/protocolpool"
	slashingkeeper "cosmossdk.io/x/slashing/keeper"
	stakingkeeper "cosmossdk.io/x/staking/keeper"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdkAddressCodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	appcodec "github.com/fuel-infrastructure/fuel-sequencer/app/codec"
	"github.com/fuel-infrastructure/fuel-sequencer/client/docs"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	commitmentsconfig "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/config"
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
	legacyAmino       registry.AminoRegistrar
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry codectypes.InterfaceRegistry

	// required keepers during wiring
	// others keepers are all in the app
	AccountsKeeper accounts.Keeper

	AuthKeeper            authkeeper.AccountKeeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	ConsensusParamsKeeper consensuskeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper

	MintKeeper     mintkeeper.Keeper // TODO: remove?
	GovKeeper      *govkeeper.Keeper // TODO: remove?
	UpgradeKeeper  *upgradekeeper.Keeper
	AuthzKeeper    authzkeeper.Keeper    // TODO: remove?
	EvidenceKeeper evidencekeeper.Keeper // TODO: remove?

	BridgeKeeper     bridgekeeper.Keeper
	SequencingKeeper sequencingkeeper.Keeper
	ReportsKeeper    reportskeeper.Keeper
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

	// simulation manager
	sm *module.SimulationManager

	// sidecar
	sidecar sidecarclient.AppSidecarClient

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

// AppConfig returns the default app config.
func AppConfig() depinject.Config {
	return depinject.Configs(
		appConfig,
		// Loads the app config from a YAML file.
		// appconfig.LoadYAML(AppConfigYAML),
		depinject.Provide(
			mint.ProvideMintFn, // override the mint module's mint function with custom minting logic`
		),
	)
}

// NewFuelSequencerApp returns a reference to an initialized Fuel Sequencer App.
func NewFuelSequencerApp(
	logger log.Logger,
	db corestore.KVStoreWithBatch,
	traceStore io.Writer,
	loadLatest bool,
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
				// For provinding a different validator and consensus address codec, add it below.
				// By default the staking module uses the bech32 prefix provided in the auth config,
				// and appends "valoper" and "valcons" for validator and consensus addresses respectively.
				// When providing a custom address codec in auth, custom address codecs must be provided here as well.
				//
				func() address.ValidatorAddressCodec {
					return appcodec.NewFuelSequencerAddressCodec(
						sdkAddressCodec.NewBech32Codec(AccountAddressPrefix + "valoper"),
					)
				},
				func() address.ConsensusAddressCodec {
					return appcodec.NewFuelSequencerAddressCodec(
						sdkAddressCodec.NewBech32Codec(AccountAddressPrefix + "valcons"),
					)
				},

				//
				// MINT
				//
				// For providing a custom inflation function for x/mint add here your
				// custom function that implements the minttypes.MintFn interface.
			),
			depinject.Provide(
				// inject desired account types:
				multisigdepinject.ProvideAccount,
				basedepinject.ProvideAccount,
				lockupdepinject.ProvideAllLockupAccounts,
				// TODO: Ethreum-Owned accounts?

				// provide base account options
				basedepinject.ProvideSecp256K1PubKey,
				// if you want to provide a custom public key you
				// can do it from here.
				// Example:
				// 		basedepinject.ProvideCustomPubkey[Ed25519PublicKey]()
				//
				// You can also provide a custom public key with a custom validation function:
				//
				// 		basedepinject.ProvideCustomPubKeyAndValidationFunc(func(pub Ed25519PublicKey) error {
				//			if len(pub.Key) != 64 {
				//				return fmt.Errorf("invalid pub key size")
				//			}
				// 		})
			),
		)
	)

	var appModules map[string]appmodule.AppModule
	if err := depinject.Inject(appConfig,
		&appBuilder,
		&appModules,
		&app.appCodec,
		&app.legacyAmino,
		&app.txConfig,
		&app.interfaceRegistry,
		&app.AuthKeeper,
		&app.AccountsKeeper,
		&app.BankKeeper,
		&app.StakingKeeper,
		&app.DistrKeeper,
		&app.ConsensusParamsKeeper,
		&app.SlashingKeeper,
		// TODO: remove? &app.MintKeeper,
		// TODO: remove? &app.GovKeeper,
		&app.UpgradeKeeper,
		// TODO: remove? &app.AuthzKeeper,
		// TODO: remove? &app.EvidenceKeeper,
		&app.BridgeKeeper,
		&app.SequencingKeeper,
		&app.ReportsKeeper,
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

	//app.UpgradeKeeper.SetUpgradeHandler(
	//	vesting_accounts_staking.UpgradeName,
	//	vesting_accounts_staking.CreateUpgradeHandler(app.ModuleManager, app.Configurator()),
	//)

	// PREPARE AND PROCESS PROPOSAL HANDLERS
	proposalHandler := abci.NewFuelSequencerProposalHandler(
		app.appCodec, app.StakingKeeper, app, app.sidecar, app.BridgeKeeper,
	)
	app.SetPrepareProposal(proposalHandler.PrepareProposalHandler())
	app.SetProcessProposal(proposalHandler.ProcessProposalHandler())

	// SET mempool to NoOp. This is required for PrepareProposal and ProcessProposal to work as expected.
	app.SetMempool(mempool.NoOpMempool{})

	/****  Module Options ****/

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required for apps that don't use the simulator for fuzz testing transactions
	//overrideModules := map[string]module.AppModuleSimulation{
	//	authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AuthKeeper, &app.AccountsKeeper, authsims.RandomGenesisAccounts, nil),
	//}
	//app.sm = module.NewSimulationManagerFromAppModules(app.ModuleManager.Modules, overrideModules)
	//app.sm.RegisterStoreDecoders()

	// A custom InitChainer can be set if extra pre-init-genesis logic is required.
	// By default, when using app wiring enabled module, this is not required.
	// For instance, the upgrade module will set automatically the module version map in its init genesis thanks to app wiring.
	// However, when registering a module manually (i.e. that does not support app wiring), the module version map
	// must be set manually as follow. The upgrade module will de-duplicate the module version map.
	//
	// app.SetInitChainer(func(ctx sdk.Context, req *abci.InitChainRequest) (*abci.InitChainResponse, error) {
	// 	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.ModuleManager.GetVersionMap())
	// 	return app.App.InitChainer(ctx, req)
	// })

	//// register custom snapshot extensions (if any)
	//if manager := app.SnapshotManager(); manager != nil {
	//	if err := manager.RegisterExtensions(
	//		unorderedtx.NewSnapshotter(app.UnorderedTxManager),
	//	); err != nil {
	//		panic(fmt.Errorf("failed to register snapshot extension: %w", err))
	//	}
	//}

	// ANTEHANDLER
	anteHandler, err := NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:            app.AuthKeeper,
			BankKeeper:               app.BankKeeper,
			ConsensusKeeper:          app.ConsensusParamsKeeper,
			SignModeHandler:          app.txConfig.SignModeHandler(),
			FeegrantKeeper:           nil,
			SigGasConsumer:           ante.DefaultSigVerificationGasConsumer,
			UnorderedTxManager:       app.UnorderedTxManager,
			Environment:              app.AuthKeeper.Environment,
			AccountAbstractionKeeper: app.AccountsKeeper,
		},
		app.BridgeKeeper,
		app.SequencingKeeper,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ante handler: %w", err)
	}
	app.SetAnteHandler(anteHandler)

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
	switch cdc := app.legacyAmino.(type) {
	case *codec.LegacyAmino:
		return cdc
	default:
		panic("unexpected codec type")
	}
}

// AppCodec returns App's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *FuelSequencerApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns FuelSequencerApp's InterfaceRegistry.
func (app *FuelSequencerApp) InterfaceRegistry() codectypes.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetMemKey returns the MemoryStoreKey for the provided store key.
func (app *FuelSequencerApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	key, ok := app.UnsafeFindStoreKey(storeKey).(*storetypes.MemoryStoreKey)
	if !ok {
		return nil
	}

	return key
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
		commitmentsservice.RegisterCommitmentsService(clientCtx, app.GRPCQueryRouter(), app.interfaceRegistry, app.commitmentsConfig.MaxQueryRange)
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
// This function takes an address.Codec parameter to maintain compatibility
// with the signature of the same function in appV1.
func BlockedAddresses(_ address.Codec) (map[string]bool, error) {
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

	return result, nil
}
