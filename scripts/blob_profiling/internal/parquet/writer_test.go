package parquet

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

func TestParquetHandler(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test_parquet_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create handler
	handler, err := New(tmpDir, logger, "test")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	defer handler.Close()

	// Create multiple test blobs
	var blobs []*types.TrackedBlob
	for i := 0; i < 5; i++ {
		testData := []byte("test blob data " + string(rune(i+'0')))
		key := store.NewKey(testData)
		blob := &types.TrackedBlob{
			StoredBlob: store.StoredBlob{
				Receipt: store.Receipt{
					Key: key,
				},
				Data: testData,
			},
			Size: len(testData),
			Submission: &types.Submission{
				Status:    types.Stored,
				StartTime: time.Now(),
				StoreTime: time.Now().Add(100 * time.Millisecond),
			},
		}
		blobs = append(blobs, blob)
	}

	// Write blobs through handler
	err = handler.WriteBlobs(blobs)
	if err != nil {
		t.Fatalf("Failed to write blobs: %v", err)
	}

	// Flush and close
	err = handler.Flush()
	if err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Close handler to ensure data is written
	err = handler.Close()
	if err != nil {
		t.Fatalf("Failed to close handler: %v", err)
	}

	// Verify file was created and has content
	blobFile := tmpDir + "/test_blobs.parquet"
	stat, err := os.Stat(blobFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("Parquet file is empty")
	}

	t.Logf("Created parquet file: %s (size: %d bytes)", blobFile, stat.Size())
}

func TestParquetWriterBatch(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test_batch_*.parquet")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// Create writer with small batch size
	writer, err := newBlobWriter(tmpFile.Name(), logger, 3)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}

	// Create multiple test blobs
	var blobs []*types.TrackedBlob
	for i := 0; i < 5; i++ {
		testData := []byte("test blob data " + string(rune(i+'0')))
		key := store.NewKey(testData)
		blob := &types.TrackedBlob{
			StoredBlob: store.StoredBlob{
				Receipt: store.Receipt{
					Key: key,
				},
				Data: testData,
			},
			Size: len(testData),
			Submission: &types.Submission{
				Status:    types.Stored,
				StartTime: time.Now(),
				StoreTime: time.Now().Add(100 * time.Millisecond),
			},
		}
		blobs = append(blobs, blob)
	}

	// Write blobs in batches (using private method)
	err = writer.writeBlobs(blobs)
	if err != nil {
		t.Fatalf("Failed to write blobs: %v", err)
	}

	// Flush and close
	err = writer.Flush()
	if err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Close writer to ensure data is written
	err = writer.Close()
	if err != nil {
		t.Fatalf("Failed to close writer: %v", err)
	}

	// Verify file was created and has content
	stat, err := os.Stat(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("Parquet file is empty")
	}

	t.Logf("Created parquet file: %s (size: %d bytes)", tmpFile.Name(), stat.Size())
}
