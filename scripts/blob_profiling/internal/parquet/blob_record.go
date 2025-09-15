package parquet

import (
	"time"

	"github.com/fuel-infrastructure/fuel-sequencer/scripts/blob_profiling/internal/types"
)

// BlobProfileRecord represents a single blob profiling record for parquet storage
type BlobProfileRecord struct {
	// Blob identification
	Count int64  `parquet:"name=count,type=INT64"`
	Hash  string `parquet:"name=hash,type=BYTE_ARRAY,convertedtype=UTF8"`
	Size  int32  `parquet:"name=size,type=INT32"`

	// Submission metadata
	TxHash string `parquet:"name=tx_hash,type=BYTE_ARRAY,convertedtype=UTF8"`
	Height int64  `parquet:"name=height,type=INT64"`

	// Status tracking
	Status string `parquet:"name=status,type=BYTE_ARRAY,convertedtype=UTF8"`

	// Timing information (nanoseconds since epoch)
	StartTime    int64 `parquet:"name=start_time,type=INT64"`
	StoreTime    int64 `parquet:"name=store_time,type=INT64"`
	MetadataTime int64 `parquet:"name=metadata_time,type=INT64"`
	BlobpoolTime int64 `parquet:"name=blobpool_time,type=INT64"`
	// ProposalTime   int64 `parquet:"name=proposal_time,type=INT64"`
	// ValidationTime int64 `parquet:"name=validation_time,type=INT64"`
	FinalizedTime int64 `parquet:"name=finalized_time,type=INT64"`
}

// FromTrackedBlob converts a TrackedBlob to a BlobProfileRecord
func FromTrackedBlob(blob *types.TrackedBlob, profileStartTime time.Time) *BlobProfileRecord {
	record := &BlobProfileRecord{
		Count:  int64(blob.Nonce),
		Hash:   blob.Key.String(), // Using Key as hash since it's derived from data
		Size:   int32(blob.Size),
		Status: blob.Submission.Status.String(),
	}

	// Set transaction hash if available
	if blob.MetadataTx != nil {
		record.TxHash = blob.MetadataTx.Hash.String()
		record.Height = blob.MetadataTx.Height
	}

	// Set sequence if available from submission
	if blob.Submission != nil {
		// Convert times to nanoseconds since epoch
		record.StartTime = blob.Submission.StartTime.UnixNano()
		record.StoreTime = blob.Submission.StoreTime.UnixNano()
		record.MetadataTime = blob.Submission.MetadataTime.UnixNano()
		record.BlobpoolTime = blob.Submission.BlobpoolTime.UnixNano()
		// record.ProposalTime = blob.Submission.ProposalTime.UnixNano()
		// record.ValidationTime = blob.Submission.ValidationTime.UnixNano()
		record.FinalizedTime = blob.Submission.FinalizedTime.UnixNano()
	}

	return record
}

func writeBlobs(blobs []*types.TrackedBlob) func(
	buffer []*BlobProfileRecord, profileStartTime time.Time) []*BlobProfileRecord {
	return func(buffer []*BlobProfileRecord, profileStartTime time.Time) []*BlobProfileRecord {
		for _, blob := range blobs {
			record := FromTrackedBlob(blob, profileStartTime)
			buffer = append(buffer, record)
		}

		return buffer
	}
}
