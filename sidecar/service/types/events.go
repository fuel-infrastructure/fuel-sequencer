package types

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
)

const (
	// Hash function signatures used to identify events

	// SendToSequencerEventHashFn Hash function signatures used to identify events
	// crypto.Keccak256Hash([]byte("SendToSequencerEvent(address,uint256,string,uint256)")).Hex()
	SendToSequencerEventHashFn = "0x5dee65305d37f37b03a10fb088c878b533e94440a61a4ad4f99cb82821398f98"

	// AuthorizeEventHashFn Hash function signatures used to identify events
	// crypto.Keccak256Hash([]byte("AuthorizeEvent(address,bytes)")).Hex()
	AuthorizeEventHashFn = "0x0de3682d77bb5d715a5dba2f9da0d61c2afa6d0e32190e6873a3790e03c5965a"

	// Event names
	SendToSequencerEventName = "SendToSequencerEvent"
	AuthorizeEventName       = "AuthorizeEvent"
)

// ParsedEvent is a common interface for parsed Ethereum events.
type ParsedEvent interface {
	Equal(ParsedEvent) bool
	ValidateBasic() error
	Marshal() (dAtA []byte, err error)
	Unmarshal(dAtA []byte) error
}

// Equal attempts to compare two SendToSequencerEvent structs for equality
func (m *SendToSequencerEvent) Equal(e ParsedEvent) bool {
	// Structs are not equal if they are of different type
	other, ok := e.(*SendToSequencerEvent)
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

	// Two SendToSequencerEvents are equal if all their elements are equal
	return m.From == other.From &&
		m.To == other.To &&
		m.Duration == other.Duration &&
		m.Amount == other.Amount
}

// ValidateBasic performs some sanity checks on the SendToSequencerEvent
func (m *SendToSequencerEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future

	// Error if the receiver is nil
	if m == nil {
		return errors.New("SendToSequencerEvent is nil")
	}

	// Check that From is a valid hex address
	if !common.IsHexAddress(m.From) {
		return errors.New("from is not a valid hex address")
	}

	// Check that To is either a valid Sequencer or hex address. Note that To is optional.
	if len(strings.TrimSpace(m.To)) != 0 {

		_, err := sdk.AccAddressFromBech32(m.To)
		if err != nil && !common.IsHexAddress(m.To) {
			return fmt.Errorf("to is not a valid Bech32 or Hex address")
		}
	}

	// Check that the Duration can be converted from a string to sdk.Int
	if _, success := sdkmath.NewIntFromString(m.Duration); !success {
		return errors.New("could not convert duration to a valid sdk.Int")
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
	return m.From == other.From && bytes.Equal(m.Message, other.Message)
}

func (m *AuthorizeEvent) ValidateBasic() error {
	// TODO: More checks can be added in the future

	// Error if the receiver is nil
	if m == nil {
		return errors.New("AuthorizeEvent is nil")
	}

	// Check that From is a valid hex address
	if !common.IsHexAddress(m.From) {
		return errors.New("from is not a valid hex address")
	}

	return nil
}
