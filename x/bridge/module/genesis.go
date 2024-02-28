package bridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	//nolint:errcheck
	k.SetParams(ctx, genState.Params)

	k.SetLastEthereumNonce(ctx, genState.LastEthereumNonce)
	k.SetLastEthereumBlockSynced(ctx, genState.LastEthereumBlockSynced)

	// this line is used by starport scaffolding # genesis/module/init
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	lastEthereumNonce, found := k.GetLastEthereumNonce(ctx)
	if found {
		genesis.LastEthereumNonce = lastEthereumNonce
	}

	lastEthereumBlockSynced, found := k.GetLastEthereumBlockSynced(ctx)
	if found {
		genesis.LastEthereumBlockSynced = lastEthereumBlockSynced
	}

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
