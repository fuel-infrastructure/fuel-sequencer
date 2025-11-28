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
		BlobhubAddress:       DefaultBlobhubAddress,
		BlobpoolRedisAddress: DefaultBlobpoolRedisAddress,
	}
	err := cfg.ValidateBasic()
	require.NoError(t, err)

	// Test with custom addresses
	cfg.BlobhubAddress = "custom-blobhub:8080"
	cfg.BlobpoolRedisAddress = "custom-redis:6379"
	err = cfg.ValidateBasic()
	require.NoError(t, err)

	// Test with empty blobhub address (should fail)
	cfg.BlobhubAddress = ""
	err = cfg.ValidateBasic()
	require.Error(t, err)

	// Test with empty blobpool redis address (should fail)
	cfg.BlobhubAddress = "custom-blobhub:8080"
	cfg.BlobpoolRedisAddress = ""
	err = cfg.ValidateBasic()
	require.Error(t, err)
}

func TestNewConfigFromAppOptions(t *testing.T) {
	// Create a viper instance with test values
	v := viper.New()
	v.Set(FlagBlobhubAddress, "test-blobhub:9090")
	v.Set(FlagBlobpoolRedisAddress, "test-redis:6379")

	// Convert to AppOptions
	appOpts := servertypes.AppOptions(v)

	cfg, err := NewConfigFromAppOptions(appOpts)
	require.NoError(t, err)
	require.Equal(t, "test-blobhub:9090", cfg.BlobhubAddress)
	require.Equal(t, "test-redis:6379", cfg.BlobpoolRedisAddress)

	// Test with default values
	v = viper.New()
	appOpts = servertypes.AppOptions(v)

	cfg, err = NewConfigFromAppOptions(appOpts)
	require.NoError(t, err)
	require.Equal(t, DefaultBlobhubAddress, cfg.BlobhubAddress)
	require.Equal(t, DefaultBlobpoolRedisAddress, cfg.BlobpoolRedisAddress)
}
