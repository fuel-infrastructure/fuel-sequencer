package keeper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	blobclient "github.com/fuel-infrastructure/blob-storage/pkg/client"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testBlobhubAddress       = "localhost:31035"
	testBlobpoolRedisAddress = "localhost:6380"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// mockBlobhubServer creates a test server that simulates the blobhub server
// It handles both WebSocket /stream endpoint and HTTP GET /get/{key} endpoint
func mockBlobhubServer(t *testing.T) (*httptest.Server, chan store.StoredBlob) {
	blobChan := make(chan store.StoredBlob, 10)
	// Store blobs in a map so they can be retrieved via HTTP GET
	blobStore := make(map[string]store.StoredBlob)
	var mu sync.RWMutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle HTTP GET /get/{key} endpoint
		if strings.HasPrefix(r.URL.Path, "/get/") && r.Method == http.MethodGet {
			keyStr := strings.TrimPrefix(r.URL.Path, "/get/")
			mu.RLock()
			blob, exists := blobStore[keyStr]
			mu.RUnlock()

			if !exists {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}

			// Set X-Stored-At header
			w.Header().Set("X-Stored-At", blob.StoredAt.Format(time.RFC3339))
			w.Header().Set("Content-Type", "application/octet-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(blob.Data)
			return
		}

		// Handle WebSocket /stream endpoint
		if strings.HasSuffix(r.URL.Path, "/stream") {
			// Upgrade connection to WebSocket
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				t.Errorf("Failed to upgrade connection: %v", err)
				return
			}
			defer conn.Close()

			// Send blob_notification messages for blobs received on the channel
			for blob := range blobChan {
				// Store blob for HTTP GET retrieval
				mu.Lock()
				blobStore[blob.Key.String()] = blob
				mu.Unlock()

				// Send blob_notification (without data)
				msg := blobclient.BlobNotification{
					Type:      "blob_notification",
					ID:        blob.Key.String(),
					Timestamp: blob.StoredAt.Unix(),
					Size:      len(blob.Data),
				}

				if err := conn.WriteJSON(msg); err != nil {
					t.Errorf("Failed to write message: %v", err)
					return
				}
			}
			return
		}

		// Unknown endpoint
		http.Error(w, "not found", http.StatusNotFound)
	}))

	return server, blobChan
}

func TestBlobhubClient_Connect(t *testing.T) {
	logger := log.NewTestLogger(t)
	ctx := context.Background()
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress, false)
	require.NoError(t, err)

	// Start mock server
	server, _ := mockBlobhubServer(t)
	defer server.Close()

	// Override default blobhub address with mock server
	origAddr := testBlobhubAddress
	testBlobhubAddress = strings.TrimPrefix(server.URL, "http://")
	defer func() { testBlobhubAddress = origAddr }()

	// Test successful connection
	client, err := newBlobhubClient(ctx, logger, pool, testBlobhubAddress)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.client)

	// Test connection to invalid address
	testBlobhubAddress = "invalid:1234"
	_, err = newBlobhubClient(ctx, logger, pool, testBlobhubAddress)
	// Note: The client creation itself succeeds, but connection will fail during sync
	// The error will be logged but won't fail client creation
	assert.NoError(t, err)
}

func TestBlobhubClient_Sync(t *testing.T) {
	logger := log.NewTestLogger(t)
	ctx := context.Background()
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress, false)
	require.NoError(t, err)

	// Start mock server
	server, blobChan := mockBlobhubServer(t)
	defer server.Close()

	// Override default blobhub address with mock server
	origAddr := testBlobhubAddress
	testBlobhubAddress = strings.TrimPrefix(server.URL, "http://")
	defer func() { testBlobhubAddress = origAddr }()

	// Create client
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := newBlobhubClient(ctx, logger, pool, testBlobhubAddress)
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

	// Send blob through mock server (will trigger blob_notification via WebSocket)
	blobChan <- blob

	// Wait for blob to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify blob was stored in pool
	require.True(t, pool.Has(ctx, key))
	retrieved, err := pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, blob.Key, retrieved.Key)
	assert.Equal(t, blob.Data, retrieved.Data)

	// Test duplicate blob (should be skipped - already exists)
	blobChan <- blob

	// Verify blob is still stored correctly
	retrieved, err = pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, blob.Data, retrieved.Data)

	// Test invalid blob key (notification with invalid key)
	invalidBlob := store.StoredBlob{
		Receipt: store.Receipt{
			Key:      store.Key{}, // Invalid/empty key
			StoredAt: time.Now(),
		},
		Data: []byte("invalid"),
	}
	blobChan <- invalidBlob // Invalid keys are skipped

	// Test context cancellation
	cancel()
	close(blobChan) // Close the blob channel to stop the mock server
}
