package types

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	// EthSendToSequencerEvent represents a SendToSequencerEvent event raised by the bridge contract. This represents
	// the structure on Ethereum, so it should be used as an intermediary type to convert into the event expected by
	// the Sequencer.
	EthSendToSequencerEvent struct {
		From     common.Address
		Amount   *big.Int
		To       string
		Duration *big.Int
	}

	// EthAuthorizeEvent represents a AuthorizeEvent event raised by the bridge contract. This represents the structure
	// on Ethereum, so it should be used as an intermediary type to convert into the event expected by the Sequencer.
	EthAuthorizeEvent struct {
		From    common.Address
		Message []byte
	}

	// EthereumBlock stores events associated with a block.
	EthereumBlock struct {
		BlockNumber *big.Int
		Events      []Event
	}
)

// Equal compares two Event structs for equality
func (m *Event) Equal(e *Event) (bool, error) {
	// If both structs are nil then they are equal
	if m == nil && e == nil {
		return true, nil
	}

	// If one of them only is nil then they are not equal
	if m == nil || e == nil {
		return false, nil
	}

	// If two events are not of the same type then they are not equal
	if m.EventType != e.EventType {
		return false, nil
	}

	// Unmarshal event according to type and check if the events are equal
	switch m.EventType {
	case SendToSequencerEventName:
		var eventData1 SendToSequencerEvent
		err := eventData1.Unmarshal(m.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", SendToSequencerEventName, err)
		}

		var eventData2 SendToSequencerEvent
		err = eventData2.Unmarshal(e.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", SendToSequencerEventName, err)
		}

		// Equality boils down to the specific equality logic of the event type
		return eventData1.Equal(&eventData2), nil
	case AuthorizeEventName:
		var eventData1 AuthorizeEvent
		err := eventData1.Unmarshal(m.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", AuthorizeEventName, err)
		}

		var eventData2 AuthorizeEvent
		err = eventData2.Unmarshal(e.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", AuthorizeEventName, err)
		}

		// Equality boils down to the specific equality logic of the event type
		return eventData1.Equal(&eventData2), nil
	default:
		return false, fmt.Errorf("unknown event type: %s", m.EventType)
	}
}

// ValidateBasic performs some sanity checks on Event
func (m *Event) ValidateBasic() error {
	// Unmarshal event according to type and sanitize the event accordingly
	switch m.EventType {
	case SendToSequencerEventName:
		var eventData SendToSequencerEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return fmt.Errorf("could not unmarshal to %s: %w", SendToSequencerEventName, err)
		}

		// Validation boils down to the specific validation logic of the event type
		return eventData.ValidateBasic()
	case AuthorizeEventName:
		var eventData AuthorizeEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return fmt.Errorf("could not unmarshal to %s: %w", AuthorizeEventName, err)
		}

		// Validation boils down to the specific validation logic of the event type
		return eventData.ValidateBasic()
	default:
		return fmt.Errorf("unknown event type: %s", m.EventType)
	}
}
