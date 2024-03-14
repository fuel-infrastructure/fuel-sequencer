package testsuite

import (
	"context"
	"strconv"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// QuerySidecarBlockEvents gets the block events for the specified block number
func (s *E2ETestSuite) QuerySidecarBlockEvents(ctx context.Context, blockNumber int) ([]*sidecartypes.Event, error) {
	sidecarClient := s.getSidecarClient()

	resp, err := sidecarClient.GetBlockEvents(ctx,
		&sidecartypes.QueryBlockEventsRequest{BlockNumber: strconv.Itoa(blockNumber)},
	)
	if err != nil {
		return nil, err
	} else {
		return resp.Events, nil
	}
}
