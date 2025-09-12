package parquet

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

const (
	defaultBatchSize = 1000
)

type Handler struct {
	logger           *slog.Logger
	blobWriter       *Writer
	throughputWriter *ThroughputWriter
}

// New creates a new parquet handler with default settings
func New(parquetDir string, logger *slog.Logger) (*Handler, error) {
	return NewWithOptions(parquetDir, logger, defaultBatchSize)
}

// NewWithOptions creates a new parquet handler with custom batch size
func NewWithOptions(parquetDir string, logger *slog.Logger, batchSize int) (*Handler, error) {
	l := logger.With("component", "parquet")

	// Validate parquet directory
	if parquetDir == "" {
		return nil, fmt.Errorf("parquet directory cannot be empty")
	}

	// Ensure directory exists
	if err := os.MkdirAll(parquetDir, 0755); err != nil {
		l.Error("failed to create parquet directory", "error", err)
		return nil, err
	}

	blobPath, err := filepath.Abs(filepath.Join(parquetDir, "blobs.parquet"))
	if err != nil {
		l.Error("failed to get absolute path for blob parquet directory", "error", err)
		return nil, err
	}

	throughputPath, err := filepath.Abs(filepath.Join(parquetDir, "throughput.parquet"))
	if err != nil {
		l.Error("failed to get absolute path for throughput parquet directory", "error", err)
		return nil, err
	}

	// Create parquet writers
	l.Info("Creating blob parquet writer", "dir", parquetDir, "file", blobPath)
	blobWriter, err := newBlobWriter(blobPath, logger, batchSize)
	if err != nil {
		l.Error("failed to create blob parquet writer", "error", err, "file", blobPath)
		return nil, err
	}

	l.Info("Creating throughput parquet writer", "dir", parquetDir, "file", throughputPath)
	throughputWriter, err := newThroughputWriter(throughputPath, logger, batchSize)
	if err != nil {
		l.Error("failed to create throughput parquet writer", "error", err, "file", throughputPath)
		return nil, err
	}

	return &Handler{
		logger:           l,
		blobWriter:       blobWriter,
		throughputWriter: throughputWriter,
	}, nil
}

func (h *Handler) WriteBlobs(blobs []*types.TrackedBlob) error {
	return h.blobWriter.writeBlobs(blobs)
}

func (h *Handler) WriteThroughput(record *ThroughputRecord) error {
	return h.throughputWriter.WriteThroughput(record)
}

func (h *Handler) GetStats() Stats {
	return h.blobWriter.GetStats()
}

func (h *Handler) Flush() error {
	if err := h.blobWriter.Flush(); err != nil {
		return err
	}
	if err := h.throughputWriter.Flush(); err != nil {
		return err
	}
	return nil
}

func (h *Handler) Close() error {
	if h.blobWriter != nil {
		if err := h.blobWriter.Close(); err != nil {
			h.logger.Error("failed to close blob parquet writer", "error", err)
			return err
		}
	}
	if h.throughputWriter != nil {
		if err := h.throughputWriter.Close(); err != nil {
			h.logger.Error("failed to close throughput parquet writer", "error", err)
			return err
		}
	}
	return nil
}
