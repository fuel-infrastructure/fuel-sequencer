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
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
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
	generator *blobgen.BlobGenerator
	config    *config.Config

	buffer     []*types.TrackedBlob
	bufferSize int

	addTime    sync.Mutex          // protects blocktimes
	blockTimes map[int64]time.Time // block height -> timestamp
}

// NewBlobProfiler creates a new blob profiler instance
func NewBlobProfiler(
	ctx context.Context, cfg *config.Config, logger *slog.Logger,
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

	return &BlobProfiler{
		logger:     logger,
		blobhub:    blobhubClient,
		sequencer:  sequencerClient,
		generator:  blobgen.NewBlobGenerator(42, cfg.BlobDistribution),
		config:     cfg,
		addTime:    sync.Mutex{},
		blockTimes: make(map[int64]time.Time),
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

	var currentThroughput, dataSubmitted int
	txCount := p.sequencer.Sender.Sequence // start by considering existing transactions
	var blobCount int
	p.supplementBuffer(0)
	blobs := make([]*types.TrackedBlob, 0)

	var duration time.Duration
	start := time.Now()
	plannedRate := p.config.Rate(0)

	lastlog := start
	for pctx.Err() == nil && plannedRate <= p.config.MaxRate {
		duration = time.Since(start)
		seconds := duration.Seconds()
		currentThroughput = int(math.Round(float64(dataSubmitted) / seconds))
		plannedRate = p.config.Rate(duration)

		if time.Since(lastlog) > logFrequency {
			p.logger.Info("throughput",
				"expected_KiB/s", plannedRate/size.KiB,
				"actual_KiB/s", currentThroughput/size.KiB,
				"submitted_txs", txCount-p.sequencer.Sender.Sequence,
				"submitted_count", blobCount,
				"submitted_KiB", dataSubmitted/size.KiB,
				"upcoming_count", len(p.buffer),
				"upcoming_KiB", p.bufferSize/size.KiB,
				"duration_s", seconds,
			)
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
		go p.catchBlobs(pctx, cancel, txHash, nextBlobs)
		blobs = append(blobs, nextBlobs...)
		blobCount += len(nextBlobs)
		txCount++

		p.supplementBuffer(duration)
	}

	return blobs, nil
}

func (p *BlobProfiler) Close() {
	p.blobhub.Close()
}

func (p *BlobProfiler) collectBlobs(duration time.Duration, currentSize int) (
	[]*types.TrackedBlob, int,
) {
	// plannedRate < currentThroughput is already handled

	// Figure out how many blobs to send
	expectedSize := p.config.Size(duration)
	needSize := expectedSize - currentSize

	// Yoink from upcomingBlobs into nextBlobs until expectedSize is reached
	nextBlobs := make([]*types.TrackedBlob, 0)
	nextBlobsSize := 0
	for _, blob := range p.buffer {
		nextBlobs = append(nextBlobs, blob)
		nextBlobsSize += blob.Size
		p.buffer, p.bufferSize = p.buffer[1:], p.bufferSize-blob.Size

		if nextBlobsSize > needSize {
			return nextBlobs, nextBlobsSize
		}
	}

	p.logger.Error("didn't get enough blobs",
		"need_size_KiB", needSize/size.KiB,
		"have_size_KiB", nextBlobsSize/size.KiB,
	)
	return nextBlobs, nextBlobsSize
}
