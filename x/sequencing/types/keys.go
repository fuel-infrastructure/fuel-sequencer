package types

const (
	// ModuleName defines the module name
	ModuleName = "sequencing"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_sequencing"

	// TopicKey is the prefix to retrieve all Topic
	TopicKey = "Topic/value/"
)

var (
	ParamsKey = []byte("p_sequencing")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// TopicKeyPrefix returns the store key to retrieve a Topic from the topic Id
func TopicKeyPrefix(
	topicId []byte,
) []byte {
	return append([]byte(topicId), topicId...)
}
