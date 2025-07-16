package types

const (
	// ModuleName defines the module name
	ModuleName = "blob"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_blob"

	// TransientStoreKey defines the transient store key for blob data during block execution
	TransientStoreKey = "transient_blob"

	// BlobMetadataKey is the prefix to retrieve all BlobMetadata
	BlobMetadataKey = "BlobMetadata/value/"
)

var (
	ParamsKey = []byte("p_blob")
)

// KeyPrefix returns the key prefix for a given string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

// BlobMetadataKeyPrefix returns the store key to retrieve blob metadata from the blob hash
func BlobMetadataKeyPrefix(blobHash []byte) []byte {
	return append([]byte(BlobMetadataKey), blobHash...)
}
