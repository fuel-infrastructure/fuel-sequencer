package sequencing

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {

	largestTopicId := math.ZeroInt()
	// Set all the topics
	for _, topic := range genState.TopicList {

		// Panic if topic fails validation
		if err := topic.ValidateBasic(); err != nil {
			panic(err)
		}

		k.SetTopic(ctx, topic)

		if topic.Id.GT(largestTopicId) {
			largestTopicId = topic.Id
		}
	}

	// Set the next topic id as current largest topic Id + 1
	k.SetNextTopicId(ctx, largestTopicId.Add(math.OneInt()))

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Sprintf("error when setting params: %x", err))
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.TopicList = k.GetAllTopic(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
