package types

import (
	"fmt"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar"
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
	case sidecar.SendToSequencerEventName:
		var eventData1 SendToSequencerEvent
		err := eventData1.Unmarshal(m.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", sidecar.SendToSequencerEventName, err)
		}

		var eventData2 SendToSequencerEvent
		err = eventData2.Unmarshal(e.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", sidecar.SendToSequencerEventName, err)
		}

		// Equality boils down to the specific equality logic of the event type
		return eventData1.Equal(&eventData2), nil
	case sidecar.AuthorizeEventName:
		var eventData1 AuthorizeEvent
		err := eventData1.Unmarshal(m.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", sidecar.AuthorizeEventName, err)
		}

		var eventData2 AuthorizeEvent
		err = eventData2.Unmarshal(e.Data)
		if err != nil {
			return false, fmt.Errorf("could not unmarshal to %s: %w", sidecar.AuthorizeEventName, err)
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
	case sidecar.SendToSequencerEventName:
		var eventData SendToSequencerEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return fmt.Errorf("could not unmarshal to %s: %w", sidecar.SendToSequencerEventName, err)
		}

		// Validation boils down to the specific validation logic of the event type
		return eventData.ValidateBasic()
	case sidecar.AuthorizeEventName:
		var eventData AuthorizeEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return fmt.Errorf("could not unmarshal to %s: %w", sidecar.AuthorizeEventName, err)
		}

		// Validation boils down to the specific validation logic of the event type
		return eventData.ValidateBasic()
	default:
		return fmt.Errorf("unknown event type: %s", m.EventType)
	}
}
