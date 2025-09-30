package parquet

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/config"
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
	runDirName := config.RunOutputDir(time.Now())
	handler, err := New(tmpDir, runDirName, logger, "test")
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
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
	// The handler creates a timestamped directory, so we need to find it
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	var blobFile string
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() == "test" {
			// Found the test directory, now look for the timestamped subdirectory
			testDir := filepath.Join(tmpDir, entry.Name())
			subEntries, err := os.ReadDir(testDir)
			if err != nil {
				t.Fatalf("Failed to read test directory: %v", err)
			}
			for _, subEntry := range subEntries {
				if subEntry.IsDir() {
					// Found the timestamped directory, now look for the parquet file
					blobFile = filepath.Join(testDir, subEntry.Name(), "blobs.parquet")
					break
				}
			}
			break
		}
	}

	if blobFile == "" {
		t.Fatalf("Could not find timestamped test directory with blobs.parquet file")
	}

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
	blobWriter, err := newWriter[BlobProfileRecord](tmpFile.Name(), logger, 3)
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
	err = blobWriter.write(writeBlobs(blobs))
	if err != nil {
		t.Fatalf("Failed to write blobs: %v", err)
	}

	// Flush and close
	err = blobWriter.Flush()
	if err != nil {
		t.Fatalf("Failed to flush: %v", err)
	}

	// Close writer to ensure data is written
	err = blobWriter.Close()
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
