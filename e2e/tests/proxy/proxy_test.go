package proxy

import (
	"crypto/tls"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/fuel-infrastructure/fuel-sequencer/e2e/testsuite"
)

type ProxyTestSuite struct {
	testsuite.E2ETestSuite
}

func TestProxyTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyTestSuite))
}

func (s *ProxyTestSuite) TestProxyEndpoints() {
	s.T().Log("Testing HTTPS proxy endpoints...")

	// Get proxy endpoints
	apiEndpoint, rpcEndpoint := s.GetProxyEndpoints()

	s.Equal("https://localhost:8443", apiEndpoint)
	s.Equal("https://localhost:8658", rpcEndpoint)

	// Create HTTP client that accepts self-signed certificates
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	// Test API endpoint
	s.T().Log("Testing API endpoint:", apiEndpoint)
	resp, err := client.Get(apiEndpoint + "/cosmos/base/tendermint/v1beta1/node_info")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Test RPC endpoint
	s.T().Log("Testing RPC endpoint:", rpcEndpoint)
	resp, err = client.Get(rpcEndpoint + "/status")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	s.T().Log("✅ All proxy endpoints are working correctly!")
}

func (s *ProxyTestSuite) TestCORSHeaders() {
	s.T().Log("Testing CORS headers...")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	apiEndpoint, _ := s.GetProxyEndpoints()

	// Test CORS preflight request
	req, err := http.NewRequest("OPTIONS", apiEndpoint+"/cosmos/base/tendermint/v1beta1/node_info", nil)
	s.Require().NoError(err)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := client.Do(req)
	s.Require().NoError(err)
	defer resp.Body.Close()

	// Check CORS headers
	s.Equal("*", resp.Header.Get("Access-Control-Allow-Origin"))
	s.Contains(resp.Header.Get("Access-Control-Allow-Methods"), "GET")
	s.NotEmpty(resp.Header.Get("Access-Control-Max-Age"))

	s.T().Log("✅ CORS headers are configured correctly!")
}
