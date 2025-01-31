package app

import (
	poolmodulev1 "cosmossdk.io/api/cosmos/protocolpool/module/v1"
	"cosmossdk.io/x/accounts"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/validate"
	bridgemodulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/bridge/module"
	_ "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/module" // import for side-effects
	bridgemoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"

	sequencingmodulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/sequencing/module"
	_ "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/module" // import for side-effects
	sequencingmoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/sequencing/types"

	reportsmodulev1 "github.com/fuel-infrastructure/fuel-sequencer/api/fuelsequencer/reports/module"
	_ "github.com/fuel-infrastructure/fuel-sequencer/x/reports/module" // import for side-effects
	reportsmoduletypes "github.com/fuel-infrastructure/fuel-sequencer/x/reports/types"

	accountsmodulev1 "cosmossdk.io/api/cosmos/accounts/module/v1"
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	authzmodulev1 "cosmossdk.io/api/cosmos/authz/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	consensusmodulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	distrmodulev1 "cosmossdk.io/api/cosmos/distribution/module/v1"
	evidencemodulev1 "cosmossdk.io/api/cosmos/evidence/module/v1"
	genutilmodulev1 "cosmossdk.io/api/cosmos/genutil/module/v1"
	govmodulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	mintmodulev1 "cosmossdk.io/api/cosmos/mint/module/v1"
	slashingmodulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	stakingmodulev1 "cosmossdk.io/api/cosmos/staking/module/v1"
	txconfigv1 "cosmossdk.io/api/cosmos/tx/config/v1"
	upgrademodulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	vestingmodulev1 "cosmossdk.io/api/cosmos/vesting/module/v1"
	"cosmossdk.io/depinject/appconfig"
	_ "cosmossdk.io/x/accounts" // import for side-effects
	"cosmossdk.io/x/authz"
	_ "cosmossdk.io/x/authz"        // import for side-effects
	_ "cosmossdk.io/x/authz/module" // import for side-effects
	_ "cosmossdk.io/x/bank"         // import for side-effects
	banktypes "cosmossdk.io/x/bank/types"
	_ "cosmossdk.io/x/consensus" // import for side-effects
	consensustypes "cosmossdk.io/x/consensus/types"
	_ "cosmossdk.io/x/distribution" // import for side-effects
	distrtypes "cosmossdk.io/x/distribution/types"
	_ "cosmossdk.io/x/epochs"   // import for side-effects
	_ "cosmossdk.io/x/evidence" // import for side-effects
	evidencetypes "cosmossdk.io/x/evidence/types"
	_ "cosmossdk.io/x/gov" // import for side-effects
	govtypes "cosmossdk.io/x/gov/types"
	_ "cosmossdk.io/x/mint" // import for side-effects
	minttypes "cosmossdk.io/x/mint/types"
	_ "cosmossdk.io/x/protocolpool" // import for side-effects
	pooltypes "cosmossdk.io/x/protocolpool/types"
	_ "cosmossdk.io/x/slashing" // import for side-effects
	slashingtypes "cosmossdk.io/x/slashing/types"
	_ "cosmossdk.io/x/staking" // import for side-effects
	stakingtypes "cosmossdk.io/x/staking/types"
	_ "cosmossdk.io/x/upgrade" // import for side-effects
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	_ "github.com/cosmos/cosmos-sdk/x/auth"           // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx"        // import for side-effects
	_ "github.com/cosmos/cosmos-sdk/x/auth/tx/config" // import for side-effects
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting" // import for side-effects
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	// this line is used by starport scaffolding # stargate/app/moduleImport
)

