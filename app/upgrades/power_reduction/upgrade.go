package power_reduction

import (
	"context"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

const UpgradeName = "increase-power-reduction"

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
) upgradetypes.UpgradeHandler {
	return func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {

		// There is no need to apply this change here since it's already applied in the app.go init function.
		//sdk.DefaultPowerReduction = sdkmath.NewIntFromUint64(1000000000)

		// returns a VersionMap with the updated module ConsensusVersions
		return mm.RunMigrations(ctx, configurator, fromVM)
	}
}
