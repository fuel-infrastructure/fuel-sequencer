package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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

	// Check for duplicated index in topic and that they are sequential
	uniqueTopics := make(map[math.Int]bool)

	for i, t := range gs.TopicList {

		// Verify topic id is sequential
		expId := math.NewInt(int64(i))
		if !t.Id.Equal(math.NewInt(int64(i))) {
			return errorsmod.Wrapf(
				ErrInvalidGenesis,
				"topic id's are not sequential expected %s got %s",
				expId.String(), t.Id.String())
		}

		// Verify topic is unique
		if _, ok := uniqueTopics[t.Id]; ok {
			return errorsmod.Wrapf(ErrTopicNotUnique, "topic not unique at id %s", t.Id.String())
		}

		uniqueTopics[t.Id] = true
	}

	return gs.Params.Validate()
}
