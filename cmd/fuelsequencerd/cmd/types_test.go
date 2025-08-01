package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSidecarConfig_Validate(t *testing.T) {
	// Create temporary test files for certificate and key
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "test.cert")
	keyFile := filepath.Join(tmpDir, "test.key")

	// Create test files
	if err := os.WriteFile(certFile, []byte("test cert"), 0644); err != nil {
		t.Fatalf("Failed to create test cert file: %v", err)
	}
	if err := os.WriteFile(keyFile, []byte("test key"), 0644); err != nil {
		t.Fatalf("Failed to create test key file: %v", err)
	}

	tests := []struct {
		name    string
		config  sidecarConfig
		wantErr bool
	}{
		{
			name: "valid configuration with localhost",
			config: sidecarConfig{
				host:           "localhost",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: false,
		},
		{
			name: "valid configuration with IP",
			config: sidecarConfig{
				host:           "127.0.0.1",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: false,
		},
		{
			name: "valid configuration with broadcast IP",
			config: sidecarConfig{
				host:           "0.0.0.0",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: sidecarConfig{
				host:           "",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "invalid host",
			config: sidecarConfig{
				host:           "invalid@host",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "invalid port format",
			config: sidecarConfig{
				host:           "localhost",
				port:           "invalid",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "port out of range",
			config: sidecarConfig{
				host:           "localhost",
				port:           "70000",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "non-existent cert file",
			config: sidecarConfig{
				host:           "localhost",
				port:           "8080",
				development:    true,
				pathToCertFile: "nonexistent.cert",
				pathToKeyFile:  keyFile,
			},
			wantErr: true,
		},
		{
			name: "non-existent key file",
			config: sidecarConfig{
				host:           "localhost",
				port:           "8080",
				development:    true,
				pathToCertFile: certFile,
				pathToKeyFile:  "nonexistent.key",
			},
			wantErr: true,
		},
		{
			name: "empty cert and key files (optional)",
			config: sidecarConfig{
				host:           "localhost",
				port:           "8080",
				development:    true,
				pathToCertFile: "",
				pathToKeyFile:  "",
			},
			wantErr: false,
		},
		{
			name: "valid IPv4 address",
			config: sidecarConfig{
				host: "192.168.1.1",
				port: "8080",
			},
			wantErr: false,
		},
		{
			name: "valid IPv6 address",
			config: sidecarConfig{
				host: "2001:db8::1",
				port: "8080",
			},
			wantErr: false,
		},
		{
			name: "valid URL",
			config: sidecarConfig{
				host: "http://example.com",
				port: "8080",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("sidecarConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSequencerConfig_Validate(t *testing.T) {
	// Create a temporary certificate file for testing
	tmpDir := t.TempDir()
	certFile := filepath.Join(tmpDir, "test.cert")
	if err := os.WriteFile(certFile, []byte("test cert content"), 0644); err != nil {
		t.Fatalf("Failed to create test certificate file: %v", err)
	}

	tests := []struct {
		name    string
		config  sequencerConfig
		wantErr bool
	}{
		{
			name: "localhost config",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "localhost config without port",
			config: sequencerConfig{
				grpcUrl:           "localhost",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "valid config with IPv4 address",
			config: sequencerConfig{
				grpcUrl:           "127.0.0.1:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "valid config with IPv6 address",
			config: sequencerConfig{
				grpcUrl:           "[2001:db8::1]:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "malformed IPv4 address (allowed due to commented validation)",
			config: sequencerConfig{
				grpcUrl:           "256.256.256.256:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "incomplete IPv4 address (allowed due to commented validation)",
			config: sequencerConfig{
				grpcUrl:           "127.0.1:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "malformed IPv6 address (allowed due to commented validation)",
			config: sequencerConfig{
				grpcUrl:           "[2001:zzzz::1]:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "invalid config with unclosed IPv6 brackets",
			config: sequencerConfig{
				grpcUrl:           "[2001:db8::1:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "invalid config with URL without port",
			config: sequencerConfig{
				grpcUrl:           "https://example.com",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "valid config without cert file",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    "",
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "empty grpc url",
			config: sequencerConfig{
				grpcUrl:           "",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "invalid grpc url",
			config: sequencerConfig{
				grpcUrl:           "invalid:url:format",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "non-existent cert file",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    "non-existent.cert",
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
		{
			name: "empty bridge denom",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "",
			},
			wantErr: true,
		},
		{
			name: "uppercase bridge denom format",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "TEST-TOKEN",
			},
			wantErr: false,
		},
		{
			name: "bridge denom with invalid characters",
			config: sequencerConfig{
				grpcUrl:           "localhost:50051",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test@token", // special characters not allowed
			},
			wantErr: true,
		},
		{
			name: "valid https url",
			config: sequencerConfig{
				grpcUrl:           "https://localhost:443",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: false,
		},
		{
			name: "invalid url without port",
			config: sequencerConfig{
				grpcUrl:           "https://localhost",
				pathToCertFile:    certFile,
				unsafeBridgeDenom: "test-token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("sequencerConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
