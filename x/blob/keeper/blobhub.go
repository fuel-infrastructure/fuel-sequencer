package keeper

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/log"
	blobhub "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
)

type blobhubClient struct {
	logger         log.Logger
	blobpool       *Blobpool // reference to the blobpool, where blobs are stored
	blobhubAddress string    // configurable blobhub address (HTTP base URL)

	client      *blobhub.Client
	validatorID string // validator ID derived from consensus key (empty for non-validator nodes)
}

// normalizeBlobhubAddress converts a blobhub address to HTTP URL format.
// If the address already contains a scheme (http:// or https://), it's returned as-is.
// Otherwise, it's assumed to be host:port and http:// is prepended.
func normalizeBlobhubAddress(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return address
	}
	return "http://" + address
}

// newBlobhubClient creates a blobhub client. If client is nil, a new one is created.
// If validatorID is provided, ACK (signature submission) will be performed after blob sync.
func newBlobhubClient(ctx context.Context, logger log.Logger, blobpool *Blobpool, blobhubAddress string, client *blobhub.Client, validatorID string) (*blobhubClient, error) {
	// Normalize address to HTTP URL format
	httpURL := normalizeBlobhubAddress(blobhubAddress)

	// Create client if not provided
	if client == nil {
		config := &blobhub.ClientConfig{
			BaseURL: httpURL,
			Logger:  nil, // Use no-op logger from client library
		}
		client = blobhub.NewClient(config)
	}

	blobhubClient := &blobhubClient{
		logger:         logger.With("module", "blobhub_sync"),
		blobpool:       blobpool,
		blobhubAddress: httpURL,
		client:         client,
		validatorID:    validatorID,
	}

	// Register validator with Blobhub before starting sync (required for signing)
	if validatorID != "" {
		if err := client.Register(ctx, validatorID); err != nil {
			blobhubClient.logger.Error("failed to register validator with Blobhub", "error", err, "validator_id", validatorID)
			return nil, fmt.Errorf("failed to register validator: %w", err)
		}
		blobhubClient.logger.Info("registered validator with Blobhub", "validator_id", validatorID)
	}

	blobhubClient.logger.Info("created new blobhub client", "url", httpURL, "validator_id", validatorID)
	go blobhubClient.sync(ctx)

	return blobhubClient, nil
}

func (c *blobhubClient) sync(ctx context.Context) {
	defer func() {
		c.logger.Info("closing blobhub connection")
		metrics.SetBlobhubConnectionStatus(false)
	}()

	c.logger.Info("starting blob sync", "url", c.blobhubAddress)

	// Retry loop for reconnection
	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping sync")
			return
		default:
			start := time.Now()

			// StreamBlobs establishes WebSocket connection and returns a channel
			blobChan, err := c.client.StreamBlobs(ctx)
			if err != nil {
				c.logger.Error("failed to connect to blobhub", "error", err, "url", c.blobhubAddress)
				metrics.IncrementBlobhubErrors()
				metrics.SetBlobhubConnectionStatus(false)

				// Wait before retrying
				select {
				case <-ctx.Done():
					return
				case <-time.After(5 * time.Second):
					continue
				}
			}

			// Record successful connection metrics
			connectionTime := time.Since(start)
			metrics.SetBlobhubConnectionStatus(true)
			metrics.ObserveBlobSyncLatency(connectionTime)
			c.logger.Info("connected to blobhub successfully")

			// Process blobs from the stream
			for {
				select {
				case <-ctx.Done():
					c.logger.Info("context cancelled, stopping sync")
					return
				case blob, ok := <-blobChan:
					if !ok {
						// Channel closed, connection lost
						c.logger.Warn("blob stream channel closed, reconnecting")
						metrics.SetBlobhubConnectionStatus(false)
						metrics.IncrementBlobhubReconnections()
						break // Break inner loop to retry connection
					}

					if blob == nil {
						c.logger.Debug("received nil blob, skipping")
						continue
					}

					// Skip if we already have this blob
					if c.blobpool.Has(ctx, blob.Key) {
						c.logger.Debug("skipping existing blob", "id", blob.Key.String())
						continue
					}

					// Store the blob
					c.logger.Info("storing new blob", "id", blob.Key.String(), "size", len(blob.Data))
					c.blobpool.Insert(ctx, blob.Data)

					// Submit signature (ACK) if validator ID is set
					if c.validatorID != "" {
						c.logger.Info("submitting signature for blob", "id", blob.Key.String(), "validator_id", c.validatorID)
						c.signBlobAsync(ctx, blob.Key)
					} else {
						c.logger.Info("skipping signature (non-validator node)", "id", blob.Key.String())
					}
				}
			}
		}
	}
}

// signBlob submits a signature for a blob to Blobhub.
// This method is called asynchronously and includes retry logic matching
// existing patterns in the sync loop.
func (c *blobhubClient) signBlob(ctx context.Context, key store.Key) error {
	start := time.Now()

	// Retry logic matching sync reconnection pattern
	maxRetries := 3
	retryDelay := 5 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug("retrying signature submission", "attempt", attempt+1, "key", key.String())
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		err := c.client.Sign(ctx, c.validatorID, key)
		if err == nil {
			// Success - record metrics
			latency := time.Since(start)
			metrics.IncrementACKSubmissions()
			metrics.ObserveACKLatency(latency)
			c.logger.Info("signature submitted to blobhub", "key", key.String(), "validator_id", c.validatorID, "latency_ms", latency.Milliseconds())
			return nil
		}

		lastErr = err
		c.logger.Warn("failed to submit signature", "error", err, "key", key.String(), "attempt", attempt+1)
		metrics.IncrementACKErrors()
	}

	// All retries failed
	c.logger.Error("failed to submit signature after retries", "error", lastErr, "key", key.String())
	return lastErr
}

// signBlobAsync submits a signature asynchronously in a goroutine.
// This ensures that signature submission doesn't block the sync flow.
func (c *blobhubClient) signBlobAsync(ctx context.Context, key store.Key) {
	go func() {
		if err := c.signBlob(ctx, key); err != nil {
			// Log error but don't block sync - errors are already logged in signBlob
			c.logger.Error("async signature submission failed", "error", err, "key", key.String())
		}
	}()
}
