package parquet

import (
	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

type WriterInterface interface {
	WriteBlobs(blobs []*types.TrackedBlob) error
	Flush() error
	Close() error
	GetStats() Stats
}

type ParquetHandler interface {
	WriterInterface
	Flush() error
	Close() error
	GetStats() Stats
}

var _ ParquetHandler = (*Handler)(nil)
