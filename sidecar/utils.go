package sidecar

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
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

type (
	// SendToSequencerEvent represents a SendToSequencerEvent event raised by the bridge contract.
	SendToSequencerEvent struct {
		From     common.Address
		Amount   *big.Int
		To       string
		Duration *big.Int
	}

	// AuthorizeEvent represents a AuthorizeEvent event raised by the bridge contract.
	AuthorizeEvent struct {
		From    common.Address
		Message []byte
	}

	// EthereumBlock stores events associated with a block.
	EthereumBlock struct {
		BlockNumber *big.Int
		Events      []sidecartypes.Event
	}
)

// processLog decodes an Ethereum log into a specific event struct.
func processLog(vLog types.Log, contractAbi abi.ABI) (sidecartypes.Event, error) {
	var genericEvent sidecartypes.Event
	var err error

	switch vLog.Topics[0].Hex() {
	case SendToSequencerEventHashFn:

		// Process the SendToSequencerEvent
		var event SendToSequencerEvent
		err = contractAbi.UnpackIntoInterface(&event, SendToSequencerEventName, vLog.Data)
		if err != nil {
			return genericEvent, err
		}

		// From is indexed, so extract it from Topics
		event.From = common.HexToAddress(vLog.Topics[1].Hex())

		// Fill up the generic event with fields
		genericEvent.EventType = SendToSequencerEventName
		genericEvent.Data, err = json.Marshal(event)

	case AuthorizeEventHashFn:

		// Process the Authorize Event
		var event AuthorizeEvent
		err = contractAbi.UnpackIntoInterface(&event, AuthorizeEventName, vLog.Data)
		if err != nil {
			return genericEvent, err
		}

		// From is indexed, so extract it from Topics
		event.From = common.HexToAddress(vLog.Topics[1].Hex())

		// Fillup the generic event with fields
		genericEvent.EventType = AuthorizeEventName
		genericEvent.Data, err = json.Marshal(event)
	default:
		return genericEvent, fmt.Errorf("unknown event type")
	}

	return genericEvent, err
}
