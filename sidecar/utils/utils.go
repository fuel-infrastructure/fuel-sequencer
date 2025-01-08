package utils

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
)

// ExtractLogDataToEvent decodes an Ethereum log into a specific event struct.
func ExtractLogDataToEvent(vLog types.Log, contractAbi abi.ABI) (*sidecartypes.Event, error) {
	var event sidecartypes.Event
	var err error

	switch vLog.Topics[0].Hex() {
	case sidecartypes.DepositEventHashFn:
		var sequencerEvent sidecartypes.DepositEvent

		// Process the event
		var ethEvent sidecartypes.EthDepositEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.DepositEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Depositor = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Recipient = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Lockup = ethEvent.Lockup.String()
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.DepositEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.DelegateEventHashFn:
		var sequencerEvent sidecartypes.DelegateEvent

		// Process the event
		var ethEvent sidecartypes.EthDelegateEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.DelegateEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Delegator = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Validator = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.DelegateEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.RedelegateEventHashFn:
		var sequencerEvent sidecartypes.RedelegateEvent

		// Process the event
		var ethEvent sidecartypes.EthRedelegateEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.RedelegateEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Delegator = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.SrcValidator = common.HexToAddress(vLog.Topics[2].Hex()).String()
		sequencerEvent.DstValidator = common.HexToAddress(vLog.Topics[3].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.RedelegateEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.ClaimRewardsEventHashFn:
		var sequencerEvent sidecartypes.ClaimRewardsEvent

		// Process the event
		var ethEvent sidecartypes.EthClaimRewardsEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.ClaimRewardsEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Delegator = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Validator = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.ClaimRewardsEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.UnbondEventHashFn:
		var sequencerEvent sidecartypes.UnbondEvent

		// Process the event
		var ethEvent sidecartypes.EthUnbondEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.UnbondEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Delegator = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Validator = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.UnbondEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.WithdrawEventHashFn:
		var sequencerEvent sidecartypes.WithdrawEvent

		// Process the event
		var ethEvent sidecartypes.EthWithdrawEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.WithdrawEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.From = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.To = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.WithdrawEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.TransferEventHashFn:
		var sequencerEvent sidecartypes.TransferEvent

		// Process the event
		var ethEvent sidecartypes.EthTransferEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.TransferEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.Recipient = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.TransferEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.VoteEventHashFn:
		var sequencerEvent sidecartypes.VoteEvent

		// Process the event
		var ethEvent sidecartypes.EthVoteEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.VoteEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Voter = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		sequencerEvent.ProposalId = ethEvent.ProposalId
		sequencerEvent.Option = ethEvent.Option
		sequencerEvent.Metadata = ethEvent.Metadata

		// Fill up the generic event with fields
		event.EventType = sidecartypes.VoteEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.SetRewardRecipientEventHashFn:
		var sequencerEvent sidecartypes.SetRewardRecipientEvent

		// Process the event
		var ethEvent sidecartypes.EthSetRewardRecipientEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.SetRewardRecipientEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sequencerEvent.Delegator = common.HexToAddress(vLog.Topics[1].Hex()).String()
		sequencerEvent.RewardRecipient = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.SetRewardRecipientEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = sequencerEvent.Marshal()

	case sidecartypes.AuthorizeEventHashFn:
		var sequencerEvent sidecartypes.AuthorizeEvent

		// Process the event
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
