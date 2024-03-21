package bridge

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	defaults := types.DefaultGenesis()

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Sprintf("error when setting params: %x", err))
	}

	// Set if defined, otherwise use default
	if genState.SupplyDeltaInfo != nil {
		k.SetSupplyDeltaInfo(ctx, *genState.SupplyDeltaInfo)
	} else {
		k.SetSupplyDeltaInfo(ctx, *defaults.SupplyDeltaInfo)
	}

	k.SetLastEthereumNonce(ctx, genState.LastEthereumNonce)
	k.SetLastEthereumBlockSynced(ctx, genState.LastEthereumBlockSynced)

	// Set if defined
	if genState.SupplyDeltaProcessed != nil {
		k.SetSupplyDeltaProcessed(ctx, *genState.SupplyDeltaProcessed)
	} else {
		k.SetSupplyDeltaProcessed(ctx, *defaults.SupplyDeltaProcessed)
	}
	// this line is used by starport scaffolding # genesis/module/init
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	supplyDeltaInfo, found := k.GetSupplyDeltaInfo(ctx)
	if found {
		genesis.SupplyDeltaInfo = &supplyDeltaInfo
	}

	lastEthereumNonce, found := k.GetLastEthereumNonce(ctx)
	if found {
		genesis.LastEthereumNonce = lastEthereumNonce
	}

	lastEthereumBlockSynced, found := k.GetLastEthereumBlockSynced(ctx)
	if found {
		genesis.LastEthereumBlockSynced = lastEthereumBlockSynced
	}

	// Get all supplyDeltaProcessed
	supplyDeltaProcessed, found := k.GetSupplyDeltaProcessed(ctx)
	if found {
		genesis.SupplyDeltaProcessed = &supplyDeltaProcessed
	}
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
