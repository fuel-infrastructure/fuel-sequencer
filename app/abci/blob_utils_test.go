package abci_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

func TestBlobValidationIntegration(t *testing.T) {
	// Create test keeper directly
	k, _ := keepertest.BlobKeeper(t)

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			StoredAt: time.Now(),
			Key:      key,
		},
		Data: data,
	}

	invalidKey := store.NewKey([]byte("invalid-hash"))

	// Store blob in keeper
	k.Insert(blob)

	// Test that blob is available
	require.True(t, k.Has(key))
	require.False(t, k.Has(invalidKey))

	// Test blob data retrieval
	retrieved, err := k.Get(key)
	require.NoError(t, err)
	require.Equal(t, data, retrieved.Data)

	// Test that we can create blob messages
	msg := &blobtypes.MsgBlobMetadataTx{
		Hash:  key.String(),
		Size_: 100,
		Topic: "test-topic",
		Nonce: 1,
	}
	require.NotNil(t, msg)
	require.Equal(t, key.String(), msg.Hash)
}

func TestBlobHashVerification(t *testing.T) {
	k, _ := keepertest.BlobKeeper(t)

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			StoredAt: time.Now(),
			Key:      key,
		},
		Data: data,
	}

	// Store blob in keeper
	k.Insert(blob)

	// Test hash verification
	keyAgain := store.NewKey(data)
	require.Equal(t, key, keyAgain)

	// Test with wrong data
	wrongData := []byte("wrong data")
	wrongKey := store.NewKey(wrongData)
	require.NotEqual(t, key, wrongKey)
}

func TestBlobPoolOperations(t *testing.T) {
	k, _ := keepertest.BlobKeeper(t)

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			StoredAt: time.Now(),
			Key:      key,
		},
		Data: data,
	}

	// Initially, blob should not be available
	require.False(t, k.Has(key))

	// Store blob in keeper
	k.Insert(blob)

	// Now blob should be available
	require.True(t, k.Has(key))

	// Test retrieval
	retrieved, err := k.Get(key)
	require.NoError(t, err)
	require.Equal(t, data, retrieved.Data)

	// Test retrieval of non-existent blob
	_, err = k.Get(store.NewKey([]byte("non-existent")))
	require.Error(t, err)
}

func TestBlobMessageTypes(t *testing.T) {
	// Test that we can create blob metadata messages
	data := []byte("test data")
	key := store.NewKey(data)

	msg := &blobtypes.MsgBlobMetadataTx{
		Hash:  key.String(),
		Size_: 100,
		Topic: "test-topic",
		Nonce: 1,
	}

	require.NotNil(t, msg)
	require.Equal(t, key.String(), msg.Hash)
	require.Equal(t, uint64(100), msg.Size_)
	require.Equal(t, "test-topic", msg.Topic)
	require.Equal(t, uint64(1), msg.Nonce)
}

func TestBlobKeeperAuthority(t *testing.T) {
	k, _ := keepertest.BlobKeeper(t)

	// Test that the blob keeper has the correct authority
	authority := k.GetAuthority()
	require.NotEmpty(t, authority)
	require.Contains(t, authority, "fuelsequencer")
}

func TestBlobValidationLogic(t *testing.T) {
	k, _ := keepertest.BlobKeeper(t)

	// Test the validation logic that would be used in ABCI handlers
	data := []byte("test blob data")
	key := store.NewKey(data)
	blob := &store.StoredBlob{
		Receipt: store.Receipt{
			StoredAt: time.Now(),
			Key:      key,
		},
		Data: data,
	}

	// Initially, blob should not be available for validation
	require.False(t, k.Has(key))

	// Store blob
	k.Insert(blob)

	// Now it should be available
	require.True(t, k.Has(key))

	// Test hash verification logic
	retrieved, err := k.Get(key)
	require.NoError(t, err)

	retrievedHash := store.NewKey(retrieved.Data)
	require.Equal(t, key, retrievedHash)

	// Test with wrong hash
	wrongKey := store.NewKey([]byte("wrong data"))
	require.NotEqual(t, key, wrongKey)
}
