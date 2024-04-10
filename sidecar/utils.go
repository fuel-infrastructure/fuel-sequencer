package sidecar

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/icza/dyno"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// processLog decodes an Ethereum log into a specific event struct.
func processLog(vLog types.Log, contractAbi abi.ABI) (*sidecartypes.Event, error) {
	var genericEvent sidecartypes.Event
	var err error

	switch vLog.Topics[0].Hex() {
	case sidecartypes.SendToSequencerEventHashFn:
		var sequencerEvent sidecartypes.SendToSequencerEvent

		// Process the SendToSequencerEvent
		var ethEvent sidecartypes.EthSendToSequencerEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.SendToSequencerEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// From is indexed, so extract it from Topics
		sequencerEvent.From = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.To = ethEvent.To
		sequencerEvent.Duration = ethEvent.Duration.String()
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		genericEvent.EventType = sidecartypes.SendToSequencerEventName
		genericEvent.Data, err = sequencerEvent.Marshal()

	case sidecartypes.AuthorizeEventHashFn:
		var sequencerEvent sidecartypes.AuthorizeEvent

		// Process the Authorize Event
		var ethEvent sidecartypes.EthAuthorizeEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.AuthorizeEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// From is indexed, so extract it from Topics
		sequencerEvent.From = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Message = ethEvent.Message

		// Fillup the generic event with fields
		genericEvent.EventType = sidecartypes.AuthorizeEventName
		genericEvent.Data, err = sequencerEvent.Marshal()
	default:
		return nil, nil
	}

	return &genericEvent, err
}

// MustGetLastEthereumBlockSyncedFromGenesis processes the genesis from an http response and returns
// the last ethereum block synced from the sequencer chain.
func MustGetLastEthereumBlockSyncedFromGenesis(genbz []byte) uint64 {
	g := make(map[string]interface{})
	err := json.Unmarshal(genbz, &g)
	if err != nil {
		log.Fatalf("failed to unmarshal genesis file: %v", err)
	}

	lastEthereumBlockSynced, err := dyno.Get(g, "result", "genesis", "app_state", "bridge", "last_ethereum_block_synced")
	if err != nil {
		log.Fatalf("failed to extract last ethereum block synced from genesis file: %v", err)
	}

	// Convert the last ethereum block synced from interface to string.
	num, ok := lastEthereumBlockSynced.(string)
	if !ok {
		log.Fatalf("failed to convert last ethereum block synced interface to string")
	}

	// Convert the string into uint64 and return.
	lastEthereumBlockSyncedUint, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		log.Fatalf("failed to convert last ethereum block synced from string to uint64: %v", err)
	}

	return lastEthereumBlockSyncedUint
}
