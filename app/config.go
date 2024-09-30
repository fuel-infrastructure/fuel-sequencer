package app

import (
	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	commitmentsconfig "github.com/fuel-infrastructure/fuel-sequencer/x/commitments/config"
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

// CustomAppConfig defines a configuration for a custom app.toml file.
// It essentially just adds the Sidecar and Commitments config to the typical Cosmos SDK Config.
type CustomAppConfig struct {
	serverconfig.Config `mapstructure:",squash"`
	SidecarConfig       sidecarconfig.SidecarConfig `mapstructure:"sidecar"`
	CommitmentsConfig   commitmentsconfig.Config    `mapstructure:"commitments"`
}

// DefaultCustomAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func DefaultCustomAppConfig() (string, interface{}) {

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
			Enabled:        sidecarconfig.DefaultSidecarEnabled,
			Address:        sidecarconfig.DefaultSidecarAddress,
			Timeout:        sidecarconfig.DefaultSidecarTimeout,
			PathToCertFile: sidecarconfig.DefaultSidecarPathToCertFile,
		},
		CommitmentsConfig: commitmentsconfig.Config{
			ApiEnabled: commitmentsconfig.DefaultCommitmentsApiEnabled,
		},
	}

	// Note: do not indent the below section, otherwise it will be indented in the config file as well.
	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[sidecar]
# This dictates whether the Sidecar will be queried.
enabled = {{ .SidecarConfig.Enabled }}
# This defines the Sidecar server to listen to.
address = "{{ .SidecarConfig.Address }}"
# This defines how long the client should wait for responses.
# This should be reasonably lower than the expected block time.
timeout = "{{ .SidecarConfig.Timeout }}"
# This defines the path to the certificate file for secure communication with the sidecar server.
# Should only be modified if the sidecar is to be configured with TLS.
# It can also be set to 'use_default_tls' for TLS with default credentials.
path_to_cert_file = "{{ .SidecarConfig.PathToCertFile }}"

[commitments]
# This dictates whether the commitments API (with bridge commitment queries) is enabled.
# Warning: The queries in this API are resource intensive and could be used to commit DOS.
#          If enabled, the queries should only be exposed to trusted clients.
api-enabled = {{ .CommitmentsConfig.ApiEnabled }}`

	return customAppTemplate, customAppConfig
}
