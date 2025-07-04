package config

import (
	"fmt"
	"net/url"
	"time"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"github.com/fuel-infrastructure/fuel-sequencer/utils/credentials"
)

const (
	FlagSidecarEnabled        = "sidecar.enabled"
	FlagSidecarAddress        = "sidecar.address"
	FlagSidecarTimeout        = "sidecar.timeout"
	FlagSidecarPathToCertFile = "sidecar.path_to_cert_file"

	DefaultSidecarEnabled        = true
	DefaultSidecarAddress        = "localhost:8080"
	DefaultSidecarTimeout        = time.Second * 5
	DefaultSidecarPathToCertFile = ""
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().Bool(FlagSidecarEnabled, DefaultSidecarEnabled, "Sidecar querying enabled")
	startCmd.Flags().String(FlagSidecarAddress, DefaultSidecarAddress, "Sidecar server address")
	startCmd.Flags().Duration(FlagSidecarTimeout, DefaultSidecarTimeout, "Sidecar queries timeout")
	startCmd.Flags().String(
		FlagSidecarPathToCertFile,
		DefaultSidecarPathToCertFile,
		fmt.Sprintf(
			"Path to the sidecar server certificate for secure communication. "+
				"Required if the server uses TLS. For default credentials, set to '%s'.",
			credentials.UseDefaultTLS,
		),
	)
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

	// PathToCertFile defines the path to the certificate file of the sidecar server.
	PathToCertFile string `mapstructure:"pathToCertFile"`
}

func NewConfigFromAppOptions(opts servertypes.AppOptions) (cfg SidecarConfig, err error) {

	// determine if the Sidecar is enabled
	if v := opts.Get(FlagSidecarEnabled); v != nil {
		if cfg.Enabled, err = cast.ToBoolE(v); err != nil {
			return
		}
	}

	// get the Sidecar address
	if v := opts.Get(FlagSidecarAddress); v != nil {
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

	// get the path to the sidecar server certificate file
	if v := opts.Get(FlagSidecarPathToCertFile); v != nil {
		if cfg.PathToCertFile, err = cast.ToStringE(v); err != nil {
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

	// If the sidecar was enabled, the server address cannot be empty
	if cfg.Address == "" {
		return fmt.Errorf("sidecar address cannot be empty")
	}

	// Prepending a "//" allows addresses without a scheme.
	if _, err := url.Parse("//" + cfg.Address); err != nil {
		return fmt.Errorf("sidecar address must be valid: %w", err)
	}

	if cfg.Timeout <= 0 {
		return fmt.Errorf("sidecar client timeout must be greater than 0")
	}

	return nil
}
