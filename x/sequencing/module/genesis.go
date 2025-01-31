package sequencing

import (
	"context"
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) error {

	// Set all the topics
	for _, topic := range genState.TopicList {

		// Panic if topic fails validation
		if err := topic.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid topic list: %w", err)
		}

		k.SetTopic(ctx, topic)
	}

	if err := k.SetParams(ctx, genState.Params); err != nil {
		return fmt.Errorf("error when setting params: %x", err)
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) (*types.GenesisState, error) {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.TopicList = k.GetAllTopic(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis, nil
}
