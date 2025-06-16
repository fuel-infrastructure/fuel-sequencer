package bond

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bond/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Initialize module account
	if acc := k.GetAccountKeeper().GetModuleAccount(ctx, types.ModuleName); acc == nil {
		panic(fmt.Sprintf("failed to get %s module account", types.ModuleName))
	}

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Sprintf("error when setting params: %x", err))
	}

	// Initialize state
	if err := k.SetState(ctx, genState.State); err != nil {
		panic(fmt.Sprintf("error when setting state: %x", err))
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.State = k.GetState(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
