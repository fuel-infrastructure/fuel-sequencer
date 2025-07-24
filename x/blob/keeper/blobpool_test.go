package keeper

import (
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

	// Test non-existent blob
	key := store.Key{0x1, 0x2, 0x3}
	assert.False(t, pool.hasBlob(key))

	// Test existing blob
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: []byte("test data"),
	}
	pool.storeBlob(blob)
	assert.True(t, pool.hasBlob(key))
}

func TestBlobpool_GetBlob(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	// Test getting non-existent blob
	key := store.Key{0x1, 0x2, 0x3}
	data, err := pool.getBlob(key)
	assert.Error(t, err)
	assert.Nil(t, data)

	// Test getting existing blob
	testData := []byte("test data")
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: testData,
	}
	pool.storeBlob(blob)

	data, err = pool.getBlob(key)
	require.NoError(t, err)
	assert.Equal(t, testData, data)
}

func TestBlobpool_StoreBlob(t *testing.T) {
	logger := log.NewTestLogger(t)
	pool := newBlobpool(logger)

	key := store.Key{0x1, 0x2, 0x3}
	testData := []byte("test data")
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: testData,
	}

	// Store blob
	pool.storeBlob(blob)

	// Verify blob was stored correctly
	storedData, err := pool.getBlob(key)
	require.NoError(t, err)
	assert.Equal(t, testData, storedData)

	// Store another blob with same key (should overwrite)
	newData := []byte("new test data")
	newBlob := &store.StoredBlob{
		Receipt: store.Receipt{
			Key: key,
		},
		Data: newData,
	}
	pool.storeBlob(newBlob)

	// Verify blob was overwritten
	storedData, err = pool.getBlob(key)
	require.NoError(t, err)
	assert.Equal(t, newData, storedData)
}
