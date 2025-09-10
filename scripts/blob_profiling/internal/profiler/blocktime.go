package profiler

import (
	"context"
	"time"
)

const (
	blockTimeTimeout = 10 * time.Second
)

func (p *BlobProfiler) blocktime(ctx context.Context, height int64) time.Time {
	if t, ok := p.blockTimes[height]; ok {
		return t
	}

	timeout := time.NewTimer(blockTimeTimeout)
	p.addTime.Lock()
	defer p.addTime.Unlock()

	for {
		select {
		case <-ctx.Done():
			return time.Time{}
		case <-timeout.C:
			return time.Time{}
		default:
			blocktime, err := p.sequencer.QueryBlockTime(ctx, height)
			if err != nil {
				return time.Time{}
			}

			p.blockTimes[height] = blocktime

			// Remove block times that are older than the blob timeout
			for h, t := range p.blockTimes {
				if t.Before(blocktime.Add(-p.config.BlobTimeout)) {
					delete(p.blockTimes, h)
				}
			}

			return blocktime
		}
	}
}
