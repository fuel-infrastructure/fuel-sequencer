package sequencing

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Sprintf("error when setting params: %x", err))
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
