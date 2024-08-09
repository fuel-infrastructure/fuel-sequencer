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

	// Enabled dictates whether the Commitments API is enabled. This defaults to false, and was put in place to prevent
	// accidental exposure of the Bridge Commitment queries which can be used to perform a DOS attack if not protected.
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
