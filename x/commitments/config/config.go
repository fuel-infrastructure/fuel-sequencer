package config

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

const (
	FlagCommitmentsApiEnabled = "commitments.api-enabled"

	DefaultCommitmentsApiEnabled = false
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(
		FlagCommitmentsApiEnabled,
		DefaultCommitmentsApiEnabled,
		"Commitments API enabled",
	)
}

// Config contains the application side Commitments configurations that must be set in the app.toml file.
type Config struct {

	// ApiEnabled dictates whether the Commitments API is enabled. This is set to false by default to prevent the
	// accidental exposure of Bridge Commitment queries, which could be exploited to perform a DoS attack if not
	// properly secured.
	ApiEnabled bool `mapstructure:"api-enabled"`
}

func NewConfigFromAppOptions(opts servertypes.AppOptions) (cfg Config, err error) {

	// determine if the API is enabled
	if v := opts.Get(FlagCommitmentsApiEnabled); v != nil {
		if cfg.ApiEnabled, err = cast.ToBoolE(v); err != nil {
			return
		}
	}

	return
}

// ValidateBasic performs basic validation of the Commitments config.
func (cfg *Config) ValidateBasic() error {

	// Nothing to verify here for now.

	return nil
}
