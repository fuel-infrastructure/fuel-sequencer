package types

const (
	// ModuleName defines the module name
	ModuleName = "sequencing"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sequencing"

	// TopicKeyPrefix is the prefix to retrieve all Topic
	TopicKeyPrefix = "Topic/value/"
)

var (
	ParamsKey = []byte("p_sequencing")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// TopicKey returns the store key to retrieve a Topic from the topic Id
func TopicKey(
	topicId []byte,
) []byte {
	return append([]byte(topicId), topicId...)
}
