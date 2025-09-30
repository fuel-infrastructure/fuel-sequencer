package profiler

import (
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
)

func (p *BlobProfiler) setupProceed(lagging *time.Time, start time.Time, plannedRate *int, currentThroughput *int) (condition func() bool) {
	// By default, always check for lag
	lagRatio := func() float32 {
		if *currentThroughput == 0 {
			return 0
		}
		return float32(*plannedRate) / float32(*currentThroughput)
	}
	lagCondition := func() bool {
		if lagging == nil {
			if lagRatio() >= p.config.MaxLagRatio { // lag detected - start timing
				lt := time.Now()
				lagging = &lt
			}
			return true
		} else {
			if lr := lagRatio(); lr < p.config.MaxLagRatio { // lag recovered within tolerance
				lagging = nil
				return true
			} else {
				if tolerating := time.Since(*lagging); tolerating > p.config.LagTolerance {
					p.logger.Info("lag exceeded tolerance - halting...",
						"current_throughput_KiB/s", *currentThroughput/size.KiB, "planned_rate_KiB/s", *plannedRate/size.KiB,
						"lag_tolerated", tolerating, "max_lag_tolerance", p.config.LagTolerance,
						"lag_ratio", lr, "max_lag_ratio", p.config.MaxLagRatio)
					return false // lag exceeded tolerance - stop
				}
			}
		}
		return true
	}

	noDuration := p.config.Duration == 0
	noMaxRate := p.config.MaxRate == 0
	if noDuration && noMaxRate {
		p.logger.Info("no duration or rate limit set, will run indefinitely")
		return lagCondition
	}

	durationCheck := func() bool {
		d := time.Since(start)
		c := p.config.Duration
		resume := d < p.config.Duration
		if !resume {
			p.logger.Info("duration reached", "current_duration", d, "limit_duration", c)
		}
		return resume
	}
	rateCheck := func() bool {
		resume := *plannedRate <= p.config.MaxRate
		if !resume {
			p.logger.Info("max planned rate reached",
				"limit_rate", p.config.MaxRate/size.KiB,
				"planned_rate", *plannedRate/size.KiB,
				"current_rate", *currentThroughput/size.KiB,
			)
		}
		return resume
	}

	if noDuration {
		p.logger.Info(
			"duration not set, only rate limit - will stop when rate limit is reached",
			"rate_KiB/s", *plannedRate/size.KiB,
		)
		return func() bool { return lagCondition() && rateCheck() }
	}
	if noMaxRate {
		p.logger.Info(
			"rate limit not set, only duration - will stop when duration is reached",
			"duration", p.config.Duration,
		)
		return func() bool { return lagCondition() && durationCheck() }
	}
	p.logger.Info(
		"both duration and rate limit set, will stop when either is reached",
		"duration", p.config.Duration, "rate_KiB/s", *plannedRate/size.KiB,
	)
	return func() bool { return lagCondition() && durationCheck() && rateCheck() }
}
