package parquet

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/parquet-go/parquet-go"
)

// ThroughputWriter handles writing throughput metrics to parquet files
type ThroughputWriter struct {
	file             *os.File
	writer           *parquet.GenericWriter[*ThroughputRecord]
	logger           *slog.Logger
	mutex            sync.Mutex
	buffer           []*ThroughputRecord
	batchSize        int
	profileStartTime time.Time
}

// newThroughputWriter creates a new parquet writer for throughput data
func newThroughputWriter(filePath string, logger *slog.Logger, batchSize int) (*ThroughputWriter, error) {
	// Create the output file
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create throughput parquet file %s: %w", filePath, err)
	}

	// Create parquet writer with optimized settings
	writer := parquet.NewGenericWriter[*ThroughputRecord](file,
		parquet.Compression(&parquet.Snappy),
		parquet.PageBufferSize(64*1024), // 64KB page buffer
	)

	return &ThroughputWriter{
		file:             file,
		writer:           writer,
		logger:           logger,
		buffer:           make([]*ThroughputRecord, 0, batchSize),
		batchSize:        batchSize,
		profileStartTime: time.Now(),
	}, nil
}

// WriteThroughput adds throughput data to the write buffer and flushes if batch size is reached
func (w *ThroughputWriter) WriteThroughput(record *ThroughputRecord) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.buffer = append(w.buffer, record)

	// Flush if buffer is full
	if len(w.buffer) >= w.batchSize {
		return w.flushUnsafe()
	}

	return nil
}

// Flush writes any pending records to the parquet file
func (w *ThroughputWriter) Flush() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.flushUnsafe()
}

// flushUnsafe performs the actual flush operation (must be called with mutex held)
func (w *ThroughputWriter) flushUnsafe() error {
	if len(w.buffer) == 0 {
		return nil
	}

	// Write the buffered records
	n, err := w.writer.Write(w.buffer)
	if err != nil {
		return fmt.Errorf("failed to write throughput parquet records: %w", err)
	}

	w.logger.Debug("Flushed throughput parquet records", "count", len(w.buffer), "written", n)

	// Clear the buffer
	w.buffer = w.buffer[:0]

	return nil
}

// Close flushes any remaining data and closes the parquet file
func (w *ThroughputWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	// Flush any remaining records
	if err := w.flushUnsafe(); err != nil {
		w.logger.Error("Failed to flush final throughput records", "error", err)
	}

	// Flush the parquet writer before closing
	if err := w.writer.Flush(); err != nil {
		w.logger.Error("Failed to flush throughput parquet writer", "error", err)
		return err
	}

	// Close the parquet writer
	if err := w.writer.Close(); err != nil {
		w.logger.Error("Failed to close throughput parquet writer", "error", err)
		return err
	}

	// Close the file
	if err := w.file.Close(); err != nil {
		w.logger.Error("Failed to close throughput parquet file", "error", err)
		return err
	}

	w.logger.Info("Throughput parquet file written successfully", "file", w.file.Name())
	return nil
}

// GetStats returns statistics about the throughput writer
func (w *ThroughputWriter) GetStats() Stats {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return Stats{
		BufferedRecords: len(w.buffer),
		BatchSize:       w.batchSize,
		ProfileDuration: time.Since(w.profileStartTime),
	}
}
