package testsuite

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func (s *E2ETestSuite) Sleep(duration time.Duration) {
	s.Logger().Info(fmt.Sprintf("Sleeping for %s", duration))
	time.Sleep(duration)
}

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

func WaitForCondition(timeoutAfter, pollingInterval time.Duration, fn func() (bool, error)) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutAfter)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed waiting for condition after %f seconds", timeoutAfter.Seconds())
		case <-time.After(pollingInterval):
			reachedCondition, err := fn()
			if err != nil {
				return fmt.Errorf("error occurred while waiting for condition: %s", err)
			}

			if reachedCondition {
				return nil
			}
		}
	}
}
