package client_test

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/stretchr/testify/require"
)

// testGRPCClientFields is a structure defining the fields of GRPCClient that are of interest for testing purposes.
type testGRPCClientFields struct {
	logger         log.Logger
	addr           string
	timeout        time.Duration
	pathToCertFile string
}

func TestNewClientFromConfig(t *testing.T) {

	testCases := []struct {
		name                string
		config              sidecarconfig.SidecarConfig
		logger              log.Logger
		expErrMsg           string
		expGRPCClientFields testGRPCClientFields
	}{
		{
			name:   "Returns expected client if config is valid",
			config: testutil.ValidSidecarConfig,
			logger: testutil.ValidLogger,
			expGRPCClientFields: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           testutil.ValidSidecarConfig.Address,
				timeout:        testutil.ValidSidecarConfig.Timeout,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
		},
		{
			name:      "Returns error if invalid config",
			config:    testutil.InvalidSidecarConfig,
			logger:    testutil.ValidLogger,
			expErrMsg: "sidecar client timeout must be greater than 0",
		},
		{
			name:      "Returns error if logger is nil",
			config:    testutil.ValidSidecarConfig,
			logger:    nil,
			expErrMsg: "logger cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appSidecarClient, err := sidecarclient.NewClientFromConfig(tc.config, tc.logger)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			grpcClient, ok := appSidecarClient.(*sidecarclient.GRPCClient)
			require.True(t, ok)
			require.Equal(t, tc.expGRPCClientFields.addr, grpcClient.Addr())
			require.Equal(t, tc.expGRPCClientFields.timeout, grpcClient.Timeout())
			require.Equal(t, tc.expGRPCClientFields.logger, grpcClient.Logger())
			require.Equal(t, tc.expGRPCClientFields.pathToCertFile, grpcClient.PathToCertFile())
		})
	}
}

func TestNewClientFromConfig_NoOpClientIfSidecarConfigDisabled(t *testing.T) {
	client, err := sidecarclient.NewClientFromConfig(testutil.DisabledSidecarConfig, testutil.ValidLogger)
	require.NoError(t, err)
	require.Equal(t, &sidecarclient.NoOpClient{}, client)
}
