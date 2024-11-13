package testutil

import (
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	"cosmossdk.io/core/appconfig"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	_ "github.com/cosmos/cosmos-sdk/x/auth"                     // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config"           // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/bank"                     // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/consensus"                // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/distribution"             // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/genutil"                  // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/mint"                     // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/params"                   // import as blank for app wiring
	_ "github.com/cosmos/cosmos-sdk/x/slashing"                 // import as blank for app wiring
	_ "github.com/fuel-infrastructure/fuel-sequencer/x/staking" // import as blank for app wiring
)

// MintModule defines a custom configuration for the mint module. We cannot use the Cosmos SDK configurator because it
// points to the Cosmos SDKs staking module.
func MintModule() configurator.ModuleOption {
	return func(config *configurator.Config) {
		config.ModuleConfigs["mint"] = &appv1alpha1.ModuleConfig{
			Name:   "mint",
			Config: appconfig.WrapAny(&mintmodulev1.Module{}),
			GolangBindings: []*appv1alpha1.GolangBinding{
				{
					InterfaceType:  "github.com/cosmos/cosmos-sdk/x/mint/types/types.StakingKeeper",
					Implementation: "github.com/fuel-infrastructure/fuel-sequencer/x/staking/keeper/*keeper.Keeper",
				},
			},
		}
	}
}

var AppConfig = configurator.NewAppConfig(
	configurator.AuthModule(),
	configurator.BankModule(),
	configurator.StakingModule(),
	configurator.TxModule(),
	configurator.ConsensusModule(),
	configurator.ParamsModule(),
	configurator.GenutilModule(),
	MintModule(),
	configurator.DistributionModule(),
)
