package config

import (
	"errors"
	"time"

	"github.com/fuel-infrastructure/blob-storage/pkg/blobgen"
)

type ProfileRate struct {
	Rate    func(since time.Duration) int64 `json:"rate_increment"` // bytes/sec
	MaxRate int64                           `json:"max_rate"`       // bytes/sec
}

// Config holds the profiler configuration
type Config struct {
	BlobhubURL       string `json:"blobhub_url"`
	SequencerGRPC    string `json:"sequencer_grpc"`
	ProfileRate      `json:"profile_rate"`
	BlobDistribution blobgen.BlobSizeDistribution `json:"blob_distribution"`
	MaxLatency       time.Duration                `json:"max_latency"`
	Topic            string                       `json:"topic"`
	Sender           string                       `json:"sender"`
}

// Configuration validation errors
var (
	ErrEmptyBlobhubURL    = errors.New("blobhub URL cannot be empty")
	ErrEmptySequencerGRPC = errors.New("sequencer gRPC address cannot be empty")
	ErrNilRateFunction    = errors.New("invalid rate function: must be non-nil")
	ErrInvalidMaxRate     = errors.New("invalid max rate: must be positive")
	ErrInvalidMaxLatency  = errors.New("invalid max latency: must be positive")
)

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.BlobhubURL == "" {
		return ErrEmptyBlobhubURL
	}
	if c.SequencerGRPC == "" {
		return ErrEmptySequencerGRPC
	}
	if c.Rate == nil {
		return ErrNilRateFunction
	}
	if c.MaxRate <= 0 {
		return ErrInvalidMaxRate
	}
	if err := blobgen.ValidateDistribution(c.BlobDistribution); err != nil {
		return err
	}
	if c.MaxLatency <= 0 {
		return ErrInvalidMaxLatency
	}
	return nil
}
