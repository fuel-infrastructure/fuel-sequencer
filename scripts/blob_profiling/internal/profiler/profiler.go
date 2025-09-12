package profiler

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	blobhub "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/parquet"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/sequencer"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

const (
	logFrequency = 1 * time.Second
)

// BlobProfiler is the main profiling engine
type BlobProfiler struct {
	logger    *slog.Logger
	blobhub   *blobhub.Client
	sequencer *sequencer.Client
	blobpool  *blobhub.Client
	generator *blobgen.BlobGenerator
	config    *config.Config
	handler   *parquet.Handler

	genBlobCount int
	buffer       []*types.TrackedBlob
	bufferSize   int

	addTime     sync.Mutex          // protects blocktimes
	blockTimes  map[int64]time.Time // block height -> timestamp
	*poolStatus                     // status for blobs in blobpool
}

// NewBlobProfiler creates a new blob profiler instance
func NewBlobProfiler(
	ctx context.Context, cfg *config.Config, logger *slog.Logger,
	handler *parquet.Handler,
) (*BlobProfiler, error) {
	// Initialize blobhub client
	blobhubClient := blobhub.NewClient(&blobhub.ClientConfig{
		BaseURL: cfg.BlobhubURL,
	})

	// Initialize sequencer client
	sequencerClient, err := sequencer.NewClient(
		ctx, cfg.SequencerRPC, cfg.Topic, cfg.Sender, cfg.BlobTimeout,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sequencer client: %w", err)
	}

	// Initialize blobpool client
	blobpoolClient := blobhub.NewClient(&blobhub.ClientConfig{
		BaseURL: cfg.BlobpoolURL,
	})

	return &BlobProfiler{
		logger:    logger,
		blobhub:   blobhubClient,
		sequencer: sequencerClient,
		blobpool:  blobpoolClient,

		generator:  blobgen.NewBlobGenerator(42, cfg.BlobDistribution),
		config:     cfg,
		handler:    handler,
		addTime:    sync.Mutex{},
		blockTimes: make(map[int64]time.Time),
		poolStatus: &poolStatus{
			pending:  0,
			refs:     make(map[store.Key]*types.TrackedBlob),
			expected: make(map[store.Key]bool),
		},
	}, nil
}

