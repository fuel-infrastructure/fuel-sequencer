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
}

// Config contains the application side Blob configurations that must be set in the app.toml file.
type Config struct {
	// BlobhubAddress defines the address of the blobhub server for blob synchronization.
	BlobhubAddress string `mapstructure:"blobhub-address"`
	// BlobpoolRedisAddress defines the address of the Redis server for blobpool storage.
	BlobpoolRedisAddress string `mapstructure:"blobpool-redis-address"`
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
	return nil
}
