package testsuite

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// WaitForBlocks blocks until all chains reach a block height delta equal to or greater than the delta argument.
// If a ChainHeighter does not monotonically increase the height, this function may block program execution indefinitely.
func (s *E2ETestSuite) WaitForBlocks(ctx context.Context, delta int, timeoutAfter time.Duration) error {

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

	// Wait for blocks and timeout if it takes too long.
	select {
	case <-time.After(timeoutAfter):
		return errors.New(fmt.Sprintf("timed out waiting for %d blocks", delta))
	case <-done:
		return nil
	}
}
