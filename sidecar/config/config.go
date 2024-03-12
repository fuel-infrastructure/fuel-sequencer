package config

import (
	"fmt"
	"net/url"
	"time"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

const (
	FlagSidecarEnabled = "sidecar.enabled"
	flagSidecarAddress = "sidecar.address"
	FlagSidecarTimeout = "sidecar.timeout"

	DefaultSidecarEnabled = false
	DefaultSidecarAddress = "localhost:8080"
	DefaultSidecarTimeout = time.Second * 5
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(FlagSidecarEnabled, DefaultSidecarEnabled, "Sidecar querying enabled")
	startCmd.Flags().String(flagSidecarAddress, DefaultSidecarAddress, "Sidecar client address")
	startCmd.Flags().Duration(FlagSidecarTimeout, DefaultSidecarTimeout, "Sidecar queries timeout")
}

// SidecarConfig contains the application side Sidecar configurations that must
// be set in the app.toml file.
type SidecarConfig struct {

	// Enabled dictates whether the Sidecar will be queried.
	Enabled bool `mapstructure:"enabled"`

	// Address defines the Sidecar server to listen to.
	Address string `mapstructure:"address"`

	// Timeout defines how long the client should wait for responses.
	Timeout time.Duration `mapstructure:"timeout"`
}

func NewConfigFromAppOptions(opts servertypes.AppOptions) (cfg SidecarConfig, err error) {

	// determine if the Sidecar is enabled
	if v := opts.Get(FlagSidecarEnabled); v != nil {
		if cfg.Enabled, err = cast.ToBoolE(v); err != nil {
			return
		}
	}

	// get the Sidecar address
	if v := opts.Get(flagSidecarAddress); v != nil {
		if cfg.Address, err = cast.ToStringE(v); err != nil {
			return
		}
	}

	// get the Sidecar client timeout
	if v := opts.Get(FlagSidecarTimeout); v != nil {
		if cfg.Timeout, err = cast.ToDurationE(v); err != nil {
			return
		}
	}

	return
}

// ValidateBasic performs basic validation of the Sidecar config.
func (cfg *SidecarConfig) ValidateBasic() error {
	if !cfg.Enabled {
		return nil
	}

	if _, err := url.ParseRequestURI(cfg.Address); err != nil {
		return fmt.Errorf("sidecar address must be valid: %w", err)
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("sidecar client timeout must be greater than 0")
	}

	return nil
}
