# E2E Proxy for Fuel Sequencer

This directory contains the nginx proxy configuration for the fuel-sequencer e2e tests, providing HTTPS endpoints with CORS support for fuel-explorer integration.

## What it does

- Provides HTTPS proxy for Sequencer API (port {{PROXY_API_PORT}} → {{SEQUENCER_API_PORT}})
- Provides HTTPS proxy for Sequencer RPC (port {{PROXY_RPC_PORT}} → {{SEQUENCER_RPC_PORT}})
- Automatically generates self-signed SSL certificates during tests
- Enables CORS for cross-origin requests
- Fully configurable port mappings and host settings
- Optimized for local development workflow

## Automatic E2E Test Integration

The proxy is **automatically integrated** into the e2e test suite. When you run e2e tests, the proxy will:

1. Start automatically after the sequencer validators are running
2. Generate SSL certificates in the test data directory
3. Load configuration from `nginx.conf` template
4. Replace template variables with actual runtime values
5. Configure nginx with CORS support
6. Provide HTTPS endpoints using configured ports
7. Clean up automatically when tests complete

### Running E2E Tests with Proxy

```bash
# Run e2e tests with automatic HTTPS proxy
make test-e2e-with-proxy

# Or run any e2e test normally (proxy starts automatically)
make test-e2e-basic
```

### Using the Proxy in Tests

The test suite provides helper methods to get configuration values:

```go
func (s *YourTestSuite) TestWithProxy() {
    // Get the configured endpoints
    apiEndpoint, rpcEndpoint := s.GetProxyEndpoints()
    // apiEndpoint = "https://localhost:8443"
    // rpcEndpoint = "https://localhost:8658"
    
    // Use these endpoints for fuel-explorer or other HTTPS clients
}
```

## Configuration

The nginx configuration is stored in `nginx.conf` and automatically loaded by the e2e test suite. You can modify this file to adjust:

- CORS policies
- SSL settings
- Proxy headers
- Upstream destinations

## Directory Structure

```
e2e/proxy/
├── nginx.conf    # Nginx configuration (loaded by e2e tests)
├── .gitignore   # Excludes generated SSL certs & logs
└── README.md    # This file
```

**Note:** SSL certificates and logs are generated automatically in the e2e test data directory during test execution.

## Fuel Explorer Configuration

Update your `fuel-explorer/chains/mainnet/fuel-localhost.json` to use the HTTPS endpoints:

```json
{
    "api": [{"provider": "Localhost", "address": "https://localhost:{{PROXY_API_PORT}}"}],
    "rpc": [{"provider": "Localhost", "address": "https://localhost:{{PROXY_RPC_PORT}}"}]
}
```

## SSL Certificates

Self-signed certificates are automatically generated during e2e test execution and stored in the test data directory. These are suitable for local development but browsers will show a security warning.

To trust the certificate in your browser:
1. Navigate to `https://localhost:{{PROXY_API_PORT}}`
2. Click "Advanced" → "Proceed to localhost (unsafe)"
3. Or add the certificate to your system's trusted certificates

## Troubleshooting

**Ports already in use:**
- The e2e tests will fail if ports {{PROXY_API_PORT}} or {{PROXY_RPC_PORT}} are occupied
- Stop any services using these ports before running tests

**Connection refused:**
- Ensure the e2e test sequencer is running before the proxy starts
- Check test logs for any startup errors

**CORS issues:**
- The current config allows all origins (`*`) for development
- Modify `nginx.conf` to adjust CORS policies if needed 