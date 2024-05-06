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

// PollForSidecarBlockEvents polls until the Sidecar returns a result for the specified block number, which can be an
// empty or non-empty list of events. We are assuming that the Sidecar will error if it did not process the specified
// Ethereum block height, in which case we will retry until the specified time delta.
//
// Note: in any case, this function times out with an error after the specified time delta.
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
