package testsuite

import (
	"context"
	"errors"
	"time"
)

// WaitForBlocks blocks until all chains reach a block height delta equal to or greater than the delta argument.
// If a ChainHeighter does not monotonically increase the height, this function may block program execution indefinitely.
func (s *E2ETestSuite) WaitForBlocks(ctx context.Context, delta int) error {

	start, err := s.chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)
	end := start + uint64(delta)

	// Poll every 1 second until the target height is reached.
	done := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			latest, err := s.chain.FuelSequencerHeight(ctx)
			s.Require().NoError(err)
			if latest >= end {
				close(done)
				return
			}
		}
	}()

	// Time out after 30 seconds.
	// TODO: make customisable
	select {
	case <-time.After(30 * time.Second):
		return errors.New("timed out")
	case <-done:
		return nil
	}
}
