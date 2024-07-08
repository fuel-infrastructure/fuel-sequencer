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
	mintModuleOut := mint.ProvideModule(mint.ModuleInputs{
		In:                     in.In,
		ModuleKey:              in.ModuleKey,
		Config:                 in.Config,
		StoreService:           in.StoreService,
		Cdc:                    in.Cdc,
		InflationCalculationFn: in.InflationCalculationFn,
		LegacySubspace:         in.LegacySubspace,
		AccountKeeper:          in.AccountKeeper,
		BankKeeper:             in.BankKeeper,
		StakingKeeper:          in.StakingKeeper,
	})

	// override mint module's AppModule with custom one
	mintModuleOut.Module = NewAppModule(
		in.Cdc,
		mintModuleOut.MintKeeper,
		in.AccountKeeper,
		in.BridgeKeeper,
		in.InflationCalculationFn,
		in.LegacySubspace,
	)

	return mintModuleOut
}
