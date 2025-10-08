package profiler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
	blobhub "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/parquet"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/report"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/sequencer"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// BlobProfiler is the main profiling engine
type BlobProfiler struct {
	logger    *slog.Logger
	report    *report.ProfilerReport
	blobhub   *blobhub.Client
	sequencer *sequencer.Client
	blobpool  *blobhub.Client
	generator *blobgen.BlobGenerator
	config    *config.Config
	handler   *parquet.Handler

	genBlobCount int

	addTime     sync.Mutex          // protects blocktimes
	blockTimes  map[int64]time.Time // block height -> timestamp
	*poolStatus                     // status for blobs in blobpool
}

// NewBlobProfiler creates a new blob profiler instance
func NewBlobProfiler(
	ctx context.Context, cfg *config.Config, logger *slog.Logger,
	handler *parquet.Handler, profilerReport *report.ProfilerReport,
) (*BlobProfiler, error) {
	// Initialize blobhub client
	blobhubClient := blobhub.NewClient(&blobhub.ClientConfig{
		BaseURL: cfg.BlobhubURL,
	})

	// Initialize sequencer client
	sequencerClient, err := sequencer.NewClient(ctx, cfg.SequencerRPC, cfg.Topic, cfg.Sender)
	if err != nil {
		profilerReport.RecordConnectionEvent("sequencer", cfg.SequencerRPC, err)
		return nil, fmt.Errorf("failed to create sequencer client: %w", err)
	}

	// Initialize blobpool client
	blobpoolClient := blobhub.NewClient(&blobhub.ClientConfig{
		BaseURL: cfg.BlobpoolURL,
	})

	return &BlobProfiler{
		logger:    logger,
		report:    profilerReport,
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

func (p *BlobProfiler) Close() {
	err := p.blobhub.Close()
	if err != nil {
		p.report.RecordConnectionEvent("blobhub", "close", err)
		p.logger.Error("failed to close blobhub connection", "error", err)
	}
	err = p.blobpool.Close()
	if err != nil {
		p.report.RecordConnectionEvent("blobpool", "close", err)
		p.logger.Error("failed to close blobpool server connection", "error", err)
	}
}
