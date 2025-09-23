package keeper

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// mockBlobhubServer creates a test server that simulates the blobhub websocket server
func mockBlobhubServer(t *testing.T) (*httptest.Server, chan store.StoredBlob) {
	blobChan := make(chan store.StoredBlob, 10)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/stream") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		// Upgrade connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Failed to upgrade connection: %v", err)
			return
		}
		defer conn.Close()

		// Send blobs received on the channel to the websocket client
		for blob := range blobChan {
			msg := blobMessage{
				Type:      "blob",
				ID:        blob.Key.String(),
				Data:      base64.StdEncoding.EncodeToString(blob.Data),
				Timestamp: blob.StoredAt.Unix(),
				Size:      len(blob.Data),
			}

			if err := conn.WriteJSON(msg); err != nil {
				t.Errorf("Failed to write message: %v", err)
				return
			}
		}
	}))

	return server, blobChan
}

func TestBlobhubClient_Connect(t *testing.T) {
	logger := log.NewTestLogger(t)
	ctx := context.Background()
	pool, err := newBlobpool(ctx, logger)
	require.NoError(t, err)

	// Start mock server
	server, _ := mockBlobhubServer(t)
	defer server.Close()

	// Override default blobhub address with mock server
	origAddr := BlobhubAddress
	BlobhubAddress = strings.TrimPrefix(server.URL, "http://")
	defer func() { BlobhubAddress = origAddr }()

	// Test successful connection
	client, err := newBlobhubClient(ctx, logger, pool)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.conn)

	// Test connection to invalid address
	BlobhubAddress = "invalid:1234"
	_, err = newBlobhubClient(ctx, logger, pool)
	assert.Error(t, err)
}

func TestBlobhubClient_Sync(t *testing.T) {
	logger := log.NewTestLogger(t)
	ctx := context.Background()
	pool, err := newBlobpool(ctx, logger)
	require.NoError(t, err)

	// Start mock server
	server, blobChan := mockBlobhubServer(t)
	defer server.Close()

	// Override default blobhub address with mock server
	origAddr := BlobhubAddress
	BlobhubAddress = strings.TrimPrefix(server.URL, "http://")
	defer func() { BlobhubAddress = origAddr }()

	// Create client
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := newBlobhubClient(ctx, logger, pool)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test receiving a blob
	data := []byte("test data")
	key := store.NewKey(data)
	blob := store.StoredBlob{
		Receipt: store.Receipt{
			Key:      key,
			StoredAt: time.Now(),
		},
		Data: data,
	}

	// Send blob through mock server
	blobChan <- blob

	// Wait for blob to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify blob was stored in pool
	assert.True(t, pool.Has(ctx, key))
	retrieved, err := pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, blob.Key, retrieved.Key)
	assert.Equal(t, blob.Data, retrieved.Data)

	// Test duplicate blob
	blobChan <- blob
	time.Sleep(100 * time.Millisecond)

	// Verify blob is still stored correctly
	retrieved, err = pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, blob.Data, retrieved.Data)

	// Test invalid blob data
	invalidMsg := blobMessage{
		Type:      "blob",
		ID:        "invalid",
		Data:      "invalid base64",
		Timestamp: time.Now().Unix(),
		Size:      0,
	}
	require.NoError(t, client.conn.WriteJSON(invalidMsg))
	time.Sleep(100 * time.Millisecond)

	// Test context cancellation
	cancel()
	close(blobChan)                    // Close the blob channel to stop the mock server
	time.Sleep(200 * time.Millisecond) // Wait for goroutines to clean up

	// Try to write to the closed connection - should fail
	err = client.conn.WriteMessage(websocket.TextMessage, []byte("test"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}
