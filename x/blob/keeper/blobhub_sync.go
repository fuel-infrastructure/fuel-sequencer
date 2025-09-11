package keeper

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"time"

	"cosmossdk.io/log"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/gorilla/websocket"

	"github.com/fuel-infrastructure/fuel-sequencer/x/blob/metrics"
)

// Default blobhub address - can be overridden for testing
var BlobhubAddress = "localhost:31035"

type blobMessage struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Data      string `json:"data"`
	Timestamp int64  `json:"timestamp"`
	Size      int    `json:"size"`
}

type blobhubClient struct {
	logger   log.Logger
	blobpool *Blobpool // reference to the blobpool, where blobs are stored

	conn *websocket.Conn

	blobs chan store.StoredBlob
}

func newBlobhubClient(ctx context.Context, logger log.Logger, blobpool *Blobpool) (*blobhubClient, error) {
	client := &blobhubClient{
		logger:   logger.With("module", "blobhub_sync"),
		blobpool: blobpool,
		blobs:    make(chan store.StoredBlob),
	}

	if err := client.connect(ctx); err != nil {
		client.logger.Error("failed to create blobhub client", "error", err)
		return nil, err
	}

	client.logger.Info("created new blobhub client")
	go client.sync(ctx)

	return client, nil
}

func (c *blobhubClient) connect(ctx context.Context) error {
	u := url.URL{Scheme: "ws", Host: BlobhubAddress, Path: "/stream"}
	c.logger.Info("connecting to blobhub", "url", u.String())

	start := time.Now()
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		c.logger.Error("failed to connect to blobhub", "error", err, "url", u.String())
		metrics.IncrementBlobhubErrors()
		return fmt.Errorf("failed to connect to blobhub: %w", err)
	}

	c.conn = conn

	// Record metrics
	connectionTime := time.Since(start)
	metrics.SetBlobhubConnectionStatus(true)
	metrics.ObserveBlobSyncLatency(connectionTime)

	c.logger.Info("connected to blobhub successfully")
	return nil
}

func (c *blobhubClient) sync(ctx context.Context) {
	defer func() {
		c.logger.Info("closing blobhub connection")
		metrics.SetBlobhubConnectionStatus(false)
		c.conn.Close()
	}()

	c.logger.Info("starting blob sync")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("context cancelled, stopping sync")
			return
		default:
			var msg blobMessage
			err := c.conn.ReadJSON(&msg)
			if err != nil {
				// Handle connection errors
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("unexpected websocket close", "error", err)
					metrics.SetBlobhubConnectionStatus(false)
					// Try to reconnect
					if err := c.connect(ctx); err != nil {
						c.logger.Error("failed to reconnect", "error", err)
						metrics.IncrementBlobhubErrors()
						continue
					}
					metrics.IncrementBlobhubReconnections()
					c.logger.Info("reconnected successfully")
				} else {
					c.logger.Debug("websocket read error", "error", err)
				}
				continue
			}

			// Decode base64 blob data
			data, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				c.logger.Error("failed to decode blob data", "error", err, "id", msg.ID)
				continue
			}

			// Parse blob key
			key, err := store.ParseKey(msg.ID)
			if err != nil {
				c.logger.Error("failed to parse blob key", "error", err, "id", msg.ID)
				continue
			}

			// Skip if we already have this blob
			if c.blobpool.Has(ctx, key) {
				c.logger.Debug("skipping existing blob", "id", msg.ID)
				continue
			}

			// Store the blob
			c.logger.Info("storing new blob", "id", msg.ID, "size", msg.Size)
			c.blobpool.Insert(ctx, data)
		}
	}
}
