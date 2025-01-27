package types

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

var (
	// Hash function signatures used to identify events

	DepositEventHashFn            = crypto.Keccak256Hash([]byte("Deposit(address,address,uint256,uint256)")).Hex()
	DelegateEventHashFn           = crypto.Keccak256Hash([]byte("Delegate(address,address,uint256)")).Hex()
	RedelegateEventHashFn         = crypto.Keccak256Hash([]byte("Redelegate(address,address,address,uint256)")).Hex()
	ClaimRewardsEventHashFn       = crypto.Keccak256Hash([]byte("ClaimRewards(address,address)")).Hex()
	UnbondEventHashFn             = crypto.Keccak256Hash([]byte("Unbond(address,address,uint256)")).Hex()
	WithdrawEventHashFn           = crypto.Keccak256Hash([]byte("Withdraw(address,address,uint256)")).Hex()
	TransferEventHashFn           = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")).Hex()
	VoteEventHashFn               = crypto.Keccak256Hash([]byte("Vote(address,uint64,uint32,string)")).Hex()
	SetRewardRecipientEventHashFn = crypto.Keccak256Hash([]byte("SetRewardRecipient(address,address)")).Hex()
	GrantEventHashFn              = crypto.Keccak256Hash([]byte("Grant(address,address,string,uint256)")).Hex()
	RevokeEventHashFn             = crypto.Keccak256Hash([]byte("Revoke(address,address,string)")).Hex()
	AuthorizeEventHashFn          = crypto.Keccak256Hash([]byte("Authorize(address,bytes)")).Hex()
)

const (
	// Event names used when parsing log data to events using the SequencerProxy contract ABI

	EthDepositEventName            = "Deposit"
	EthDelegateEventName           = "Delegate"
	EthRedelegateEventName         = "Redelegate"
	EthClaimRewardsEventName       = "ClaimRewards"
	EthUnbondEventName             = "Unbond"
	EthWithdrawEventName           = "Withdraw"
	EthTransferEventName           = "Transfer"
	EthVoteEventName               = "Vote"
	EthSetRewardRecipientEventName = "SetRewardRecipient"
	EthGrantEventName              = "Grant"
	EthRevokeEventName             = "Revoke"
	EthAuthorizeEventName          = "Authorize"

	// Event names recognised by the Sequencer

	DepositEventName   = "Deposit"
	AuthorizeEventName = "Authorize"
)

// ParsedEvent is a common interface for parsed Ethereum events.
type ParsedEvent interface {
	ValidateBasic() error
	Marshal() (dAtA []byte, err error)
	Unmarshal(dAtA []byte) error
	Messages(cdc codec.BinaryCodec, authority string) ([]*codectypes.Any, error)
}

// ---------------------------------------------------------------------- DEPOSIT

// ValidateBasic performs some sanity checks on the DepositEvent
func (m *DepositEvent) ValidateBasic() error {

	// NOTE: we intentionally skip the validation of the Depositor and Recipient since in the message handler we still
	// want to mint the deposited tokens even if we do not know who they are coming from or who they are going to.

	// Error if the receiver is nil
	if m == nil {
		return errors.New("DepositEvent is nil")
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

// ToMsgDepositFromEthereum is a convenient function for getting a MsgDepositFromEthereum from the event.
func (m *DepositEvent) ToMsgDepositFromEthereum(authority string) *bridgetypes.MsgDepositFromEthereum {
	return &bridgetypes.MsgDepositFromEthereum{
		Authority: authority,
		Depositor: m.Depositor,
		Recipient: m.Recipient,
		Amount:    m.Amount,
		Lockup:    m.Lockup,
	}
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *DepositEvent) Messages(_ codec.BinaryCodec, authority string) ([]*codectypes.Any, error) {
	return NewAnysWithValue(m.ToMsgDepositFromEthereum(authority))
}

// ---------------------------------------------------------------------- AUTHORIZE

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

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we get the messages included in the AuthorizeTx encoded in the event data, which are already Anys.
func (m *AuthorizeEvent) Messages(cdc codec.BinaryCodec, _ string) ([]*codectypes.Any, error) {

	// This is a defensive check to ensure only the ProtoCodec is used for message unmarshalling
	if _, ok := cdc.(*codec.ProtoCodec); !ok {
		return nil, bridgetypes.ErrCodecIsNotSupported.Wrap(bridgetypes.ErrStrOnlyProtoCodecAllowed)
	}

	var authorizeTx bridgetypes.AuthorizeTx
	err := cdc.Unmarshal(m.Data, &authorizeTx)
	if err != nil {
		return nil, err
	}

	return authorizeTx.Messages, nil
}
