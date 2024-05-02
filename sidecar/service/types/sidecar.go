package types

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	// EthDepositEvent represents a DepositEvent event raised by the proxy contract. This represents the structure on
	// Ethereum, so it should be used as an intermediary type to convert into the event expected by the Sequencer.
	// Note: Depositor and Recipient are indexed, so they will show up as a vLog topics instead of fields here.
	EthDepositEvent struct {
		Amount *big.Int `json:"amount"`
		Lockup *big.Int `json:"lockup"`
	}

	// EthAuthorizeEvent represents a AuthorizeEvent event raised by the bridge contract. This represents the structure
	// on Ethereum, so it should be used as an intermediary type to convert into the event expected by the Sequencer.
	// Note: Sender is indexed, so it will show up as a vLog topic instead of a field here.
	EthAuthorizeEvent struct {
		Data []byte `json:"data"`
	}

	// EthereumBlock stores events associated with a block.
	EthereumBlock struct {
		BlockNumber *big.Int
		Events      []Event
	}
)

// UnmarshalParsedEvent attempts to unmarshal a parsed Ethereum event from the Event sent by the sidecar
func (m *Event) UnmarshalParsedEvent() (ParsedEvent, error) {
	switch m.EventType {
	case DepositEventName:
		var eventData DepositEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal to %s: %w", DepositEventName, err)
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

	// If two events do not originate from the same contract then they are not equal
	if m.ContractAddress != e.ContractAddress {
		return false, nil
	}

	// Get parsed event from the first event
	event1, err := m.UnmarshalParsedEvent()
	if err != nil {
		return false, err
	}

	// Get parsed event from the second event
	event2, err := e.UnmarshalParsedEvent()
	if err != nil {
		return false, err
	}

	// Equality boils down to the specific equality logic of the event type
	return event1.Equal(event2), nil
}

// ValidateBasic performs some sanity checks on Event
func (m *Event) ValidateBasic() error {
	// Error if the receiver is nil
	if m == nil {
		return errors.New("event is nil")
	}

	// Error if the ContractAddress is not a valid Ethereum hex address
	if !common.IsHexAddress(m.ContractAddress) {
		return errors.New("contract_address is not a valid hex address")
	}

	// Get parsed event
	event, err := m.UnmarshalParsedEvent()
	if err != nil {
		return err
	}

	return event.ValidateBasic()
}
