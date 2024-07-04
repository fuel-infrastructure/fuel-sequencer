package mint

import (
	"context"

	modulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/mint/exported"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/mint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// AppModule implements an application module for the mint module.
type AppModule struct {
	mint.AppModule

	keeper       mintkeeper.Keeper
	bridgeKeeper types.BridgeKeeper

	// inflationCalculator is used to calculate the inflation rate during BeginBlock.
	// If inflationCalculator is nil, the default inflation calculation logic is used.
	inflationCalculator minttypes.InflationCalculationFn
}

// NewAppModule creates a new AppModule object. If the InflationCalculationFn
// argument is nil, then the SDK's default inflation function will be used.
func NewAppModule(
	cdc codec.Codec,
	keeper mintkeeper.Keeper,
	ak minttypes.AccountKeeper,
	bk types.BridgeKeeper,
	ic minttypes.InflationCalculationFn,
	ss exported.Subspace,
) AppModule {
	if ic == nil {
		ic = minttypes.DefaultInflationCalculationFn
	}

	return AppModule{
		AppModule:           mint.NewAppModule(cdc, keeper, ak, ic, ss),
		keeper:              keeper,
		bridgeKeeper:        bk,
		inflationCalculator: ic,
	}
}

// BeginBlock returns the begin blocker for the mint module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return BeginBlocker(ctx, am.keeper, am.bridgeKeeper, am.inflationCalculator)
}

//
// App Wiring Setup
//

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	ModuleKey              depinject.OwnModuleKey
	Config                 *modulev1.Module
	StoreService           store.KVStoreService
	Cdc                    codec.Codec
	InflationCalculationFn minttypes.InflationCalculationFn `optional:"true"`

	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace exported.Subspace `optional:"true"`

	AccountKeeper minttypes.AccountKeeper
	BankKeeper    minttypes.BankKeeper
	StakingKeeper minttypes.StakingKeeper
	BridgeKeeper  types.BridgeKeeper
}

func ProvideModule(in ModuleInputs) mint.ModuleOutputs {
	feeCollectorName := in.Config.FeeCollectorName
	if feeCollectorName == "" {
		feeCollectorName = authtypes.FeeCollectorName
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := mintkeeper.NewKeeper(
		in.Cdc,
		in.StoreService,
		in.StakingKeeper,
		in.AccountKeeper,
		in.BankKeeper,
		feeCollectorName,
		authority.String(),
	)

	// when no inflation calculation function is provided it will use the default minttypes.DefaultInflationCalculationFn
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BridgeKeeper, in.InflationCalculationFn, in.LegacySubspace)

	return mint.ModuleOutputs{MintKeeper: k, Module: m}
}
