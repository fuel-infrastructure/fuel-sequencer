package apptesting

import sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"

type MockQueryBlockEventsResponse struct {
	Response *sidecartypes.QueryBlockEventsResponse
	Error    error
}
