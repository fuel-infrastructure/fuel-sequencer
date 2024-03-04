package abci

/**
TODO: This file should be replaced because it was implemented for demonstration purposes. Please note that here we are
    : using JSON to marshal and unmarshal the transaction for simplicity. In the case of vote extensions the Cosmos SDK
    : docs suggest using a more lightweight encoding that produces a small output, such as compressed bytes or custom
    : encodings. Therefore, consider applying this suggestion to injected transactions as well.
*/

import (
	"encoding/json"
)

type EthEventsTx struct {
	EventsData []string
}

func (e *EthEventsTx) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

func (e *EthEventsTx) Unmarshal(bz []byte) error {
	return json.Unmarshal(bz, e)
}

// isEqualEventsData compares two slices of strings for equality
func isEqualEventsData(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false // Slices of different lengths cannot be equal
	}

	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false // Found unequal elements
		}
	}

	return true // All elements are equal
}

// Equal compares two EthEventsTx structs for equality based on their EventsData
func (e *EthEventsTx) Equal(f *EthEventsTx) bool {
	return isEqualEventsData(e.EventsData, f.EventsData)
}
