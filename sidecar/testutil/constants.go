package testutil

import (
	"time"

	"cosmossdk.io/log"

	sidecarconfig "github.com/fuel-infrastructure/fuel-sequencer/sidecar/config"
)

var (
	ValidSidecarConfig = sidecarconfig.SidecarConfig{
		Enabled:        true,
		Address:        "localhost:8000",
		Timeout:        3 * time.Second,
		PathToCertFile: "path/to/cert/file",
	}
	InvalidSidecarConfig = sidecarconfig.SidecarConfig{
		Enabled:        true,
		Address:        "localhost:8000",
		Timeout:        0, // Timeout cannot be zero
		PathToCertFile: "path/to/cert/file",
	}
	DisabledSidecarConfig = sidecarconfig.SidecarConfig{
		Enabled:        false,
		Address:        "localhost:8000",
		Timeout:        3 * time.Second,
		PathToCertFile: "path/to/cert/file",
	}

	ValidLogger = log.NewNopLogger()
)
