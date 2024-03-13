package types

import (
	"fmt"
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
	// Check for duplicated index in topic
	topicIndexMap := make(map[string]struct{})

	for _, elem := range gs.TopicList {
		index := string(TopicKey(elem.Id.String()))
		if _, ok := topicIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for topic")
		}
		topicIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
