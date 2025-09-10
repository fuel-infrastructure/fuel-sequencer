package types

import (
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
)

// Status represents the current status of a blob submission
type Status int

const (
	Pending Status = iota
	Stored
	Submitted
	InBlobpool
	Proposed
	Finalized
	Failed
)

func (s Status) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Stored:
		return "Stored"
	case Submitted:
		return "Submitted"
	case InBlobpool:
		return "InBlobpool"
	case Proposed:
		return "Proposed"
	case Finalized:
		return "Finalized"
	case Failed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// TrackedBlob represents a blob with its metadata
type TrackedBlob struct {
	Submission   *Submission
	Size         int
	MetadataTx   *coretypes.ResultTx
	MetadataHash string
	store.StoredBlob
}

// Submission tracks the lifecycle of a single blob submission
type Submission struct {
	Status         `json:"status"`
	StartTime      time.Time `json:"generate_time"`   // T0: Start of blob being submitted
	StoreTime      time.Time `json:"store_time"`      // T1: Stored on Blobhub
	MetadataTime   time.Time `json:"metadata_time"`   // T2: Metadata submitted to sequencer
	BlobpoolTime   time.Time `json:"blobpool_time"`   // T3: Blob stored in Blobpool
	ProposalTime   time.Time `json:"proposal_time"`   // T4: Blob Metadata included in Proposal
	ValidationTime time.Time `json:"validation_time"` // T5: Blob Metadata being validated
	FinalizedTime  time.Time `json:"finalized_time"`  // T6: Blob metadata included in finalized block
}
