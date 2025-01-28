package utils

import (
	"fmt"
	"math/big"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// AuthorizeTxFromMsg packs a message into an AuthorizeTx which the Sequencer can then unpack.
func AuthorizeTxFromMsg(msg sdk.Msg) ([]byte, error) {
	anyMsgs, err := sidecartypes.NewAnysWithValue(msg)
	if err != nil {
		return nil, err
	}

	// Serialize the message to bytes
	bz, err := proto.Marshal(&bridgetypes.AuthorizeTx{Messages: anyMsgs})
	if err != nil {
		return nil, err
	}

	return bz, nil
}

// ExtractLogDataToEvent decodes an Ethereum log into a specific event struct.
func ExtractLogDataToEvent(
	vLog ethereumtypes.Log,
	contractAbi abi.ABI,
	bridgeDenom string,
) (*sidecartypes.Event, error) {
	var event sidecartypes.Event
	var err error

	switch vLog.Topics[0].Hex() {
	case sidecartypes.DepositEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthDepositEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthDepositEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		var depositEvent sidecartypes.DepositEvent

		// Some values are indexed, so extract them from Topics
		depositEvent.Depositor = common.HexToAddress(vLog.Topics[1].Hex()).String()
		depositEvent.Recipient = common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		depositEvent.Lockup = ethEvent.Lockup.String()
		depositEvent.Amount = ethEvent.Amount.String()

		// Fill up the generic event with fields
		event.EventType = sidecartypes.DepositEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = depositEvent.Marshal()

	case sidecartypes.DelegateEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthDelegateEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthDelegateEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		delegator := common.HexToAddress(vLog.Topics[1].Hex()).String()
		validator := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		amount := sdkmath.NewIntFromBigInt(ethEvent.Amount)

		// Build an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = delegator
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&stakingtypes.MsgDelegate{
			DelegatorAddress: delegator,
			ValidatorAddress: validator,
			Amount:           sdk.NewCoin(bridgeDenom, amount),
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.RedelegateEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthRedelegateEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthRedelegateEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		delegator := common.HexToAddress(vLog.Topics[1].Hex()).String()
		srcValidator := common.HexToAddress(vLog.Topics[2].Hex()).String()
		dstValidator := common.HexToAddress(vLog.Topics[3].Hex()).String()

		// Convert the rest of the fields as required
		amount := sdkmath.NewIntFromBigInt(ethEvent.Amount)

		// Build an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = delegator
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    delegator,
			ValidatorSrcAddress: srcValidator,
			ValidatorDstAddress: dstValidator,
			Amount:              sdk.NewCoin(bridgeDenom, amount),
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.ClaimRewardsEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthClaimRewardsEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthClaimRewardsEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		delegator := common.HexToAddress(vLog.Topics[1].Hex()).String()
		validator := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = delegator
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&distributiontypes.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegator,
			ValidatorAddress: validator,
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.UnbondEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthUnbondEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthUnbondEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		delegator := common.HexToAddress(vLog.Topics[1].Hex()).String()
		validator := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		amount := sdkmath.NewIntFromBigInt(ethEvent.Amount)

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = delegator
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&stakingtypes.MsgUndelegate{
			DelegatorAddress: delegator,
			ValidatorAddress: validator,
			Amount:           sdk.NewCoin(bridgeDenom, amount),
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.WithdrawEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthWithdrawEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthWithdrawEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		from := common.HexToAddress(vLog.Topics[1].Hex()).String()
		to := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		amount := sdkmath.NewIntFromBigInt(ethEvent.Amount)

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = from
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&bridgetypes.MsgWithdrawToEthereum{
			From:   from,
			To:     to,
			Amount: sdk.NewCoin(bridgeDenom, amount),
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.TransferEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthTransferEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthTransferEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		sender := common.HexToAddress(vLog.Topics[1].Hex()).String()
		recipient := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Convert the rest of the fields as required
		amount := sdkmath.NewIntFromBigInt(ethEvent.Amount)

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = sender
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&banktypes.MsgSend{
			FromAddress: sender,
			ToAddress:   recipient,
			Amount:      sdk.NewCoins(sdk.NewCoin(bridgeDenom, amount)),
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.VoteEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthVoteEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthVoteEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		voter := common.HexToAddress(vLog.Topics[1].Hex()).String()
		proposalIdBig := new(big.Int)
		proposalIdBig.SetBytes(vLog.Topics[2].Bytes())
		proposalId := proposalIdBig.Uint64()

		// Convert the rest of the fields as required
		option := ethEvent.Option
		metadata := ethEvent.Metadata

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = voter
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&govtypesv1.MsgVote{
			ProposalId: proposalId,
			Voter:      voter,
			Option:     govtypesv1.VoteOption(option),
			Metadata:   metadata,
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.SetRewardRecipientEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthSetRewardRecipientEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthSetRewardRecipientEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		delegator := common.HexToAddress(vLog.Topics[1].Hex()).String()
		rewardRecipient := common.HexToAddress(vLog.Topics[2].Hex()).String()

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = delegator
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&distributiontypes.MsgSetWithdrawAddress{
			DelegatorAddress: delegator,
			WithdrawAddress:  rewardRecipient,
		})
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.GrantEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthGrantEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthGrantEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		granter := common.HexToAddress(vLog.Topics[1].Hex()).String()
		grantee := common.HexToAddress(vLog.Topics[2].Hex()).String()

		msgAuth := authz.GenericAuthorization{Msg: ethEvent.MsgTypeUrl}

		var authExpiration *time.Time = nil
		if ethEvent.Expiration > 0 {
			a := time.Unix(int64(ethEvent.Expiration), 0)
			authExpiration = &a
		}
		// Generate grant msg
		grantMsg := authz.MsgGrant{
			Granter: granter,
			Grantee: grantee,
			Grant: authz.Grant{
				Expiration: authExpiration,
			},
		}
		grantMsg.SetAuthorization(&msgAuth)

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = granter
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&grantMsg)
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()

	case sidecartypes.RevokeEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthRevokeEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthRevokeEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		// Some values are indexed, so extract them from Topics
		granter := common.HexToAddress(vLog.Topics[1].Hex()).String()
		grantee := common.HexToAddress(vLog.Topics[2].Hex()).String()
		msgTypeUrl := ethEvent.MsgTypeUrl

		// Generate revoke msg
		revokeMsg := authz.MsgRevoke{
			Granter:    granter,
			Grantee:    grantee,
			MsgTypeUrl: msgTypeUrl,
		}

		// Generate an authorize event
		var authorizeEvent sidecartypes.AuthorizeEvent
		authorizeEvent.Sender = granter
		authorizeEvent.Data, err = AuthorizeTxFromMsg(&revokeMsg)
		if err != nil {
			return nil, err
		}

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()
	case sidecartypes.AuthorizeEventHashFn:
		// Process the event
		var ethEvent sidecartypes.EthAuthorizeEvent
		err = contractAbi.UnpackIntoInterface(&ethEvent, sidecartypes.EthAuthorizeEventName, vLog.Data)
		if err != nil {
			return nil, err
		}

		var authorizeEvent sidecartypes.AuthorizeEvent

		// Some values are indexed, so extract them from Topics
		authorizeEvent.Sender = common.HexToAddress(vLog.Topics[1].Hex()).String()

		// Convert the rest of the fields as required
		authorizeEvent.Data = ethEvent.Data

		// Fill up the generic event with fields
		event.EventType = sidecartypes.AuthorizeEventName
		event.ContractAddress = common.HexToAddress(vLog.Address.Hex()).String()
		event.Data, err = authorizeEvent.Marshal()
	default:
		return nil, nil
	}

	return &event, err
}

// ValidateIsLogSequential checks if the log is sequential based on TxIndex and LogIndex.
func ValidateIsLogSequential(vLog ethereumtypes.Log, lastBlockNumber *uint64, lastTxIndex, lastLogIndex *int) error {
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
