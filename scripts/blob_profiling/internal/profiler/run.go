package profiler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/size"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// RunProfile executes the complete profiling sequence
func (p *BlobProfiler) RunProfile(ctx context.Context) ([]*types.TrackedBlob, error) {
	p.logger.Info("Starting blob throughput profiling")
	p.logger.Info("Initial rate", "rate_KiB", p.config.Rate(0)/size.KiB)
	p.logger.Info("Maximum rate", "rate_MiB", p.config.MaxRate/size.MiB)
	p.logger.Info("Max lag ratio threshold", "ratio", p.config.MaxLagRatio)
	p.logger.Info("Lag tolerance", "time_s", p.config.LagTolerance.Seconds())

	pctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup blob buffer and counting
	var currentThroughput, dataSubmitted float64
	txCount := p.sequencer.Sender.Sequence // start by considering existing transactions
	var blobCount int
	blobs := make([]*types.TrackedBlob, 0)
	catching := sync.WaitGroup{} // Ensure txs are caught before final flushing of parquet data

	// Handle blobpool tracking and messaging
	stream, err := p.blobpool.StreamBlobs(pctx)
	if err != nil {
		p.report.RecordConnectionEvent("blobpool", p.config.BlobpoolURL, err)
		return nil, fmt.Errorf("failed to create blobpool stream: %w", err)
	}
	expect := make(chan *types.TrackedBlob, p.bufferSize)
	consume := make(chan store.Key, p.bufferSize)
	go p.catchBlobpool(pctx, expect, stream, consume)

	// Setup timing and rate tracking
	var duration time.Duration
	var lagging *time.Time
	start := time.Now()
	plannedRate := p.config.Rate(0)
	proceed := p.setupProceed(lagging, start, &plannedRate, &currentThroughput)

	lastlog := start
	for pctx.Err() == nil && proceed() {
		duration = time.Since(start)
		seconds := duration.Seconds()
		currentThroughput = dataSubmitted / seconds
		plannedRate = p.config.Rate(duration)

		if time.Since(lastlog) > logFrequency {
			lastlog = time.Now()
			p.logThroughput(plannedRate, currentThroughput, txCount, dataSubmitted, blobCount, seconds, lastlog)

		}
		if plannedRate < currentThroughput {
			continue // ahead of planned throughput, slow down till back on track
		}

		nextBlobs, nextBlobsSize := p.collectBlobs(plannedRate, duration, dataSubmitted)
		if nextBlobsSize == 0 {
			continue
		}

		sizeKiB := nextBlobsSize / size.KiB
		p.logger.Debug("posting new blobs",
			"next_count", len(nextBlobs),
			"size_KiB", sizeKiB,
			"avg_size_KiB", sizeKiB/len(nextBlobs),
		)

		dataSubmitted += float64(nextBlobsSize)
		txHash, err := p.castBlobs(pctx, cancel, nextBlobs, txCount, blobCount)
		if err != nil {
			p.report.RecordCastingEventWithBlobs(txHash, nextBlobs, err)
			p.logger.Error("failed to cast blobs", "error", err)
			break
		}
		catching.Add(1)
		go func() {
			p.catchBlobs(pctx, cancel, expect, consume, txHash, nextBlobs)
			catching.Done()
		}()
		blobs = append(blobs, nextBlobs...)

		blobCount += len(nextBlobs)
		txCount++
	}
	catching.Wait()

	// Final flush of any remaining parquet data
	if err := p.handler.Flush(); err != nil {
		p.report.RecordParquetEvent("flush", err)
		p.logger.Error("Failed to flush final parquet data", "error", err)
	}

	// Update profiler report summary
	p.updateReportSummary(blobs, dataSubmitted)

	return blobs, nil
}
