package keeper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"cosmossdk.io/log"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlobpool(t *testing.T) {
	ctx := context.TODO()
	logger := log.NewTestLogger(t)
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress)
	require.NoError(t, err)

	require.NotNil(t, pool)
	require.NotNil(t, pool.logger)
	// Test that blobs map exists by attempting an operation on it
	exists, err := pool.store.Has(ctx, store.Key{})
	require.False(t, exists)
	require.NoError(t, err)
}

func TestBlobpool_HasBlob(t *testing.T) {
	ctx := context.TODO()
	logger := log.NewTestLogger(t)
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress)
	require.NoError(t, err)

	data := []byte(fmt.Sprintf("test blobpool has data-%s", time.Now().Format(time.RFC3339)))
	key := store.NewKey(data)

	// Test non-existent blob
	assert.False(t, pool.Has(ctx, key))

	// Test existing blob
	pool.Insert(ctx, data)
	assert.True(t, pool.Has(ctx, key))
}

func TestBlobpool_GetBlob(t *testing.T) {
	ctx := context.TODO()
	logger := log.NewTestLogger(t)
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress)
	require.NoError(t, err)

	// Test getting non-existent blob
	key := store.Key{0x1, 0x2, 0x3}
	blob, err := pool.Get(ctx, key)
	assert.Error(t, err)
	assert.Nil(t, blob)

	// Test getting existing blob
	data := []byte("test data")
	key = store.NewKey(data)
	pool.Insert(ctx, data)

	blob, err = pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, key, blob.Key)
	assert.Equal(t, data, blob.Data)
}

func TestBlobpool_StoreBlob(t *testing.T) {
	ctx := context.TODO()
	logger := log.NewTestLogger(t)
	pool, err := newBlobpool(ctx, logger, testBlobpoolRedisAddress)
	require.NoError(t, err)

	data := []byte("test data")
	key := store.NewKey(data)

	// Store blob
	pool.Insert(ctx, data)

	// Verify blob was stored correctly
	verifyBlob := func() *store.StoredBlob {
		storedData, err := pool.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, key, storedData.Key)
		assert.Equal(t, data, storedData.Data)
		return storedData
	}
	storedData := verifyBlob()

	// Store another blob (should not overwrite)
	newData := []byte("new test data")
	newKey := store.NewKey(newData)
	pool.Insert(ctx, newData)

	// Verify blob was not overwritten
	newStoredData, err := pool.Get(ctx, newKey)
	require.NoError(t, err)
	assert.Equal(t, newKey, newStoredData.Key)
	assert.Equal(t, newData, newStoredData.Data)
	assert.NotEqual(t, newKey, storedData.Key)
	assert.NotEqual(t, newData, storedData.Data)

	// Re-verify original blob is still stored
	verifyBlob()
}
