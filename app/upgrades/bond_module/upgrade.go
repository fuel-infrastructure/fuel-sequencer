package bond_module

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const UpgradeName = "bond-module"

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// The bond module is already registered in the module manager through app wiring
		// We just need to set its initial version in the version map
		fromVM["bond"] = 1 // Initial version of the bond module

		// Run any migrations
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
