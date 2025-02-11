package vesting_accounts_staking

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const UpgradeName = "vesting-accounts-staking"

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		// returns a VersionMap with the updated module ConsensusVersions
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
