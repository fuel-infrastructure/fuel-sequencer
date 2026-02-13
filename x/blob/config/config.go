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

	FlagBlobpoolSqlitePath    = "blob.blobpool-sqlite-path"
	DefaultBlobpoolSqlitePath = "./data/blobpool.db"

	FlagBlobpoolServerEnabled    = "blob.blobpool-server-enabled"
	DefaultBlobpoolServerEnabled = false

	FlagBlobpoolServerAddress    = "blob.blobpool-server-address"
	DefaultBlobpoolServerAddress = "localhost:21025"

	FlagChunkMode                   = "blob.chunk-mode"
	DefaultChunkMode                = false
	FlagChunkValidatorIndex         = "blob.chunk-validator-index"
	DefaultChunkValidatorIndex  int = 0
)

func AddStartCmdFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(
		FlagBlobhubAddress,
		DefaultBlobhubAddress,
		"Blobhub server address for blob synchronization",
	)
	startCmd.Flags().String(
		FlagBlobpoolSqlitePath,
		DefaultBlobpoolSqlitePath,
		"SQLite database path for blobpool storage",
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
	startCmd.Flags().Bool(
		FlagChunkMode,
		DefaultChunkMode,
		"Enable chunk-mode attestation (EigenDA-inspired erasure coding)",
	)
	startCmd.Flags().Int(
		FlagChunkValidatorIndex,
		DefaultChunkValidatorIndex,
		"Validator chunk index for chunk-mode attestation (which chunk to verify)",
	)
}

// Config contains the application side Blob configurations that must be set in the app.toml file.
type Config struct {
	// BlobhubAddress defines the address of the blobhub server for blob synchronization.
	BlobhubAddress string `mapstructure:"blobhub-address"`
	// BlobpoolSqlitePath defines the path to the SQLite database for blobpool storage.
	BlobpoolSqlitePath string `mapstructure:"blobpool-sqlite-path"`
	// BlobpoolServerEnabled defines whether the blobpool server should be enabled for querying and profiling.
	BlobpoolServerEnabled bool `mapstructure:"blobpool-server-enabled"`
	// BlobpoolServerAddress defines the address of the blobpool server for querying and profiling.
	BlobpoolServerAddress string `mapstructure:"blobpool-server-address"`
	// ChunkMode enables chunk-mode attestation (erasure coding + Merkle proof verification).
	ChunkMode bool `mapstructure:"chunk-mode"`
	// ChunkValidatorIndex specifies which chunk this validator attests in chunk mode.
	ChunkValidatorIndex int `mapstructure:"chunk-validator-index"`
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

	// determine the blobpool sqlite path
	if v := opts.Get(FlagBlobpoolSqlitePath); v != nil {
		if cfg.BlobpoolSqlitePath, err = cast.ToStringE(v); err != nil {
			return
		}
	} else {
		cfg.BlobpoolSqlitePath = DefaultBlobpoolSqlitePath
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

	// determine chunk mode
	if v := opts.Get(FlagChunkMode); v != nil {
		if cfg.ChunkMode, err = cast.ToBoolE(v); err != nil {
			return
		}
	} else {
		cfg.ChunkMode = DefaultChunkMode
	}

	// determine chunk validator index
	if v := opts.Get(FlagChunkValidatorIndex); v != nil {
		if cfg.ChunkValidatorIndex, err = cast.ToIntE(v); err != nil {
			return
		}
	} else {
		cfg.ChunkValidatorIndex = DefaultChunkValidatorIndex
	}

	return
}

// ValidateBasic performs basic validation of the Blob config.
func (cfg *Config) ValidateBasic() error {
	if cfg.BlobhubAddress == "" {
		return fmt.Errorf("blobhub address cannot be empty")
	}
	if cfg.BlobpoolSqlitePath == "" {
		return fmt.Errorf("blobpool sqlite path cannot be empty")
	}
	if cfg.BlobpoolServerEnabled && cfg.BlobpoolServerAddress == "" {
		return fmt.Errorf("blobpool server address cannot be empty when server is enabled")
	}
	if cfg.ChunkMode && cfg.ChunkValidatorIndex < 0 {
		return fmt.Errorf("chunk validator index cannot be negative")
	}
	return nil
}
