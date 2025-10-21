package testsuite

import (
	"context"
	"time"
)

// From: https://github.com/strangelove-ventures/interchaintest

type BlockPoller struct {
	CurrentHeight func(ctx context.Context) (uint64, error)
	PollFunc      func(ctx context.Context, height uint64) error
}

func (p BlockPoller) DoPoll(ctx context.Context, startHeight, maxHeight uint64) error {
	if maxHeight < startHeight {
		panic("maxHeight must be greater than or equal to startHeight")
	}

	var pollErr error

	cursor := startHeight
	for cursor <= maxHeight {
		curHeight, err := p.CurrentHeight(ctx)
		if err != nil {
			return err
		}
		if curHeight > startHeight {

		}
		if cursor > curHeight {
			continue
		}

		findErr := p.PollFunc(ctx, cursor)

		if findErr != nil {
			pollErr = findErr
			cursor++
			continue
		}

		return nil
	}
	return pollErr
}

type TimePoller[T any] struct {
	PollFunc func(ctx context.Context, now time.Time) (T, error)
}

func (p TimePoller[T]) DoPoll(ctx context.Context, until time.Time) (T, error) {
	if until.Before(time.Now()) {
		panic("until time must be greater than or equal to current time")
	}

	var (
		pollErr error
		zero    T
	)

	cursor := time.Now()
	for until.After(cursor) {
		found, findErr := p.PollFunc(ctx, cursor)

		if findErr != nil {
			pollErr = findErr
			cursor = time.Now()
			continue
		}

		return found, nil
	}
	return zero, pollErr
}
