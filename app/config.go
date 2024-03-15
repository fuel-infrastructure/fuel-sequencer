package app

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
)

func InitSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := AccountAddressPrefix + "pub"
	validatorAddressPrefix := AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := AccountAddressPrefix + "valconspub"

	// Set and seal config
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount(AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)
	config.Seal()
}

// InitCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func InitCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

// InitAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func InitAppConfig() (string, interface{}) {
	// CustomAppConfig defines a custom app configuration for a custom app.toml file.
	// It essentially just adds SidecarConfig to the typical Cosmos SDK Config.
	type CustomAppConfig struct {
		serverconfig.Config `mapstructure:",squash"`
		SidecarConfig       sidecarconfig.SidecarConfig `mapstructure:"sidecar"`
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In tests, we set the min gas prices to 0.
	// srvCfg.MinGasPrices = "0stake"
	// srvCfg.BaseConfig.IAVLDisableFastNode = true // disable fastnode by default

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		SidecarConfig: sidecarconfig.SidecarConfig{
			Enabled: sidecarconfig.DefaultSidecarEnabled,
			Address: sidecarconfig.DefaultSidecarAddress,
			Timeout: sidecarconfig.DefaultSidecarTimeout,
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
	[sidecar]
	# This dictates whether the Sidecar will be queried.
	enabled = true
	# This defines the Sidecar server to listen to.
	address = "http://localhost:8080"
	# This defines how long the client should wait for responses.
	timeout = "5s"`

	return customAppTemplate, customAppConfig
}
