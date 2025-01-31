package reports

import (
	"context"
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx context.Context, k keeper.Keeper, genState types.GenesisState) error {

	// Set all the slash reports.
	// We're assuming that the slash report heights are unique because they are validated in GenesisState.Validate()
	for _, slashReport := range genState.SlashReportList {

		// Panic if slash report fails validation
		if err := slashReport.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid slash report: %w", err)
		}

		k.SetSlashReport(ctx, slashReport)
	}
	// this line is used by starport scaffolding # genesis/module/init

	if err := k.SetParams(ctx, genState.Params); err != nil {
		return fmt.Errorf("error while setting params: %w", err)
	}

	return nil
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx context.Context, k keeper.Keeper) (*types.GenesisState, error) {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.SlashReportList = k.GetAllSlashReport(ctx)

	// this line is used by starport scaffolding # genesis/module/export

	return genesis, nil
}
