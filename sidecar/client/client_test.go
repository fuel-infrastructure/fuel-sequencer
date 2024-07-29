package client_test

import (
	"context"
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

func TestNewClient(t *testing.T) {
	validGRPCClientFields := testGRPCClientFields{
		logger:         testutil.ValidLogger,
		addr:           testutil.ValidSidecarConfig.Address,
		timeout:        testutil.ValidSidecarConfig.Timeout,
		pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
	}

	testCases := []struct {
		name                string
		fnInput             testGRPCClientFields
		expErrMsg           string
		expGRPCClientFields testGRPCClientFields
	}{
		{
			name:                "Returns expected client if function input is valid",
			fnInput:             validGRPCClientFields,
			expGRPCClientFields: validGRPCClientFields,
		},
		{
			name: "Returns error if logger is nil",
			fnInput: testGRPCClientFields{
				logger:         nil,
				addr:           testutil.ValidSidecarConfig.Address,
				timeout:        testutil.ValidSidecarConfig.Timeout,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "logger cannot be nil",
		},
		{
			name: "Returns error if address is invalid (space between host and port)",
			fnInput: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           "1.1.1.1 :80",
				timeout:        testutil.ValidSidecarConfig.Timeout,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "invalid Sidecar address",
		},
		{
			name: "Returns error if address is invalid (invalid characters and structure)",
			fnInput: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           "!@#$%^&*()",
				timeout:        testutil.ValidSidecarConfig.Timeout,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "invalid Sidecar address",
		},
		{
			name: "Returns error if address is invalid (empty address)",
			fnInput: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           "",
				timeout:        testutil.ValidSidecarConfig.Timeout,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "sidecar address cannot be empty",
		},
		{
			name: "Returns error if timeout is invalid (zero)",
			fnInput: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           testutil.ValidSidecarConfig.Address,
				timeout:        0,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "timeout must be positive",
		},
		{
			name: "Returns error if timeout is invalid (negative)",
			fnInput: testGRPCClientFields{
				logger:         testutil.ValidLogger,
				addr:           testutil.ValidSidecarConfig.Address,
				timeout:        -1,
				pathToCertFile: testutil.ValidSidecarConfig.PathToCertFile,
			},
			expErrMsg: "timeout must be positive",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			appSidecarClient, err := sidecarclient.NewClient(
				tc.fnInput.logger, tc.fnInput.addr, tc.fnInput.timeout, tc.fnInput.pathToCertFile,
			)
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

func TestStart(t *testing.T) {
	// Create a TestSidecarServer. Let the OS pick the port.
	testSidecarServer := testutil.MustMakeTestSidecarServer("localhost:0")
	testSidecarServer.Start()
	defer testSidecarServer.Stop()

	// Create the SidecarClient configuration
	timeout := 5 * time.Second
	address := testSidecarServer.GetAddress()
	configuration := sidecarconfig.SidecarConfig{
		Enabled:        true,
		Address:        address,
		Timeout:        timeout,
		PathToCertFile: "",
	}

	// Create AppSidecarClient from configuration
	appSidecarClient, err := sidecarclient.NewClientFromConfig(configuration, log.NewNopLogger())
	require.NoError(t, err, "expected no error when creating client")

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start client. This should create an insecure connection with the Sidecar server
	err = appSidecarClient.Start(ctx)
	require.NoError(t, err, "expected no error when starting client")

	// Check client state
	grpcClient, ok := appSidecarClient.(*sidecarclient.GRPCClient)
	require.True(t, ok)
	require.NotNil(t, grpcClient.Conn())
	require.NotNil(t, grpcClient.Client())
	require.NotNil(t, grpcClient.Mutex())
}

// TODO: Test stop and QueryBlockEvents
