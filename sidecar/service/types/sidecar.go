package types

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
)

// These EthEvent types represent events raised by the SequencerProxy contract. They represent the structure on Ethereum
// and are used as an intermediary type to convert into the event expected by the Sequencer. Some fields are indexed,
// so they will show up as vLog topics instead of fields within these structs, when we are parsing the Ethereum events.
type (
	EthDepositEvent struct {
		Amount *big.Int `json:"amount"`
		Lockup *big.Int `json:"lockup"`
	}

	EthDelegateEvent struct {
		Amount *big.Int `json:"amount"`
	}

	EthRedelegateEvent struct {
		Amount *big.Int `json:"amount"`
	}

	EthClaimRewardsEvent struct{}

	EthUnbondEvent struct {
		Amount *big.Int `json:"amount"`
	}

	EthWithdrawEvent struct {
		Amount *big.Int `json:"amount"`
	}

	EthTransferEvent struct {
		Amount *big.Int `json:"amount"`
	}

	EthVoteEvent struct {
		Option   uint32 `json:"option"`
		Metadata string `json:"metadata"`
	}

	EthSetRewardRecipientEvent struct{}

	EthGrantEvent struct {
		Grantee    string   `json:"grantee"`
		Grant      string   `json:"grant"`
		MsgTypeUrl string   `json:"msgTypeUrl"`
		Expiration *big.Int `json:"expiration"`
	}

	EthRevokeEvent struct {
		Grantee    string `json:"grantee"`
		Grant      string `json:"grant"`
		MsgTypeUrl string `json:"msgTypeUrl"`
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
			return nil, fmt.Errorf("could not unmarshal to %s: %w", m.EventType, err)
		}
		return &eventData, nil
	case AuthorizeEventName:
		var eventData AuthorizeEvent
		err := eventData.Unmarshal(m.Data)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal to %s: %w", m.EventType, err)
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
// The sequence, presumed to be unique, ensures that the generated tx is unique and thus has a unique tx hash.
func (m *Event) RawTxBytes(cdc codec.BinaryCodec, authority string, sequence uint64) ([]byte, error) {

	messages, err := m.Messages(cdc, authority)
	if err != nil {
		return nil, err
	}

	return utils.ValidRawTxBytesFromAnyMsgs(messages, sequence)
}

// RawTxBytesWithMaxBytes makes use of RawTxBytes with an additional size verification. This function will error if the
// bytes returned from RawTxBytes exceed the specified max bytes.
func (m *Event) RawTxBytesWithMaxBytes(
	cdc codec.BinaryCodec, authority string, maxBytes uint64, sequence uint64,
) ([]byte, error) {

	bz, err := m.RawTxBytes(cdc, authority, sequence)
	if err != nil {
		return nil, err
	}

	txSize := utils.TxSize(bz)
	if txSize > maxBytes {
		return nil, fmt.Errorf("generated raw tx bytes exceeded max bytes; %d > %d", txSize, maxBytes)
	}

	return bz, nil
}

// RawTxBytesWithLimitChecks computes the bytes using RawTxBytesWithMaxBytes. This function will error if the computed
// bytes exceed the specified maxBytes or if an Authorize event has more messages than the specified
// maxAuthorizeMessages
func (m *Event) RawTxBytesWithLimitChecks(
	cdc codec.BinaryCodec, authority string, maxBytes, maxAuthorizeMessages, sequence uint64,
) ([]byte, error) {

	if m.EventType == AuthorizeEventName {
		messages, err := m.Messages(cdc, authority)
		if err != nil {
			return nil, err
		}

		if uint64(len(messages)) > maxAuthorizeMessages {
			return nil, fmt.Errorf(
				"authorize event has too many messages; %d > %d", uint64(len(messages)), maxAuthorizeMessages,
			)
		}
	}

	return m.RawTxBytesWithMaxBytes(cdc, authority, maxBytes, sequence)
}
