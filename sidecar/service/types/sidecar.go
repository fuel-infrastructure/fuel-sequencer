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

// UnmarshalConcreteEvent attempts to unmarshal a specific event from the Event sent by the sidecar
func (m *Event) UnmarshalConcreteEvent() (ConcreteEvent, error) {
	switch m.EventType {
	case SendToSequencerEventName:
		var eventData SendToSequencerEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal to %s: %w", SendToSequencerEventName, err)
		}
		return &eventData, nil
	case AuthorizeEventName:
		var eventData AuthorizeEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal to %s: %w", AuthorizeEventName, err)
		}
		return &eventData, nil
	default:
		return nil, fmt.Errorf("unknown event type: %s", m.EventType)
	}
}

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

	// Get concrete event from the first event
	event1, err := m.UnmarshalConcreteEvent()
	if err != nil {
		return false, err
	}

	// Get concrete event from the second event
	event2, err := e.UnmarshalConcreteEvent()
	if err != nil {
		return false, err
	}

	// Equality boils down to the specific equality logic of the event type
	return event1.Equal(event2), nil
}

// ValidateBasic performs some sanity checks on Event
func (m *Event) ValidateBasic() error {
	// Get concrete event
	event, err := m.UnmarshalConcreteEvent()
	if err != nil {
		return err
	}

	return event.ValidateBasic()
}
