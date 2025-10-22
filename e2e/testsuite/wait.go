package testsuite

import (
	"context"
	"fmt"
	"time"
)

func (s *E2ETestSuite) Sleep(duration time.Duration) {
	s.Logger().Info(fmt.Sprintf("Sleeping for %s", duration))
	time.Sleep(duration)
}

// WaitForSequencerBlocks blocks until Sequencer reaches a block height delta greater or equal to the delta argument.
func (s *E2ETestSuite) WaitForSequencerBlocks(ctx context.Context, delta int, timeoutAfter time.Duration) error {

	s.Logger().Info(fmt.Sprintf("Waiting for %d Sequencer block(s)", delta))

	start, err := s.Chain.FuelSequencerHeight(ctx)
	s.Require().NoError(err)
	end := start + uint64(delta)

	// Poll every 1 second until the target height is reached.
	done := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			latest, err := s.Chain.FuelSequencerHeight(ctx)
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
		return fmt.Errorf("timed out waiting for %d blocks", delta)
	case <-done:
		return nil
	}
}

// WaitUntilSequencerBlock blocks until Sequencer reaches a height greater or equal to the block argument.
func (s *E2ETestSuite) WaitUntilSequencerBlock(ctx context.Context, block int, timeoutAfter time.Duration) error {

	s.Logger().Info(fmt.Sprintf("Waiting for Sequencer block %d", block))

	// Poll every 1 second until the target height is reached.
	done := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			latest, err := s.Chain.FuelSequencerHeight(ctx)
			s.Require().NoError(err)
			if latest >= uint64(block) {
				close(done)
				return
			}
		}
	}()

	// Wait for block and timeout if it takes too long.
	select {
	case <-time.After(timeoutAfter):
		return fmt.Errorf("timed out waiting for block %d", block)
	case <-done:
		return nil
	}
}

// WaitForEthereumBlocks blocks until Ethereum reaches a block height delta greater or equal to the delta argument.
func (s *E2ETestSuite) WaitForEthereumBlocks(ctx context.Context, delta int, timeoutAfter time.Duration) error {

	s.Logger().Info(fmt.Sprintf("Waiting for %d Ethereum block(s)", delta))

	start, err := s.Chain.EthereumHeight(ctx)
	s.Require().NoError(err)
	end := start + uint64(delta)

	// Poll every 1 second until the target height is reached.
	done := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			latest, err := s.Chain.EthereumHeight(ctx)
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
		return fmt.Errorf("timed out waiting for %d blocks", delta)
	case <-done:
		return nil
	}
}

// WaitUntilEthereumBlock blocks until Ethereum reaches a height greater or equal to the block argument.
func (s *E2ETestSuite) WaitUntilEthereumBlock(ctx context.Context, block int, timeoutAfter time.Duration) error {

	s.Logger().Info(fmt.Sprintf("Waiting for Ethereum block %d", block))

	// Poll every 1 second until the target height is reached.
	done := make(chan struct{}, 1)
	go func() {
		for {
			time.Sleep(time.Second)
			latest, err := s.Chain.EthereumHeight(ctx)
			s.Require().NoError(err)
			if latest >= uint64(block) {
				close(done)
				return
			}
		}
	}()

	// Wait for block and timeout if it takes too long.
	select {
	case <-time.After(timeoutAfter):
		return fmt.Errorf("timed out waiting for block %d", block)
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
