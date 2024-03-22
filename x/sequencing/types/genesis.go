package types

import (
	"encoding/hex"

	errorsmod "cosmossdk.io/errors"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:    DefaultParams(),
		TopicList: []Topic{},
		// this line is used by starport scaffolding # genesis/types/default
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	// Check for duplicated ids in topic
	uniqueTopics := make(map[string]bool)

	for _, t := range gs.TopicList {

		// Convert topic ID (bytes) to a hex string for uniqueness check
		topicIdHex := hex.EncodeToString(t.Id)

		// Verify topic is unique
		if _, ok := uniqueTopics[topicIdHex]; ok {
			return errorsmod.Wrapf(ErrTopicNotUnique, "topic not unique at id %s", topicIdHex)
		}

		uniqueTopics[topicIdHex] = true
	}

	return gs.Params.Validate()
}
