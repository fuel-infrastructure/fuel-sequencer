package types

import (
	"bytes"
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
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
