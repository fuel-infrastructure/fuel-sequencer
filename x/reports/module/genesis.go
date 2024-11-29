package reports

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {

	// Set all the slash reports
	for _, slashReport := range genState.SlashReportList {

		// Panic if slash report fails validation
		if err := slashReport.ValidateBasic(); err != nil {
			panic(err)
		}

		k.SetSlashReport(ctx, slashReport)
	}
	// this line is used by starport scaffolding # genesis/module/init

	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(fmt.Sprintf("error when setting params: %x", err))
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.SlashReportList = k.GetAllSlashReport(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
