package testsuite

import (
	"context"
	"time"
)

// From: https://github.com/strangelove-ventures/interchaintest

// sleepBetweenPolls prevents extremely fast polling but still keeps polling fast.
const sleepBetweenPolls = time.Second / 4

type BlockPoller[T any] struct {
	CurrentHeight func(ctx context.Context) (uint64, error)
	PollFunc      func(ctx context.Context, height uint64) (T, error)
}

func (p BlockPoller[T]) DoPoll(ctx context.Context, startHeight, maxHeight uint64) (T, error) {
	if maxHeight < startHeight {
		panic("maxHeight must be greater than or equal to startHeight")
	}

	var (
		pollErr error
		zero    T
	)

	cursor := startHeight
	for cursor <= maxHeight {
		curHeight, err := p.CurrentHeight(ctx)
		if err != nil {
			return zero, err
		}
		if cursor > curHeight {
			time.Sleep(sleepBetweenPolls)
			continue
		}

		found, findErr := p.PollFunc(ctx, cursor)

		if findErr != nil {
			pollErr = findErr
			cursor++
			time.Sleep(sleepBetweenPolls)
			continue
		}

		return found, nil
	}
	return zero, pollErr
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
			time.Sleep(sleepBetweenPolls)
			continue
		}

		return found, nil
	}
	return zero, pollErr
}
