package profiler

import (
	"context"
	"fmt"
	"log/slog"
	"math"
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
	sequencerClient, err := sequencer.NewClient(ctx, cfg.SequencerRPC, cfg.Topic, cfg.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to create sequencer client: %w", err)
	}

	return &BlobProfiler{
		logger:    logger,
		blobhub:   blobhubClient,
		sequencer: sequencerClient,
		generator: blobgen.NewBlobGenerator(42, cfg.BlobDistribution),
		config:    cfg,
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

	var currentThroughput, dataSubmitted int64
	var blobCount uint64
	nextBlob := p.generateBlob()
	blobs := make([]*types.TrackedBlob, 0)

	var duration time.Duration
	start := time.Now()
	plannedRate := p.config.Rate(0)

	lastlog := start
	for pctx.Err() == nil && plannedRate <= p.config.MaxRate {
		duration = time.Since(start)
		seconds := duration.Seconds()
		currentThroughput = int64(math.Round(float64(dataSubmitted) / seconds))
		plannedRate = p.config.Rate(duration)

		if time.Since(lastlog) > logFrequency {
			p.logger.Info("Throughput",
				"expected_KiB/s", plannedRate/size.KiB,
				"throughput_KiB/s", currentThroughput/size.KiB,
				"blob_count", blobCount,
				"total_KiB", dataSubmitted/size.KiB,
				"duration_s", seconds,
			)
			lastlog = time.Now()
		}
		if plannedRate < currentThroughput {
			continue // ahead of planned throughput, slow down till back on track
		}

		dataSubmitted += nextBlob.Size
		p.castBlob(pctx, cancel, nextBlob, blobCount)
		blobs = append(blobs, nextBlob)
		blobCount++

		nextBlob = p.generateBlob()
	}

	return blobs, nil
}

func (p *BlobProfiler) Close() {
	p.blobhub.Close()
}
