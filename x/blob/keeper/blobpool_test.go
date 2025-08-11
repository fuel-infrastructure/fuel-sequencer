package keeper

import (
	"context"
	"testing"

	"cosmossdk.io/log"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlobpool(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	require.NotNil(t, pool)
	require.NotNil(t, pool.logger)
	// Test that blobs map exists by attempting an operation on it
	_, exists := pool.blobs.Load(store.Key{})
	require.False(t, exists) // Should not exist, but operation should work
}

func TestBlobpool_HasBlob(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	ctx := context.Background()

	// Test non-existent blob
	key := store.Key{0x1, 0x2, 0x3}
	assert.False(t, pool.Has(ctx, key))

	// Test existing blob
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: []byte("test data"),
	}
	pool.Insert(ctx, blob)
	assert.True(t, pool.Has(ctx, key))
}

func TestBlobpool_GetBlob(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	ctx := context.Background()

	// Test getting non-existent blob
	key := store.Key{0x1, 0x2, 0x3}
	blob, err := pool.Get(ctx, key)
	assert.Error(t, err)
	assert.Nil(t, blob)

	// Test getting existing blob
	testData := []byte("test data")
	expected := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: testData,
	}
	pool.Insert(ctx, expected)

	blob, err = pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, expected.Key, blob.Key)
	assert.Equal(t, expected.Data, blob.Data)
}

func TestBlobpool_StoreBlob(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	ctx := context.Background()

	key := store.Key{0x1, 0x2, 0x3}
	testData := []byte("test data")
	expected := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: testData,
	}

	// Store blob
	pool.Insert(ctx, expected)

	// Verify blob was stored correctly
	storedData, err := pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, expected.Key, storedData.Key)
	assert.Equal(t, expected.Data, storedData.Data)

	// Store another blob with same key (should overwrite)
	newData := []byte("new test data")
	newBlob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: newData,
	}
	pool.Insert(ctx, newBlob)

	// Verify blob was overwritten
	storedData, err = pool.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, newBlob.Key, storedData.Key)
	assert.Equal(t, newBlob.Data, storedData.Data)
}
