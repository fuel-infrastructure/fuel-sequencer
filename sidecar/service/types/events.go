package types

import (
	"bytes"
	"encoding/hex"
	"errors"

	sdk "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
)

// Equal compares two SendToSequencerEvent structs for equality
func (m *SendToSequencerEvent) Equal(e *SendToSequencerEvent) bool {
	// If both structs are nil then they are equal
	if m == nil && e == nil {
		return true
	}

	// If one of them only is nil then they are not equal
	if m == nil || e == nil {
		return false
	}

	// Two SendToSequencerEvents are equal if all their elements are equal
	return m.From == e.From &&
		m.To == e.To &&
		m.Duration == e.Duration &&
		m.Amount == e.Amount
}

// ValidateBasic performs some sanity checks on the SendToSequencerEvent
func (m *SendToSequencerEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future

	// Check that From is a valid hex address
	if !common.IsHexAddress(m.From) {
		return errors.New("from is not a valid hex address")
	}

	// Check that To is a valid hex address
	if !common.IsHexAddress(m.To) {
		return errors.New("to is not a valid hex address")
	}

	// Check that the Duration can be converted from a string to sdk.Int
	if _, success := sdk.NewIntFromString(m.Duration); !success {
		return errors.New("could not convert duration to a valid sdk.Int")
	}

	// Check that the Amount can be converted from a string to sdk.Int and is bigger than zero
	amount, success := sdk.NewIntFromString(m.Amount)
	if !success {
		return errors.New("could not convert amount to a valid sdk.Int")
	}
	if amount.IsZero() {
		return errors.New("amount must be bigger than zero")
	}

	return nil
}

// Equal compares two AuthorizeEvent structs for equality
func (m *AuthorizeEvent) Equal(e *AuthorizeEvent) bool {
	// If both structs are nil then they are equal
	if m == nil && e == nil {
		return true
	}

	// If one of them only is nil then they are not equal
	if m == nil || e == nil {
		return false
	}

	// Two AuthorizeEvents are equal if all their elements are equal
	return m.From == e.From && bytes.Equal(m.Message, e.Message)
}

func (m *AuthorizeEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future

	// Check that From is a valid hex address
	if !common.IsHexAddress(m.From) {
		return errors.New("from is not a valid hex address")
	}

	// Confirm that Message is encoded as a proper Hex
	decodedBytes := make([]byte, hex.DecodedLen(len(m.Message)))
	_, err := hex.Decode(decodedBytes, m.Message)
	if err != nil {
		return errors.New("message is not a valid hex")
	}

	return nil
}
