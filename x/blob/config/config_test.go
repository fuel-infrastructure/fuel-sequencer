package config

import (
	"testing"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	// Test default configuration
	cfg := Config{
		BlobhubAddress:     DefaultBlobhubAddress,
		BlobpoolSqlitePath: DefaultBlobpoolSqlitePath,
	}
	err := cfg.ValidateBasic()
	require.NoError(t, err)

	// Test with custom addresses
	cfg.BlobhubAddress = "custom-blobhub:8080"
	cfg.BlobpoolSqlitePath = "./custom/blobpool.db"
	err = cfg.ValidateBasic()
	require.NoError(t, err)

	// Test with empty blobhub address (should fail)
	cfg.BlobhubAddress = ""
	err = cfg.ValidateBasic()
	require.Error(t, err)

	// Test with empty blobpool sqlite path (should fail)
	cfg.BlobhubAddress = "custom-blobhub:8080"
	cfg.BlobpoolSqlitePath = ""
	err = cfg.ValidateBasic()
	require.Error(t, err)
}

func TestNewConfigFromAppOptions(t *testing.T) {
	// Create a viper instance with test values
	v := viper.New()
	v.Set(FlagBlobhubAddress, "test-blobhub:9090")
	v.Set(FlagBlobpoolSqlitePath, "./test/blobpool.db")

	// Convert to AppOptions
	appOpts := servertypes.AppOptions(v)

	cfg, err := NewConfigFromAppOptions(appOpts)
	require.NoError(t, err)
	require.Equal(t, "test-blobhub:9090", cfg.BlobhubAddress)
	require.Equal(t, "./test/blobpool.db", cfg.BlobpoolSqlitePath)

	// Test with default values
	v = viper.New()
	appOpts = servertypes.AppOptions(v)

	cfg, err = NewConfigFromAppOptions(appOpts)
	require.NoError(t, err)
	require.Equal(t, DefaultBlobhubAddress, cfg.BlobhubAddress)
	require.Equal(t, DefaultBlobpoolSqlitePath, cfg.BlobpoolSqlitePath)
}
