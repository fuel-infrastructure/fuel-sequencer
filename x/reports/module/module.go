package reports

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/log"
	govtypes "cosmossdk.io/x/gov/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	// this line is used by starport scaffolding # 1

	stakingtypes "cosmossdk.io/x/staking/types"
	modulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/reports/module"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// ConsensusVersion defines the current x/reports module consensus version.
const ConsensusVersion = 1

var (
	_ module.HasGRPCGateway      = AppModule{}
	_ module.HasRegisterServices = AppModule{}

	_ appmodule.AppModule             = AppModule{}
	_ appmodule.HasConsensusVersion   = AppModule{}
	_ appmodule.HasGenesis            = AppModule{}
	_ appmodule.HasBeginBlocker       = AppModule{}
	_ appmodule.HasEndBlocker         = AppModule{}
	_ appmodule.HasRegisterInterfaces = AppModule{}
)

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	cdc           codec.Codec
	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) AppModule {
	return AppModule{
		cdc:           cdc,
		keeper:        keeper,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// Name returns the name of the module as a string.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message.
func (AppModule) RegisterInterfaces(reg registry.InterfaceRegistrar) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage.
// The default GenesisState need to be defined by the module developer and is primarily used for testing.
func (am AppModule) DefaultGenesis() json.RawMessage {
	return am.cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis used to validate the GenesisState, given in its json.RawMessage form.
func (am AppModule) ValidateGenesis(bz json.RawMessage) error {
	var genState types.GenesisState
	if err := am.cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (AppModule) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, gs json.RawMessage) error {
	var genState types.GenesisState
	if err := am.cdc.UnmarshalJSON(gs, &genState); err != nil {
		return err
	}
	return InitGenesis(ctx, am.keeper, genState)
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	genState, err := ExportGenesis(ctx, am.keeper)
	if err != nil {
		return nil, err
	}
	return am.cdc.MarshalJSON(genState)
}

// ConsensusVersion is a sequence number for state-breaking change of the module.
// It should be incremented on each consensus-breaking change introduced by the module.
// To avoid wrong/empty versions, the initial version should be set to 1.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock contains the logic that is automatically triggered at the beginning of each block.
// The begin block implementation is optional.
func (am AppModule) BeginBlock(goCtx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyBeginBlocker)

	ctx := sdk.UnwrapSDKContext(goCtx)

	return am.keeper.UpdateSlashReportBalancesAtCurrentHeight(ctx)
}

// EndBlock contains the logic that is automatically triggered at the end of each block.
// The end block implementation is optional.
func (am AppModule) EndBlock(goCtx context.Context) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, telemetry.Now(), telemetry.MetricKeyEndBlocker)

	ctx := sdk.UnwrapSDKContext(goCtx)

	return am.keeper.PruneSlashReports(ctx)
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// ----------------------------------------------------------------------------
// App Wiring Setup
// ----------------------------------------------------------------------------

func init() {
	appconfig.RegisterModule(
		&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Cdc         codec.Codec
	Environment appmodule.Environment
	Config      *modulev1.Module
	Logger      log.Logger

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
}

type ModuleOutputs struct {
	depinject.Out

	ReportsKeeper keeper.Keeper
	Module        appmodule.AppModule
	Hooks         stakingtypes.StakingHooksWrapper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}
	authorityAddr, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(err)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Environment,
		in.Logger,
		in.StakingKeeper,
		authorityAddr,
	)
	m := NewAppModule(
		in.Cdc,
		k,
		in.AccountKeeper,
		in.BankKeeper,
	)

	return ModuleOutputs{ReportsKeeper: k, Module: m, Hooks: stakingtypes.StakingHooksWrapper{StakingHooks: k.Hooks()}}
}
