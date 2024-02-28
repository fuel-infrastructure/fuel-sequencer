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
	// DataCommitmentStoredHashFn Hash function signatures used to identify events
	// crypto.Keccak256Hash([]byte("DataCommitmentStored(uint256,uint64,uint64,bytes32)")).Hex()
	DataCommitmentStoredHashFn = "0x34dd3689f5bd77a60a3ff2e09483dcab032fa2f1fd7227af3e24bed21beab1cb"

	// SendToSequencerEventHashFn Hash function signatures used to identify events
	// crypto.Keccak256Hash([]byte("SendToSequencerEvent(address,uint256,string,uint256)")).Hex()
	SendToSequencerEventHashFn = "0x5dee65305d37f37b03a10fb088c878b533e94440a61a4ad4f99cb82821398f98"

	// AuthorizeEventHashFn Hash function signatures used to identify events
	// crypto.Keccak256Hash([]byte("AuthorizeEvent(address,bytes)")).Hex()
	AuthorizeEventHashFn = "0x0de3682d77bb5d715a5dba2f9da0d61c2afa6d0e32190e6873a3790e03c5965a"

	DataCommitmentStoredName = "DataCommitmentStored"
	SendToSequencerEventName = "SendToSequencerEvent"
	AuthorizeEventName       = "AuthorizeEvent"
)

type (
	// DataCommitmentStored represents a DataCommitmentStored event raised by the bridge contract.
	DataCommitmentStored struct {
		ProofNonce     *big.Int
		StartBlock     uint64
		EndBlock       uint64
		DataCommitment [32]byte
	}

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
	case DataCommitmentStoredHashFn:

		var event DataCommitmentStored
		err = contractAbi.UnpackIntoInterface(&event, DataCommitmentStoredName, vLog.Data)
		if err != nil {
			return genericEvent, err
		}

		// Since StartBlock, EndBlock, and DataCommitment are indexed, extract them from Topics
		event.StartBlock = new(big.Int).SetBytes(vLog.Topics[1].Bytes()).Uint64()
		event.EndBlock = new(big.Int).SetBytes(vLog.Topics[2].Bytes()).Uint64()
		event.DataCommitment = vLog.Topics[3]

		// Fill up the generic event with fields
		genericEvent.EventType = DataCommitmentStoredName
		genericEvent.Data, err = json.Marshal(event)

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

		// Fill up the generic event with fields
		genericEvent.EventType = AuthorizeEventName
		genericEvent.Data, err = json.Marshal(event)
	default:
		return genericEvent, fmt.Errorf("unknown event type")
	}

	return genericEvent, err
}
