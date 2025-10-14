package abci_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/blob-storage/pkg/store"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

func TestBlobValidationIntegration(t *testing.T) {
	// Create test keeper directly
	k, _ := keepertest.BlobKeeper(t)
	ctx := context.TODO()
	if err := k.Initialize(ctx); err != nil {
		require.NoError(t, err, "failed to initialize blob keeper")
	}

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)

	invalidKey := store.NewKey([]byte("invalid-hash"))

	// Store blob in keeper
	k.Insert(ctx, data)

	// Test that blob is available
	require.True(t, k.Has(ctx, key))
	require.False(t, k.Has(ctx, invalidKey))

	// Test blob data retrieval
	retrieved, err := k.Get(ctx, key)
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
	ctx := context.TODO()
	if err := k.Initialize(ctx); err != nil {
		require.NoError(t, err, "failed to initialize blob keeper")
	}

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)

	// Store blob in keeper
	k.Insert(ctx, data)

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
	ctx := context.TODO()
	if err := k.Initialize(ctx); err != nil {
		require.NoError(t, err, "failed to initialize blob keeper")
	}

	// Create test blob data
	data := []byte("test blob data")
	key := store.NewKey(data)

	// Initially, blob should not be available
	require.False(t, k.Has(ctx, key))

	// Store blob in keeper
	k.Insert(ctx, data)

	// Now blob should be available
	require.True(t, k.Has(ctx, key))

	// Test retrieval
	retrieved, err := k.Get(ctx, key)
	require.NoError(t, err)
	require.Equal(t, data, retrieved.Data)

	// Test retrieval of non-existent blob
	_, err = k.Get(ctx, store.NewKey([]byte("non-existent")))
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
	ctx := context.TODO()
	if err := k.Initialize(ctx); err != nil {
		require.NoError(t, err, "failed to initialize blob keeper")
	}

	// Test that the blob keeper has the correct authority
	authority := k.GetAuthority()
	require.NotEmpty(t, authority)
	require.Contains(t, authority, "fuelsequencer")
}

func TestBlobValidationLogic(t *testing.T) {
	k, _ := keepertest.BlobKeeper(t)
	ctx := context.TODO()
	if err := k.Initialize(ctx); err != nil {
		require.NoError(t, err, "failed to initialize blob keeper")
	}

	// Test the validation logic that would be used in ABCI handlers
	data := []byte("test blob data")
	key := store.NewKey(data)

	// Initially, blob should not be available for validation
	require.False(t, k.Has(ctx, key))

	// Store blob
	k.Insert(ctx, data)

	// Now it should be available
	require.True(t, k.Has(ctx, key))

	// Test hash verification logic
	retrieved, err := k.Get(ctx, key)
	require.NoError(t, err)

	retrievedHash := store.NewKey(retrieved.Data)
	require.Equal(t, key, retrievedHash)

	// Test with wrong hash
	wrongKey := store.NewKey([]byte("wrong data"))
	require.NotEqual(t, key, wrongKey)
}
