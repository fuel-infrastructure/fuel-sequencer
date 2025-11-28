package config

import (
	"fmt"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

const (
	FlagBlobhubAddress    = "blob.blobhub-address"
	DefaultBlobhubAddress = "localhost:31035"

	FlagBlobpoolRedisAddress    = "blob.blobpool-redis-address"
	DefaultBlobpoolRedisAddress = "localhost:6380"

	FlagBlobpoolServerEnabled    = "blob.blobpool-server-enabled"
	DefaultBlobpoolServerEnabled = false

	FlagBlobpoolServerAddress    = "blob.blobpool-server-address"
	DefaultBlobpoolServerAddress = "localhost:21025"
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(
		FlagBlobhubAddress,
		DefaultBlobhubAddress,
		"Blobhub server address for blob synchronization",
	)
	startCmd.Flags().String(
		FlagBlobpoolRedisAddress,
		DefaultBlobpoolRedisAddress,
		"Redis server address for blobpool storage",
	)
	startCmd.Flags().Bool(
		FlagBlobpoolServerEnabled,
		DefaultBlobpoolServerEnabled,
		"Enable blobpool server for querying and profiling",
	)
	startCmd.Flags().String(
		FlagBlobpoolServerAddress,
		DefaultBlobpoolServerAddress,
		"Blobpool server address for querying and profiling",
	)
}

// Config contains the application side Blob configurations that must be set in the app.toml file.
type Config struct {
	// BlobhubAddress defines the address of the blobhub server for blob synchronization.
	BlobhubAddress string `mapstructure:"blobhub-address"`
	// BlobpoolRedisAddress defines the address of the Redis server for blobpool storage.
	BlobpoolRedisAddress string `mapstructure:"blobpool-redis-address"`
	// BlobpoolServerEnabled defines whether the blobpool server should be enabled for querying and profiling.
	BlobpoolServerEnabled bool `mapstructure:"blobpool-server-enabled"`
	// BlobpoolServerAddress defines the address of the blobpool server for querying and profiling.
	BlobpoolServerAddress string `mapstructure:"blobpool-server-address"`
}

func NewConfigFromAppOptions(opts servertypes.AppOptions) (cfg Config, err error) {
	// determine the blobhub address
	if v := opts.Get(FlagBlobhubAddress); v != nil {
		if cfg.BlobhubAddress, err = cast.ToStringE(v); err != nil {
			return
		}
	} else {
		cfg.BlobhubAddress = DefaultBlobhubAddress
	}

	// determine the blobpool redis address
	if v := opts.Get(FlagBlobpoolRedisAddress); v != nil {
		if cfg.BlobpoolRedisAddress, err = cast.ToStringE(v); err != nil {
			return
		}
	} else {
		cfg.BlobpoolRedisAddress = DefaultBlobpoolRedisAddress
	}

	// determine the blobpool server enabled setting
	if v := opts.Get(FlagBlobpoolServerEnabled); v != nil {
		if cfg.BlobpoolServerEnabled, err = cast.ToBoolE(v); err != nil {
			return
		}
	} else {
		cfg.BlobpoolServerEnabled = DefaultBlobpoolServerEnabled
	}

	// determine the blobpool server address
	if v := opts.Get(FlagBlobpoolServerAddress); v != nil {
		if cfg.BlobpoolServerAddress, err = cast.ToStringE(v); err != nil {
			return
		}
	} else {
		cfg.BlobpoolServerAddress = DefaultBlobpoolServerAddress
	}

	return
}

// ValidateBasic performs basic validation of the Blob config.
func (cfg *Config) ValidateBasic() error {
	if cfg.BlobhubAddress == "" {
		return fmt.Errorf("blobhub address cannot be empty")
	}
	if cfg.BlobpoolRedisAddress == "" {
		return fmt.Errorf("blobpool redis address cannot be empty")
	}
	if cfg.BlobpoolServerEnabled && cfg.BlobpoolServerAddress == "" {
		return fmt.Errorf("blobpool server address cannot be empty when server is enabled")
	}
	return nil
}
