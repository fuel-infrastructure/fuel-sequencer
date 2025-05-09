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

// AppModule is an application module that incorporates the Cosmos SDK mint module, to override some of its logic,
// namely the BeginBlock, which is overridden so that inflation is based on a supply value reported by Bridge module.
type AppModule struct {
	mint.AppModule

	// The below fields are not exported by mint.AppModule, but they are needed to call the BeginBlocker in BeginBlock,
	// so we have copies of them here that we can set in NewAppModule for usage in BeginBlock.
	keeper       mintkeeper.Keeper
	bridgeKeeper types.BridgeKeeper
}

// NewAppModule creates a new AppModule object. If the InflationCalculationFn
// argument is nil, then the SDK's default inflation function will be used.
// In our case, we ignore InflationCalculationFn in the BeginBlocker anyway.
func NewAppModule(
	cdc codec.Codec,
	keeper mintkeeper.Keeper,
	ak minttypes.AccountKeeper,
	bk types.BridgeKeeper,
	ss exported.Subspace,
) AppModule {

	ic := minttypes.InflationCalculationFn(nil)
	return AppModule{
		AppModule:    mint.NewAppModule(cdc, keeper, ak, ic, ss),
		keeper:       keeper,
		bridgeKeeper: bk,
	}
}

// BeginBlock overrides the BeginBlock of the mint module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return BeginBlocker(ctx, am.keeper, am.bridgeKeeper)
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

// ModuleInputs is almost identical to the original mint module ModuleInputs, but adds a BridgeKeeper dependency.
type ModuleInputs struct {
	depinject.In

	ModuleKey    depinject.OwnModuleKey
	Config       *modulev1.Module
	StoreService store.KVStoreService
	Cdc          codec.Codec

	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace exported.Subspace `optional:"true"`

	AccountKeeper minttypes.AccountKeeper
	BankKeeper    minttypes.BankKeeper
	StakingKeeper minttypes.StakingKeeper
	BridgeKeeper  types.BridgeKeeper
}

// ProvideModule calls the original mint module ProvideModule but then overrides the AppModule with the custom one.
func ProvideModule(in ModuleInputs) mint.ModuleOutputs {
	out := mint.ProvideModule(mint.ModuleInputs{
		In:             in.In,
		ModuleKey:      in.ModuleKey,
		Config:         in.Config,
		StoreService:   in.StoreService,
		Cdc:            in.Cdc,
		LegacySubspace: in.LegacySubspace,
		AccountKeeper:  in.AccountKeeper,
		BankKeeper:     in.BankKeeper,
		StakingKeeper:  in.StakingKeeper,
	})

	// override mint module's AppModule with custom one
	out.Module = NewAppModule(
		in.Cdc,
		out.MintKeeper,
		in.AccountKeeper,
		in.BridgeKeeper,
		in.LegacySubspace,
	)

	return out
}
