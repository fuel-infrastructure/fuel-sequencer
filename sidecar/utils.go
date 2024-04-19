package sidecar

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

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
		genericEvent.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
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
		genericEvent.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		genericEvent.Data, err = sequencerEvent.Marshal()
	default:
		return nil, nil
	}

	return &genericEvent, err
}
