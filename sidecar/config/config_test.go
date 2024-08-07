package config_test

import (
	"testing"
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
)

func TestSidecarConfig_ValidateBasic(t *testing.T) {

	type fields struct {
		Enabled        bool
		Address        string
		Timeout        time.Duration
		PathToCertFile string
	}

	invalidTimeoutValue := time.Duration(0) // zero
	validTimeoutValue := time.Duration(1)   // non-zero
	validPathToCertFile := ""
	validFieldsWithAddress := func(address string) fields {
		return fields{
			Enabled:        true,
			Address:        address,
			Timeout:        validTimeoutValue,
			PathToCertFile: validPathToCertFile,
		}
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "disabled sidecar => valid config regardless of other values",
			fields: fields{
				Enabled:        false,
				Address:        "not a valid address",
				Timeout:        invalidTimeoutValue,
				PathToCertFile: validPathToCertFile,
			},
		},
		{
			name:    "invalid address (not a URL) => invalid config",
			fields:  validFieldsWithAddress("not a valid address"),
			wantErr: true,
		},
		{
			name:    "invalid address (empty) => invalid config",
			fields:  validFieldsWithAddress(""),
			wantErr: true,
		},
		{
			name:   "valid address (localhost:8080) and timeout => valid config",
			fields: validFieldsWithAddress("localhost:8080"),
		},
		{
			name:   "valid address (http://localhost:8080) and timeout => valid config",
			fields: validFieldsWithAddress("http://localhost:8080"),
		},
		{
			name:   "valid address (http://127.0.0.1:8080) and timeout => valid config",
			fields: validFieldsWithAddress("http://127.0.0.1:8080"),
		},
		{
			name:   "valid address (http://12.34.56.78:1234) and timeout => valid config",
			fields: validFieldsWithAddress("http://12.34.56.78:1234"),
		},
		{
			name:   "valid address (https://12.34.56.78:1234) and timeout => valid config",
			fields: validFieldsWithAddress("https://12.34.56.78:1234"),
		},
		{
			name:   "valid address (127.0.0.1:8080 without scheme) and timeout => valid config",
			fields: validFieldsWithAddress("127.0.0.1:8080"),
		},
		{
			name:   "valid address (12.34.56.78:1234 without scheme) and timeout => valid config",
			fields: validFieldsWithAddress("12.34.56.78:1234"),
		},
		{
			name:   "weird address (//http://12.34.56.78:1234) and timeout => config still valid",
			fields: validFieldsWithAddress("//http://12.34.56.78:1234"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfg := &config.SidecarConfig{
				Enabled:        tt.fields.Enabled,
				Address:        tt.fields.Address,
				Timeout:        tt.fields.Timeout,
				PathToCertFile: tt.fields.PathToCertFile,
			}

			if err := cfg.ValidateBasic(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateBasic() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
