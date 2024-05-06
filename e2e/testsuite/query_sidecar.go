package testsuite

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

// PollForSidecarBlockEvents polls until events are found at the specified block number.
// Note: This function times out after deltaTime duration.
func (s *E2ETestSuite) PollForSidecarBlockEvents(
	ctx context.Context, deltaTime time.Duration, blockNumber int,
) ([]*sidecartypes.Event, error) {
	s.T().Log(fmt.Sprintf("Polling for sidecar block events at height %d", blockNumber))

	doPoll := func(ctx context.Context, now time.Time) (any, error) {
		res, err := s.QuerySidecarBlockEvents(ctx, blockNumber)
		if err != nil {
			return nil, fmt.Errorf("events not found at block %d: %s", blockNumber, err.Error())
		}
		return res, nil
	}

	bp := TimePoller[any]{PollFunc: doPoll}
	res, err := bp.DoPoll(ctx, time.Now().Add(deltaTime))
	if err != nil {
		return nil, err
	} else {
		return res.([]*sidecartypes.Event), nil
	}
}
