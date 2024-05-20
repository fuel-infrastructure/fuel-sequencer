package types

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

type (
	// EthDepositEvent represents a DepositEvent event raised by the proxy contract. This represents the structure on
	// Ethereum, so it should be used as an intermediary type to convert into the event expected by the Sequencer.
	// Note: Depositor and Recipient are indexed, so they will show up as a vLog topics instead of fields here.
	EthDepositEvent struct {
		Amount *big.Int `json:"amount"`
		Lockup *big.Int `json:"lockup"`
	}

	// EthAuthorizeEvent represents an AuthorizeEvent event raised by the bridge contract. This represents the structure
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

// Validate ensures that the event is valid by checking that:
// - It originated from the expected contract address.
// - It can be parsed, and the parsed version is valid.
func (m *Event) Validate(ethereumProxyContractAddress string) error {
	// Error if the receiver is nil
	if m == nil {
		return errors.New("event is nil")
	}

	// Error if the event does not belong to the Ethereum Proxy Contract.
	// Note: there is no need to check that m.ContractAddress is hex, assuming ethereumProxyContractAddress is correct.
	if m.ContractAddress != ethereumProxyContractAddress {
		return fmt.Errorf(
			"event's contract address does not match expected proxy contract address; got %s, expected %s",
			m.ContractAddress,
			ethereumProxyContractAddress,
		)
	}

	parsedEvent, err := m.UnmarshalParsedEvent()
	if err != nil {
		return err
	}

	err = parsedEvent.ValidateBasic()
	if err != nil {
		return err
	}

	return nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// To do so, we unmarshal the event into a parsed event, and then extract the messages from it.
func (m *Event) Messages(cdc codec.BinaryCodec, authority string) ([]*codectypes.Any, error) {

	parsedEvent, err := m.UnmarshalParsedEvent()
	if err != nil {
		return nil, err
	}

	return parsedEvent.Messages(cdc, authority)
}

// RawTxBytes converts the event to a valid tx that can be injected into a block and produces a tx result.
func (m *Event) RawTxBytes(cdc codec.BinaryCodec, authority string) ([]byte, error) {

	messages, err := m.Messages(cdc, authority)
	if err != nil {
		return nil, err
	}

	return utils.ValidRawTxBytesFromAnyMsgs(messages)
}
