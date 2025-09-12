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
	logger     *slog.Logger
	blobWriter *Writer
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

	path, err := filepath.Abs(filepath.Join(parquetDir, "blobs.parquet"))
	if err != nil {
		l.Error("failed to get absolute path for parquet directory", "error", err)
		return nil, err
	}

	// Create parquet writer if output file is specified
	l.Info("Creating parquet writer", "dir", parquetDir, "file", path)
	parquetWriter, err := newWriter(path, logger, batchSize)
	if err != nil {
		l.Error("failed to create parquet writer", "error", err, "file", path)
		return nil, err
	}

	return &Handler{
		logger:     l,
		blobWriter: parquetWriter,
	}, nil
}

func (h *Handler) WriteBlobs(blobs []*types.TrackedBlob) error {
	return h.blobWriter.writeBlobs(blobs)
}

func (h *Handler) GetStats() Stats {
	return h.blobWriter.GetStats()
}

func (h *Handler) Flush() error {
	return h.blobWriter.Flush()
}

func (h *Handler) Close() error {
	if h.blobWriter != nil {
		if err := h.blobWriter.Close(); err != nil {
			h.logger.Error("failed to close parquet writer", "error", err)
			return err
		}
	}
	return nil
}
