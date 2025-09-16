package report

import (
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
)

// EventType represents the type of event being recorded
type EventType string

const (
	EventTypeError      EventType = "error"
	EventTypeCasting    EventType = "casting"
	EventTypeCatching   EventType = "catching"
	EventTypeConnection EventType = "connection"
	EventTypeParquet    EventType = "parquet"
	EventTypeSequencer  EventType = "sequencer"
	EventTypeBlobpool   EventType = "blobpool"
	EventTypeInfo       EventType = "info"
	EventTypeWarning    EventType = "warning"
)

// Event represents a single event in the profiler report
type Event struct {
	Timestamp  time.Time              `json:"timestamp"`
	EventType  EventType              `json:"event_type"`
	Component  string                 `json:"component"`
	Message    string                 `json:"message"`
	Error      string                 `json:"error,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
}

// SerializableProfile represents a JSON-safe version of the profile configuration
type SerializableProfile struct {
	Description   string        `json:"description"`
	Duration      time.Duration `json:"duration"`
	DurationHuman string        `json:"duration_human"`
	MaxRate       int           `json:"max_rate"`
	MaxRateHuman  string        `json:"max_rate_human"`
	Type          string        `json:"profile_type"`
	Purpose       string        `json:"profile_purpose"`
	Explanation   string        `json:"profile_explanation"`
	BlobSizeInfo  string        `json:"blob_size_info"`
	// Note: Rate and Size functions are omitted as they cannot be serialized
}

// SerializableConfig represents a JSON-safe version of the configuration
type SerializableConfig struct {
	BlobhubURL            string                       `json:"blobhub_url"`
	SequencerRPC          string                       `json:"sequencer_rpc"`
	BlobpoolURL           string                       `json:"blobpool_url"`
	ParquetDir            string                       `json:"parquet_dir"`
	Profile               SerializableProfile          `json:"profile"`
	BlobDistribution      blobgen.BlobSizeDistribution `json:"blob_distribution"`
	BlobDistributionHuman string                       `json:"blob_distribution_human"`
	MaxLagRatio           float32                      `json:"max_lag_ratio"`
	LagTolerance          time.Duration                `json:"lag_tolerance"`
	LagToleranceHuman     string                       `json:"lag_tolerance_human"`
	BufferDuration        time.Duration                `json:"buffer_duration"`
	BufferDurationHuman   string                       `json:"buffer_duration_human"`
	Topic                 string                       `json:"topic"`
	Sender                string                       `json:"sender"`
}

// ProfilerEnvironment contains all environment and setup information
type ProfilerEnvironment struct {
	Timestamp       time.Time              `json:"timestamp"`
	CommitHash      string                 `json:"commit_hash"`
	CommitHashShort string                 `json:"commit_hash_short"`
	GitBranch       string                 `json:"git_branch"`
	GitStatus       string                 `json:"git_status"`
	GoVersion       string                 `json:"go_version"`
	OS              string                 `json:"os"`
	Architecture    string                 `json:"architecture"`
	Config          SerializableConfig     `json:"config"`
	RuntimeInfo     map[string]interface{} `json:"runtime_info"`
	EnvironmentVars map[string]string      `json:"environment_vars,omitempty"`
}

// ProfilerSummary contains summary statistics
type ProfilerSummary struct {
	TotalBlobs                 int           `json:"total_blobs"`
	SuccessfulBlobs            int           `json:"successful_blobs"`
	FailedBlobs                int           `json:"failed_blobs"`
	TotalEvents                int           `json:"total_events"`
	ErrorCount                 int           `json:"error_count"`
	WarningCount               int           `json:"warning_count"`
	ProfileDuration            time.Duration `json:"profile_duration"`
	ProfileDurationHuman       string        `json:"profile_duration_human"`
	StartTime                  time.Time     `json:"start_time"`
	EndTime                    time.Time     `json:"end_time"`
	DataSubmittedBytes         int           `json:"data_submitted_bytes"`
	DataSubmittedKiB           int           `json:"data_submitted_kib"`
	AverageThroughput          float64       `json:"average_throughput_bytes_per_sec"`
	AverageThroughputKiBPerSec float64       `json:"average_throughput_kib_per_sec"`
}

// ProfilerReportData contains the complete profiling report data
type ProfilerReportData struct {
	Environment ProfilerEnvironment `json:"environment"`
	Events      []Event             `json:"events"`
	Summary     ProfilerSummary     `json:"summary"`
}
