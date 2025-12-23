package keeper

import (
	"context"
	"strings"
	"time"

	"cosmossdk.io/log"
	blobclient "github.com/fuel-infrastructure/blob-storage/pkg/client"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
)

type blobhubClient struct {
	logger         log.Logger
	blobpool       *Blobpool // reference to the blobpool, where blobs are stored
	blobhubAddress string    // configurable blobhub address (HTTP base URL)

	client *blobclient.Client
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

func newBlobhubClient(ctx context.Context, logger log.Logger, blobpool *Blobpool, blobhubAddress string) (*blobhubClient, error) {
	// Normalize address to HTTP URL format
	httpURL := normalizeBlobhubAddress(blobhubAddress)

	// Create client with HTTP base URL
	// Note: Logger is optional - client library will use no-op logger if nil
	config := &blobclient.ClientConfig{
		BaseURL: httpURL,
		Logger:  nil, // Use no-op logger from client library
	}
	client := blobclient.NewClient(config)

	blobhubClient := &blobhubClient{
		logger:         logger.With("module", "blobhub_sync"),
		blobpool:       blobpool,
		blobhubAddress: httpURL,
		client:         client,
	}

	blobhubClient.logger.Info("created new blobhub client", "url", httpURL)
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
				}
			}
		}
	}
}