var (
	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	initGenesisOrder = []string{
		// cosmos-sdk/ibc modules
		consensustypes.ModuleName,
		accounts.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		pooltypes.ModuleName,
		// chain modules
		bridgemoduletypes.ModuleName,
		sequencingmoduletypes.ModuleName,
		reportsmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/initGenesis
	}

	exportGenesisOrder = []string{
		// cosmos-sdk/ibc modules
		consensustypes.ModuleName,
		accounts.ModuleName,
		authtypes.ModuleName,
		pooltypes.ModuleName, // Must be exported before bank
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		// chain modules
		bridgemoduletypes.ModuleName,
		sequencingmoduletypes.ModuleName,
		reportsmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/exportGenesis
	}

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	beginBlockers = []string{
		// cosmos sdk modules
		minttypes.ModuleName,
		distrtypes.ModuleName,
		pooltypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authz.ModuleName,
		genutiltypes.ModuleName,
		// chain modules
		bridgemoduletypes.ModuleName, // Must be after modules that can change supply, since it tracks supply changes.
		sequencingmoduletypes.ModuleName,
		reportsmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	}

	endBlockers = []string{
		// cosmos sdk modules
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		genutiltypes.ModuleName,
		pooltypes.ModuleName,
		// chain modules
		bridgemoduletypes.ModuleName,
		sequencingmoduletypes.ModuleName,
		reportsmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	}

	preBlockers = []string{
		upgradetypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/preBlockers
	}

	// module account permissions
	moduleAccPerms = []*authmodulev1.ModuleAccountPermission{
		{Account: authtypes.FeeCollectorName},
		{Account: distrtypes.ModuleName},
		{Account: pooltypes.ModuleName},
		{Account: pooltypes.StreamAccount},
		{Account: pooltypes.ProtocolPoolDistrAccount},
		{Account: minttypes.ModuleName, Permissions: []string{authtypes.Minter}},
		{Account: stakingtypes.BondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: stakingtypes.NotBondedPoolName, Permissions: []string{authtypes.Burner, stakingtypes.ModuleName}},
		{Account: govtypes.ModuleName, Permissions: []string{authtypes.Burner, authtypes.Minter}},
		{Account: bridgemoduletypes.ModuleName, Permissions: []string{authtypes.Burner, authtypes.Minter}},
		// this line is used by starport scaffolding # stargate/app/maccPerms
	}

	// blocked account addresses
	blockAccAddrs = []string{
		authtypes.FeeCollectorName,
		distrtypes.ModuleName,
		minttypes.ModuleName,
		stakingtypes.BondedPoolName,
		stakingtypes.NotBondedPoolName,
		bridgemoduletypes.ModuleName,
		// We allow the following module accounts to receive funds:
		// govtypes.ModuleName
		// pooltypes.ModuleName
	}

	// appConfig application configuration (used by depinject)
	appConfig = appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: runtime.ModuleName,
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName:       Name,
					PreBlockers:   preBlockers,
					BeginBlockers: beginBlockers,
					EndBlockers:   endBlockers,
					InitGenesis:   initGenesisOrder,
					OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
						{
							ModuleName: authtypes.ModuleName,
							KvStoreKey: "acc",
						},
						{
							ModuleName: accounts.ModuleName,
							KvStoreKey: accounts.StoreKey,
						},
					},
					// When ExportGenesis is not specified, the export genesis module order
					// is equal to the init genesis order
					ExportGenesis: exportGenesisOrder,
					// Uncomment if you want to set a custom migration order here.
					// OrderMigrations: nil,
					// SkipStoreKeys is an optional list of store keys to skip when constructing the
					// module's keeper. This is useful when a module does not have a store key.
					SkipStoreKeys: []string{
						authtxconfig.DepinjectModuleName,
						validate.ModuleName,
						genutiltypes.ModuleName,
						runtime.ModuleName,
						vestingtypes.ModuleName,
					},
				}),
			},
			{
				Name:   authtxconfig.DepinjectModuleName, // x/auth/tx/config depinject module (not app module), use to provide tx configuration
				Config: appconfig.WrapAny(&txconfigv1.Config{}),
			},
			{
				Name: authtypes.ModuleName,
				Config: appconfig.WrapAny(&authmodulev1.Module{
					Bech32Prefix:             AccountAddressPrefix,
					ModuleAccountPermissions: moduleAccPerms,
					// By default modules authority is the governance module. This is configurable with the following:
					// Authority: "group", // A custom module authority can be set using a module name
					// Authority: "cosmos1cwwv22j5ca08ggdv9c2uky355k908694z577tv", // or a specific address
				}),
			},
			{
				Name:   vestingtypes.ModuleName,
				Config: appconfig.WrapAny(&vestingmodulev1.Module{}),
			},
			{
				Name: banktypes.ModuleName,
				Config: appconfig.WrapAny(&bankmodulev1.Module{
					BlockedModuleAccountsOverride: blockAccAddrs,
				}),
			},
			{
				Name: stakingtypes.ModuleName,
				Config: appconfig.WrapAny(&stakingmodulev1.Module{
					// NOTE: specifying a prefix is only necessary when using bech32 addresses
					// If not specified, the auth Bech32Prefix appended with "valoper" and "valcons" is used by default
					Bech32PrefixValidator: AccountAddressPrefix + "valoper",
					Bech32PrefixConsensus: AccountAddressPrefix + "valcons",
				}),
			},
			{
				Name:   slashingtypes.ModuleName,
				Config: appconfig.WrapAny(&slashingmodulev1.Module{}),
			},
			{
				Name:   genutiltypes.ModuleName,
				Config: appconfig.WrapAny(&genutilmodulev1.Module{}),
			},
			{
				Name:   authz.ModuleName,
				Config: appconfig.WrapAny(&authzmodulev1.Module{}),
			},
			{
				Name:   upgradetypes.ModuleName,
				Config: appconfig.WrapAny(&upgrademodulev1.Module{}),
			},
			{
				Name:   distrtypes.ModuleName,
				Config: appconfig.WrapAny(&distrmodulev1.Module{}),
			},
			{
				Name:   evidencetypes.ModuleName,
				Config: appconfig.WrapAny(&evidencemodulev1.Module{}),
			},
			{
				Name:   minttypes.ModuleName,
				Config: appconfig.WrapAny(&mintmodulev1.Module{}),
			},
			{
				Name:   govtypes.ModuleName,
				Config: appconfig.WrapAny(&govmodulev1.Module{}),
			},
			{
				Name:   consensustypes.ModuleName,
				Config: appconfig.WrapAny(&consensusmodulev1.Module{}),
			},
			{
				Name:   bridgemoduletypes.ModuleName,
				Config: appconfig.WrapAny(&bridgemodulev1.Module{}),
			},
			{
				Name:   sequencingmoduletypes.ModuleName,
				Config: appconfig.WrapAny(&sequencingmodulev1.Module{}),
			},
			{
				Name:   reportsmoduletypes.ModuleName,
				Config: appconfig.WrapAny(&reportsmodulev1.Module{}),
			},
			{
				Name:   pooltypes.ModuleName,
				Config: appconfig.WrapAny(&poolmodulev1.Module{}),
			},
			{
				Name:   accounts.ModuleName,
				Config: appconfig.WrapAny(&accountsmodulev1.Module{}),
			},
			// this line is used by starport scaffolding # stargate/app/moduleConfig
		},
	})
)
