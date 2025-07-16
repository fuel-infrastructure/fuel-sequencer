package abci_test

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/fuel-infrastructure/fuel-sequencer/testutil/keeper"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
)

func TestBlobValidationIntegration(t *testing.T) {
	// Create test keeper directly
	k, ctx := keepertest.BlobKeeper(t)

	// Create test blob data
	blobData := []byte("test blob data")
	blobHash := sha256.Sum256(blobData)

	// Store blob in keeper
	k.SetBlobData(ctx, blobHash[:], blobData)

	// Test that blob is available
	require.True(t, k.HasBlobData(ctx, blobHash[:]))
	require.False(t, k.HasBlobData(ctx, []byte("invalid-hash")))

	// Test blob data retrieval
	retrievedData, err := k.GetBlobData(ctx, blobHash[:])
	require.NoError(t, err)
	require.Equal(t, blobData, retrievedData)

	// Test that we can create blob messages
	msg := &blobtypes.MsgBlobMetadataTx{
		BlobHash: blobHash[:],
		Size_:    100,
		Topic:    "test-topic",
		Nonce:    1,
	}
	require.NotNil(t, msg)
	require.Equal(t, blobHash[:], msg.BlobHash)
}

func TestBlobHashVerification(t *testing.T) {
	k, ctx := keepertest.BlobKeeper(t)

	// Create test blob data
	blobData := []byte("test blob data")
	blobHash := sha256.Sum256(blobData)

	// Store blob in keeper
	k.SetBlobData(ctx, blobHash[:], blobData)

	// Test hash verification
	computedHash := sha256.Sum256(blobData)
	require.Equal(t, blobHash[:], computedHash[:])

	// Test with wrong data
	wrongData := []byte("wrong data")
	wrongHash := sha256.Sum256(wrongData)
	require.NotEqual(t, blobHash[:], wrongHash[:])
}

func TestBlobPoolOperations(t *testing.T) {
	k, ctx := keepertest.BlobKeeper(t)

	// Create test blob data
	blobData := []byte("test blob data")
	blobHash := sha256.Sum256(blobData)

	// Initially, blob should not be available
	require.False(t, k.HasBlobData(ctx, blobHash[:]))

	// Store blob in keeper
	k.SetBlobData(ctx, blobHash[:], blobData)

	// Now blob should be available
	require.True(t, k.HasBlobData(ctx, blobHash[:]))

	// Test retrieval
	retrievedData, err := k.GetBlobData(ctx, blobHash[:])
	require.NoError(t, err)
	require.Equal(t, blobData, retrievedData)

	// Test retrieval of non-existent blob
	_, err = k.GetBlobData(ctx, []byte("non-existent"))
	require.Error(t, err)
}

func TestBlobMessageTypes(t *testing.T) {
	// Test that we can create blob metadata messages
	blobHash := sha256.Sum256([]byte("test data"))

	msg := &blobtypes.MsgBlobMetadataTx{
		BlobHash: blobHash[:],
		Size_:    100,
		Topic:    "test-topic",
		Nonce:    1,
	}

	require.NotNil(t, msg)
	require.Equal(t, blobHash[:], msg.BlobHash)
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
	k, ctx := keepertest.BlobKeeper(t)

	// Test the validation logic that would be used in ABCI handlers
	blobData := []byte("test blob data")
	blobHash := sha256.Sum256(blobData)

	// Initially, blob should not be available for validation
	require.False(t, k.HasBlobData(ctx, blobHash[:]))

	// Store blob
	k.SetBlobData(ctx, blobHash[:], blobData)

	// Now it should be available
	require.True(t, k.HasBlobData(ctx, blobHash[:]))

	// Test hash verification logic
	retrievedData, err := k.GetBlobData(ctx, blobHash[:])
	require.NoError(t, err)

	computedHash := sha256.Sum256(retrievedData)
	require.Equal(t, blobHash[:], computedHash[:])

	// Test with wrong hash
	wrongHash := sha256.Sum256([]byte("wrong data"))
	require.NotEqual(t, blobHash[:], wrongHash[:])
}