// RunProfile executes the complete profiling sequence
func (p *BlobProfiler) RunProfile(ctx context.Context) ([]*types.TrackedBlob, error) {
	p.logger.Info("Starting blob throughput profiling")
	p.logger.Info("Initial rate", "rate_KiB", p.config.Rate(0)/size.KiB)
	p.logger.Info("Maximum rate", "rate_MiB", p.config.MaxRate/size.MiB)
	p.logger.Info("Max latency threshold", "latency", p.config.MaxLatency)

	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup blob buffer and counting
	var currentThroughput, dataSubmitted int
	txCount := p.sequencer.Sender.Sequence // start by considering existing transactions
	var blobCount int
	p.supplementBuffer(0)
	blobs := make([]*types.TrackedBlob, 0)

	// Handle blobpool tracking and messaging
	stream, err := p.blobpool.StreamBlobs(pctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create blobpool stream: %w", err)
	}
	expect := make(chan *types.TrackedBlob, p.bufferSize)
	consume := make(chan store.Key, p.bufferSize)
	go p.catchBlobpool(pctx, expect, stream, consume)

	// Setup timing and rate tracking
	var duration time.Duration
	start := time.Now()
	plannedRate := p.config.Rate(0)
	proceed := func() func() bool {
		noDuration := p.config.Duration == 0
		noMaxRate := p.config.MaxRate == 0
		if noDuration && noMaxRate {
			p.logger.Info("no duration or rate limit set, will run indefinitely")
			return func() bool { return true }
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
			resume := plannedRate <= p.config.MaxRate
			if !resume {
				p.logger.Info("max planned rate reached",
					"limit_rate", p.config.MaxRate/size.KiB,
					"planned_rate", plannedRate/size.KiB,
					"current_rate", currentThroughput/size.KiB,
				)
			}
			return resume
		}

		if noDuration {
			p.logger.Info(
				"duration not set, only rate limit - will stop when rate limit is reached",
				"rate_KiB/s", plannedRate/size.KiB,
			)
			return rateCheck
		}
		if noMaxRate {
			p.logger.Info(
				"rate limit not set, only duration - will stop when duration is reached",
				"duration", p.config.Duration,
			)
			return durationCheck
		}
		p.logger.Info(
			"both duration and rate limit set, will stop when either is reached",
			"duration", p.config.Duration, "rate_KiB/s", plannedRate/size.KiB,
		)
		return func() bool { return durationCheck() && rateCheck() }
	}()

	lastlog := start
	for pctx.Err() == nil && proceed() {
		duration = time.Since(start)
		seconds := duration.Seconds()
		currentThroughput = int(math.Round(float64(dataSubmitted) / seconds))
		plannedRate = p.config.Rate(duration)

		if time.Since(lastlog) > logFrequency {
			p.logThroughput(plannedRate, currentThroughput, txCount, dataSubmitted, blobCount, p.bufferSize, p.pending, seconds)

			lastlog = time.Now()
		}
		if plannedRate < currentThroughput {
			continue // ahead of planned throughput, slow down till back on track
		}

		nextBlobs, nextBlobsSize := p.collectBlobs(duration, dataSubmitted)

		sizeKiB := nextBlobsSize / size.KiB
		p.logger.Debug("posting new blobs",
			"next_count", len(nextBlobs),
			"size_KiB", sizeKiB,
			"avg_size_KiB", sizeKiB/len(nextBlobs),
		)

		dataSubmitted += nextBlobsSize
		txHash, err := p.castBlobs(pctx, cancel, nextBlobs, txCount, blobCount)
		if err != nil {
			p.logger.Error("failed to cast blobs", "error", err)
			cancel()
			return nil, err
		}
		go p.catchBlobs(pctx, cancel, expect, consume, txHash, nextBlobs)
		blobs = append(blobs, nextBlobs...)

		blobCount += len(nextBlobs)
		txCount++

		p.supplementBuffer(duration)
	}

	// Final flush of any remaining parquet data
	if err := p.handler.Flush(); err != nil {
		p.logger.Error("Failed to flush final parquet data", "error", err)
	}

	return blobs, nil
}

func (p *BlobProfiler) Close() {
	err := p.blobhub.Close()
	if err != nil {
		p.logger.Error("failed to close blobhub connection", "error", err)
	}
	err = p.blobpool.Close()
	if err != nil {
		p.logger.Error("failed to close blobpool server connection", "error", err)
	}
}

func (p *BlobProfiler) logThroughput(
	plannedRate int,
	currentThroughput int,
	txCount uint64,
	dataSubmitted int,
	blobCount int,
	bufferSize int,
	pending int,
	seconds float64,
) {
	expectedKiBPerSec := plannedRate / size.KiB
	actualKiBPerSec := currentThroughput / size.KiB
	submittedTxs := txCount - p.sequencer.Sender.Sequence
	submittedKiB := dataSubmitted / size.KiB
	upcomingKiB := p.bufferSize / size.KiB

	p.logger.Info("throughput",
		"expected_KiB/s", expectedKiBPerSec,
		"actual_KiB/s", actualKiBPerSec,
		"submitted_txs", submittedTxs,
		"submitted_count", blobCount,
		"submitted_KiB", submittedKiB,
		"upcoming_count", len(p.buffer),
		"upcoming_KiB", upcomingKiB,
		"pending_blobpool_count", p.pending, // concurrent access, but should be safe enough
		"duration_s", seconds,
	)

	// Write throughput data to parquet
	throughputRecord := parquet.NewThroughputRecord(
		int64(expectedKiBPerSec), int64(actualKiBPerSec),
		int64(submittedTxs), int64(blobCount), int64(submittedKiB),
		int64(len(p.buffer)), int64(upcomingKiB),
		int64(p.pending), seconds,
	)
	if err := p.handler.WriteThroughput(throughputRecord); err != nil {
		p.logger.Error("failed to write throughput data", "error", err)
	}
}
