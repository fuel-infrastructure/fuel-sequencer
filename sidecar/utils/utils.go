package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/icza/dyno"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// ExtractLogDataToEvent decodes an Ethereum log into a specific event struct.
func ExtractLogDataToEvent(vLog types.Log, contractAbi abi.ABI) (*sidecartypes.Event, error) {
	var event sidecartypes.Event
	var err error

	switch vLog.Topics[0].Hex() {
	case sidecartypes.MockDepositEventHashFn:
		var sequencerEvent sidecartypes.DepositEvent

		// Process the SendToSequencerEvent
		var ethEvent sidecartypes.MockEthDepositEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.MockDepositEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// From is indexed, so extract it from Topics
		sequencerEvent.Depositor = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Recipient = ethEvent.To
		sequencerEvent.Lockup = ethEvent.Duration.String()
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.MockDepositEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.MockAuthorizeEventHashFn:
		var sequencerEvent sidecartypes.AuthorizeEvent

		// Process the Authorize Event
		var ethEvent sidecartypes.MockEthAuthorizeEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.MockAuthorizeEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// From is indexed, so extract it from Topics
		sequencerEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Data = ethEvent.Message

		// Fillup the generic event with fields
		event.EventType = sidecartypes.MockAuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.DepositEventHashFn:
		var sequencerEvent sidecartypes.DepositEvent

		// Process the DepositEvent
		var ethEvent sidecartypes.EthDepositEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.DepositEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Depositor and Recipient are indexed, so extract them from Topics
		sequencerEvent.Depositor = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Recipient = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Lockup = ethEvent.Lockup.String()
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.DepositEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.AuthorizeEventHashFn:
		var sequencerEvent sidecartypes.AuthorizeEvent

		// Process the Authorize Event
		var ethEvent sidecartypes.EthAuthorizeEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.AuthorizeEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Sender is indexed, so extract it from Topics
		sequencerEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Data = ethEvent.Data

		// Fillup the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()
	default:
		return nil, nil
	}

	return &event, err
}

// MustGetLastEthereumBlockSyncedFromGenesis processes the genesis from an http response and returns
// the last ethereum block synced from the sequencer chain.
func MustGetLastEthereumBlockSyncedFromGenesis(genbz []byte) uint64 {
	g := make(map[string]interface{})
	err := json.Unmarshal(genbz, &g)
	if err != nil {
		log.Fatalf("failed to unmarshal genesis file: %v", err)
	}

	lastEthereumBlockSynced, err := dyno.Get(g, "bridge", "last_ethereum_block_synced")
	if err != nil {
		log.Fatalf("failed to extract LastEthereumBlockSynced from genesis file: %v", err)
	}

	// Convert the last ethereum block synced from interface to string.
	num, ok := lastEthereumBlockSynced.(string)
	if !ok {
		log.Fatalf("failed to convert LastEthereumBlockSynced interface to string")
	}

	// Convert the string into uint64 and return.
	lastEthereumBlockSyncedUint, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		log.Fatalf("failed to convert LastEthereumBlockSynced from string to uint64: %v", err)
	}

	return lastEthereumBlockSyncedUint
}

// ValidateIsLogSequential checks if the log is sequential based on TxIndex and LogIndex.
func ValidateIsLogSequential(vLog types.Log, lastBlockNumber *uint64, lastTxIndex, lastLogIndex *int) error {
	currentBlockNumber := vLog.BlockNumber
	currentTxIndex := int(vLog.TxIndex)
	currentLogIndex := int(vLog.Index)

	// Initial verification to ascertain that the current block's number sequentially follows the last processed block's number.
	if currentBlockNumber != *lastBlockNumber {
		if currentBlockNumber < *lastBlockNumber {
			return fmt.Errorf(
				"non-sequential block detected: current block number %d precedes last processed block number %d",
				currentBlockNumber, *lastBlockNumber,
			)
		}

		// Resetting indices for the new block, acknowledging the transition to a subsequent block in the sequence.
		*lastTxIndex = -1
		*lastLogIndex = -1
	}

	// Ensuring within-block log sequentiality by comparing the current log's indices against the last processed log's indices.
	if currentTxIndex < *lastTxIndex || currentLogIndex <= *lastLogIndex {
		return fmt.Errorf(
			"log sequentiality violation within block %d: currentTxIndex=%d, lastTxIndex=%d, currentLogIndex=%d, lastLogIndex=%d",
			currentBlockNumber, currentTxIndex, *lastTxIndex, currentLogIndex, *lastLogIndex,
		)
	}

	// Upon successful validation, updating tracking variables to reflect the most recent log's indices.
	*lastBlockNumber = currentBlockNumber
	*lastTxIndex = currentTxIndex
	*lastLogIndex = currentLogIndex

	return nil
}
