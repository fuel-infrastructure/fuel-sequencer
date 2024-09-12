package client_test

import (
	"context"
	"testing"
	"time"

	"cosmossdk.io/log"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil/servers"
	apptesttypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/connectivity"
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

func TestStartAndStop(t *testing.T) {
	testCases := []struct {
		name           string
		pathToCertFile string
		pathToKeyFile  string
	}{
		{
			name:           "Start establishes connection and stop closes it (with TLS)",
			pathToCertFile: "../testutil/certificates/server-cert.pem",
			pathToKeyFile:  "../testutil/certificates/server-key.pem",
		},
		{
			name: "Start establishes connection and stop closes it (without TLS)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a TestSidecarServer for the AppSidecarClient to connect with. We use localhost:0 so that the
			// OS picks up an available port.
			testSidecarServer := servers.MustMakeTestSidecarServer("localhost:0", tc.pathToCertFile, tc.pathToKeyFile)
			testSidecarServer.Start()
			defer testSidecarServer.Stop()

			// Create the SidecarClient configuration
			timeout := 10 * time.Second
			address := testSidecarServer.GetAddress()
			configuration := sidecarconfig.SidecarConfig{
				Enabled:        true,
				Address:        address,
				Timeout:        timeout,
				PathToCertFile: tc.pathToCertFile,
			}

			// Create AppSidecarClient from configuration
			appSidecarClient, err := sidecarclient.NewClientFromConfig(configuration, testutil.ValidLogger)
			require.NoError(t, err, "expected no error when creating client")

			// Create context
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Start AppSidecarClient. This will error if the client could not establish a connection with the Server.
			err = appSidecarClient.Start(ctx)
			require.NoError(t, err, "expected no error when starting client")

			// Check client state
			grpcClient, ok := appSidecarClient.(*sidecarclient.GRPCClient)
			require.True(t, ok)
			require.NotNil(t, grpcClient.Client())

			// Wait 5 seconds and check that the connection was established
			time.Sleep(time.Second * 5)
			require.Equal(t, connectivity.Ready, grpcClient.Conn().GetState())

			// Stop AppSidecarClient and confirm that the connection was shut down.
			err = appSidecarClient.Stop()
			require.NoError(t, err)
			require.Equal(t, connectivity.Shutdown, grpcClient.Conn().GetState())
		})
	}
}

func TestStop_DoesNotErrorIfClientDidNotStart(t *testing.T) {
	// Create AppSidecarClient from configuration
	appSidecarClient, err := sidecarclient.NewClientFromConfig(testutil.ValidSidecarConfig, testutil.ValidLogger)
	require.NoError(t, err, "expected no error when creating client")

	// Stop AppSidecarClient and confirm that the client did not error
	err = appSidecarClient.Stop()
	require.NoError(t, err)

	// Make sure that the connection is still nil
	grpcClient, ok := appSidecarClient.(*sidecarclient.GRPCClient)
	require.True(t, ok)
	require.Nil(t, grpcClient.Conn())
}

func TestGetBlockEvents(t *testing.T) {
	testCases := []struct {
		name           string
		pathToCertFile string
		pathToKeyFile  string
		startClient    bool
		expErrMsg      string
	}{
		{
			name:           "Returns block events if query successful (with TLS)",
			pathToCertFile: "../testutil/certificates/server-cert.pem",
			pathToKeyFile:  "../testutil/certificates/server-key.pem",
			startClient:    true,
		},
		{
			name:        "Returns block events if query successful (without TLS)",
			startClient: true,
		},
		{
			name:           "errors if client not started (with TLS)",
			pathToCertFile: "../testutil/certificates/server-cert.pem",
			pathToKeyFile:  "../testutil/certificates/server-key.pem",
			startClient:    false,
			expErrMsg:      "sidecar client not started",
		},
		{
			name:        "errors if client not started (without TLS)",
			startClient: false,
			expErrMsg:   "sidecar client not started",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a TestSidecarServer for the AppSidecarClient to connect with. We use localhost:0 so that the
			// OS picks up an available port.
			testSidecarServer := servers.MustMakeTestSidecarServer("localhost:0", tc.pathToCertFile, tc.pathToKeyFile)
			testSidecarServer.Start()
			defer testSidecarServer.Stop()

			// Create the SidecarClient configuration
			timeout := 10 * time.Second
			address := testSidecarServer.GetAddress()
			configuration := sidecarconfig.SidecarConfig{
				Enabled:        true,
				Address:        address,
				Timeout:        timeout,
				PathToCertFile: tc.pathToCertFile,
			}

			// Create AppSidecarClient from configuration
			appSidecarClient, err := sidecarclient.NewClientFromConfig(configuration, testutil.ValidLogger)
			require.NoError(t, err, "expected no error when creating client")

			// Create context
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			// Start AppSidecarClient if enabled in test
			if tc.startClient {
				err = appSidecarClient.Start(ctx)
				require.NoError(t, err, "expected no error when starting client")
			}

			// Mock the events returned by GetBlockEvents
			testSidecarServer.SetEvents(apptesttypes.TestEvents)

			// Call GetBlockEvents
			resp, err := appSidecarClient.GetBlockEvents(ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"})

			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			// Make sure that the correct events are returned
			require.Equal(t, apptesttypes.TestEvents, resp.Events)

			// Stop AppSidecarClient if startup was enabled in test
			if tc.startClient {
				err = appSidecarClient.Stop()
				require.NoError(t, err)
			}
		})
	}
}
