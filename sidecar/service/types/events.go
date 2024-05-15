package types

import (
	"bytes"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	// Hash function signatures used to identify events

	DepositEventHashFn   = crypto.Keccak256Hash([]byte("Deposit(address,address,uint256,uint256)")).Hex()
	AuthorizeEventHashFn = crypto.Keccak256Hash([]byte("Authorize(address,bytes)")).Hex()

	MockDepositEventHashFn   = crypto.Keccak256Hash([]byte("SendToSequencerEvent(address,uint256,string,uint256)")).Hex()
	MockAuthorizeEventHashFn = crypto.Keccak256Hash([]byte("AuthorizeEvent(address,bytes)")).Hex()
)

const (
	// Event names used when parsing log data to events

	DepositEventName   = "Deposit"
	AuthorizeEventName = "Authorize"

	MockDepositEventName   = "SendToSequencerEvent"
	MockAuthorizeEventName = "AuthorizeEvent"
)

// ParsedEvent is a common interface for parsed Ethereum events.
type ParsedEvent interface {
	Equal(ParsedEvent) bool
	ValidateBasic() error
	Marshal() (dAtA []byte, err error)
	Unmarshal(dAtA []byte) error
	Messages(cdc codec.BinaryCodec) ([]*codectypes.Any, error)
}

// Equal attempts to compare two DepositEvent structs for equality
func (m *DepositEvent) Equal(e ParsedEvent) bool {
	// Structs are not equal if they are of different type
	other, ok := e.(*DepositEvent)
	if !ok {
		return false
	}

	// If both structs are nil then they are equal
	if m == nil && other == nil {
		return true
	}

	// If one of them only is nil then they are not equal
	if m == nil || other == nil {
		return false
	}

	// Two DepositEvents are equal if all their elements are equal
	return m.Depositor == other.Depositor &&
		m.Recipient == other.Recipient &&
		m.Lockup == other.Lockup &&
		m.Amount == other.Amount
}

// ValidateBasic performs some sanity checks on the DepositEvent
func (m *DepositEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future
	// TODO: Consider clearing out ValidateBasic since this might cause the deposit event to get skipped!

	// Error if the receiver is nil
	if m == nil {
		return errors.New("DepositEvent is nil")
	}

	// Check that Depositor is a valid hex address
	if !common.IsHexAddress(m.Depositor) {
		return errors.New("depositor is not a valid hex address")
	}

	// Check that Recipient is either a valid Sequencer or hex address.
	_, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil && !common.IsHexAddress(m.Recipient) {
		return fmt.Errorf("recipient is not a valid Bech32 or Hex address")
	}

	// Check that the Lockup can be converted from a string to sdk.Int
	if _, success := sdkmath.NewIntFromString(m.Lockup); !success {
		return errors.New("could not convert lockup to a valid sdk.Int")
	}

	// Check that the Amount can be converted from a string to sdk.Int and is bigger than zero
	amount, success := sdkmath.NewIntFromString(m.Amount)
	if !success {
		return errors.New("could not convert amount to a valid sdk.Int")
	}
	if amount.IsZero() {
		return errors.New("amount must be bigger than zero")
	}

	return nil
}

// ToMsgDepositFromEthereum is a convenient function for getting a MsgDepositFromEthereum from the DepositEvent.
// This is easy because these two have the exact same fields.
func (m *DepositEvent) ToMsgDepositFromEthereum() *bridgetypes.MsgDepositFromEthereum {
	return &bridgetypes.MsgDepositFromEthereum{
		Depositor: m.Depositor,
		Recipient: m.Recipient,
		Amount:    m.Amount,
		Lockup:    m.Lockup,
	}
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message, i.e. a MsgDepositFromEthereum.
func (m *DepositEvent) Messages(codec.BinaryCodec) ([]*codectypes.Any, error) {

	msgDepositFromEthereumAny, err := codectypes.NewAnyWithValue(m.ToMsgDepositFromEthereum())
	if err != nil {
		return nil, err
	}

	return []*codectypes.Any{msgDepositFromEthereumAny}, nil
}

// Equal attempts to compare two AuthorizeEvent structs for equality
func (m *AuthorizeEvent) Equal(e ParsedEvent) bool {
	// Structs are not equal if they are of different type
	other, ok := e.(*AuthorizeEvent)
	if !ok {
		return false
	}

	// If both structs are nil then they are equal
	if m == nil && other == nil {
		return true
	}

	// If one of them only is nil then they are not equal
	if m == nil || other == nil {
		return false
	}

	// Two AuthorizeEvents are equal if all their elements are equal
	return m.Sender == other.Sender && bytes.Equal(m.Data, other.Data)
}

func (m *AuthorizeEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future

	// Error if the receiver is nil
	if m == nil {
		return errors.New("AuthorizeEvent is nil")
	}

	// Check that Sender is a valid hex address
	if !common.IsHexAddress(m.Sender) {
		return errors.New("sender is not a valid hex address")
	}

	return nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we get the messages included in the AuthorizeTx encoded in the event data, which are already Anys.
func (m *AuthorizeEvent) Messages(cdc codec.BinaryCodec) ([]*codectypes.Any, error) {

	// This is a defensive check to ensure only the ProtoCodec is used for message unmarshalling
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, bridgetypes.ErrCodecIsNotSupported.Wrap(bridgetypes.ErrStrOnlyProtoCodecAllowed)
	}

	var authorizeTx bridgetypes.AuthorizeTx
	err := cdc.Unmarshal(m.Data, &authorizeTx)
	if err != nil {
		return nil, err
	}

	return authorizeTx.Messages, nil
}
