package config

import (
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

const (
	FlagCommitmentsApiEnabled    = "commitments.api-enabled"
	FlagCommitmentsMaxQueryRange = "commitments.max-query-range"

	DefaultCommitmentsApiEnabled    = false
	DefaultCommitmentsMaxQueryRange = 4096
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(
		FlagCommitmentsApiEnabled,
		DefaultCommitmentsApiEnabled,
		"Commitments API enabled",
	)
	startCmd.Flags().Uint64(
		FlagCommitmentsMaxQueryRange,
		DefaultCommitmentsMaxQueryRange,
		"Commitments API max query range",
	)
}

// Config contains the application side Commitments configurations that must be set in the app.toml file.
type Config struct {

	// ApiEnabled dictates whether the Commitments API is enabled. This is set to false by default to prevent the
	// accidental exposure of Bridge Commitment queries, which could be exploited to perform a DoS attack if not
	// properly secured.
	ApiEnabled bool `mapstructure:"api-enabled"`

	// MaxQueryRange determines the maximum difference between the start block and end block when querying for bridge
	// commitments and bridge commitment inclusion proofs. It allows the node operator to limit the query size.
	// Otherwise, the entire block range of the chain could be queried (i.e. from block 1 to the latest block).
	MaxQueryRange uint64 `mapstructure:"max-query-range"`
}

func NewConfigFromAppOptions(opts servertypes.AppOptions) (cfg Config, err error) {

	// determine if the API is enabled
	if v := opts.Get(FlagCommitmentsApiEnabled); v != nil {
		if cfg.ApiEnabled, err = cast.ToBoolE(v); err != nil {
			return
		}
	}

	// determine the max query range
	if v := opts.Get(FlagCommitmentsMaxQueryRange); v != nil {
		if cfg.MaxQueryRange, err = cast.ToUint64E(v); err != nil {
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
