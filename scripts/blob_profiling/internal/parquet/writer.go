package parquet

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"
)

// Writer handles incremental writing of blob profiling data to parquet files
type Writer struct {
	file             *os.File
	writer           *parquet.GenericWriter[*BlobProfileRecord]
	logger           *slog.Logger
	mutex            sync.Mutex
	buffer           []*BlobProfileRecord
	batchSize        int
	profileStartTime time.Time
}

// newBlobWriter creates a new parquet writer for blob profiling data
func newBlobWriter(filePath string, logger *slog.Logger, batchSize int) (*Writer, error) {
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create parquet file %s: %w", filePath, err)
	}

	// Create parquet writer with optimized settings
	writer := parquet.NewGenericWriter[*BlobProfileRecord](file,
		parquet.Compression(&parquet.Snappy),
		parquet.PageBufferSize(64*1024), // 64KB page buffer
	)

	return &Writer{
		file:             file,
		writer:           writer,
		logger:           logger,
		buffer:           make([]*BlobProfileRecord, 0, batchSize),
		batchSize:        batchSize,
		profileStartTime: time.Now(),
	}, nil
}

// Flush writes any pending records to the parquet file
func (w *Writer) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.flushUnsafe()
}

// flushUnsafe performs the actual flush operation (must be called with mutex held)
func (w *Writer) flushUnsafe() error {
	if len(w.buffer) == 0 {
		return nil
	}

	// Write the buffered records
	n, err := w.writer.Write(w.buffer)
	if err != nil {
		return fmt.Errorf("failed to write parquet records: %w", err)
	}

	w.logger.Debug("Flushed parquet records", "count", len(w.buffer), "written", n)

	// Clear the buffer
	w.buffer = w.buffer[:0]

	return nil
}

// Close flushes any remaining data and closes the parquet file
func (w *Writer) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Flush any remaining records
	if err := w.flushUnsafe(); err != nil {
		w.logger.Error("Failed to flush final records", "error", err)
	}

	// Flush the parquet writer before closing
	if err := w.writer.Flush(); err != nil {
		w.logger.Error("Failed to flush parquet writer", "error", err)
		return err
	}

	// Close the parquet writer
	if err := w.writer.Close(); err != nil {
		w.logger.Error("Failed to close parquet writer", "error", err)
		return err
	}

	// Close the file
	if err := w.file.Close(); err != nil {
		w.logger.Error("Failed to close parquet file", "error", err)
		return err
	}

	w.logger.Info("Parquet file written successfully", "file", w.file.Name())
	return nil
}
