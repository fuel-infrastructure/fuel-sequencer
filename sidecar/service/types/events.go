package types

import (
	"errors"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	AuthorizeEventHashFn          = crypto.Keccak256Hash([]byte("Authorize(address,bytes)")).Hex()
)

const (
	// Event names used when parsing log data to events

	DepositEventName            = "Deposit"
	DelegateEventName           = "Delegate"
	RedelegateEventName         = "Redelegate"
	ClaimRewardsEventName       = "ClaimRewards"
	UnbondEventName             = "Unbond"
	WithdrawEventName           = "Withdraw"
	TransferEventName           = "Transfer"
	VoteEventName               = "Vote"
	SetRewardRecipientEventName = "SetRewardRecipient"
	AuthorizeEventName          = "Authorize"
)

// ParsedEvent is a common interface for parsed Ethereum events.
type ParsedEvent interface {
	ValidateBasic() error
	Marshal() (dAtA []byte, err error)
	Unmarshal(dAtA []byte) error
	Messages(cdc codec.BinaryCodec, authority, denom string) ([]*codectypes.Any, error)
	Signer() string
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
func (m *DepositEvent) Messages(_ codec.BinaryCodec, authority, _ string) ([]*codectypes.Any, error) {
	return NewAnysWithValue(m.ToMsgDepositFromEthereum(authority))
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *DepositEvent) Signer() string {
	return m.Depositor
}

// ---------------------------------------------------------------------- DELEGATE

// ValidateBasic performs some sanity checks on the DelegateEvent
func (m *DelegateEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("DelegateEvent is nil")
	}

	return nil
}

// ToMsgDelegate is a convenient function for getting a MsgDelegate from the event.
func (m *DelegateEvent) ToMsgDelegate(denom string) (*stakingtypes.MsgDelegate, error) {

	amount, ok := sdkmath.NewIntFromString(m.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount in DelegateEvent: %s", m.Amount)
	}

	return &stakingtypes.MsgDelegate{
		DelegatorAddress: m.Delegator,
		ValidatorAddress: m.Validator,
		Amount:           sdk.NewCoin(denom, amount),
	}, nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *DelegateEvent) Messages(_ codec.BinaryCodec, _, denom string) ([]*codectypes.Any, error) {

	msg, err := m.ToMsgDelegate(denom)
	if err != nil {
		return nil, err
	}

	return NewAnysWithValue(msg)
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *DelegateEvent) Signer() string {
	return m.Delegator
}

// ---------------------------------------------------------------------- REDELEGATE

// ValidateBasic performs some sanity checks on the RedelegateEvent
func (m *RedelegateEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("RedelegateEvent is nil")
	}

	return nil
}

// ToMsgBeginRedelegate is a convenient function for getting a MsgBeginRedelegate from the event.
func (m *RedelegateEvent) ToMsgBeginRedelegate(denom string) (*stakingtypes.MsgBeginRedelegate, error) {

	amount, ok := sdkmath.NewIntFromString(m.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount in RedelegateEvent: %s", m.Amount)
	}

	return &stakingtypes.MsgBeginRedelegate{
		DelegatorAddress:    m.Delegator,
		ValidatorSrcAddress: m.SrcValidator,
		ValidatorDstAddress: m.DstValidator,
		Amount:              sdk.NewCoin(denom, amount),
	}, nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *RedelegateEvent) Messages(_ codec.BinaryCodec, _, denom string) ([]*codectypes.Any, error) {

	msg, err := m.ToMsgBeginRedelegate(denom)
	if err != nil {
		return nil, err
	}

	return NewAnysWithValue(msg)
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *RedelegateEvent) Signer() string {
	return m.Delegator
}

// ---------------------------------------------------------------------- CLAIM REWARDS

// ValidateBasic performs some sanity checks on the ClaimRewardsEvent
func (m *ClaimRewardsEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("ClaimRewardsEvent is nil")
	}

	return nil
}

// ToMsgWithdrawDelegatorReward is a convenient function for getting a MsgWithdrawDelegatorReward from the event.
func (m *ClaimRewardsEvent) ToMsgWithdrawDelegatorReward() *distributiontypes.MsgWithdrawDelegatorReward {
	return &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: m.Delegator,
		ValidatorAddress: m.Validator,
	}
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *ClaimRewardsEvent) Messages(_ codec.BinaryCodec, _, _ string) ([]*codectypes.Any, error) {
	return NewAnysWithValue(m.ToMsgWithdrawDelegatorReward())
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *ClaimRewardsEvent) Signer() string {
	return m.Delegator
}

// ---------------------------------------------------------------------- UNBOND

// ValidateBasic performs some sanity checks on the DelegateEvent
func (m *UnbondEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("UnbondEvent is nil")
	}

	return nil
}

// ToMsgUndelegate is a convenient function for getting a MsgUndelegate from the event.
func (m *UnbondEvent) ToMsgUndelegate(denom string) (*stakingtypes.MsgUndelegate, error) {

	amount, ok := sdkmath.NewIntFromString(m.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount in UnbondEvent: %s", m.Amount)
	}

	return &stakingtypes.MsgUndelegate{
		DelegatorAddress: m.Delegator,
		ValidatorAddress: m.Validator,
		Amount:           sdk.NewCoin(denom, amount),
	}, nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *UnbondEvent) Messages(_ codec.BinaryCodec, _, denom string) ([]*codectypes.Any, error) {

	msg, err := m.ToMsgUndelegate(denom)
	if err != nil {
		return nil, err
	}

	return NewAnysWithValue(msg)
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *UnbondEvent) Signer() string {
	return m.Delegator
}

// ---------------------------------------------------------------------- WITHDRAW

// ValidateBasic performs some sanity checks on the WithdrawEvent
func (m *WithdrawEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("WithdrawEvent is nil")
	}

	return nil
}

// ToMsgWithdrawToEthereum is a convenient function for getting a MsgWithdrawToEthereum from the event.
func (m *WithdrawEvent) ToMsgWithdrawToEthereum(denom string) (*bridgetypes.MsgWithdrawToEthereum, error) {

	amount, ok := sdkmath.NewIntFromString(m.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount in WithdrawEvent: %s", m.Amount)
	}

	return &bridgetypes.MsgWithdrawToEthereum{
		From:   m.From,
		To:     m.To,
		Amount: sdk.NewCoin(denom, amount),
	}, nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *WithdrawEvent) Messages(_ codec.BinaryCodec, _, denom string) ([]*codectypes.Any, error) {

	msg, err := m.ToMsgWithdrawToEthereum(denom)
	if err != nil {
		return nil, err
	}

	return NewAnysWithValue(msg)
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *WithdrawEvent) Signer() string {
	return m.From
}

// ---------------------------------------------------------------------- TRANSFER

// ValidateBasic performs some sanity checks on the WithdrawEvent
func (m *TransferEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("TransferEvent is nil")
	}

	return nil
}

// ToMsgSend is a convenient function for getting a MsgSend from the event.
func (m *TransferEvent) ToMsgSend(denom string) (*banktypes.MsgSend, error) {

	amount, ok := sdkmath.NewIntFromString(m.Amount)
	if !ok {
		return nil, fmt.Errorf("invalid amount in WithdrawEvent: %s", m.Amount)
	}

	return &banktypes.MsgSend{
		FromAddress: m.Sender,
		ToAddress:   m.Recipient,
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, amount)),
	}, nil
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *TransferEvent) Messages(_ codec.BinaryCodec, _, denom string) ([]*codectypes.Any, error) {

	msg, err := m.ToMsgSend(denom)
	if err != nil {
		return nil, err
	}

	return NewAnysWithValue(msg)
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *TransferEvent) Signer() string {
	return m.Sender
}

// ---------------------------------------------------------------------- VOTE

// ValidateBasic performs some sanity checks on the WithdrawEvent
func (m *VoteEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("VoteEvent is nil")
	}

	return nil
}

// ToMsgVote is a convenient function for getting a MsgVote from the event.
func (m *VoteEvent) ToMsgVote() *govtypesv1.MsgVote {
	return &govtypesv1.MsgVote{
		ProposalId: m.ProposalId,
		Voter:      m.Voter,
		Option:     govtypesv1.VoteOption(m.Option),
		Metadata:   m.Metadata,
	}
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *VoteEvent) Messages(_ codec.BinaryCodec, _, _ string) ([]*codectypes.Any, error) {
	return NewAnysWithValue(m.ToMsgVote())
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *VoteEvent) Signer() string {
	return m.Voter
}

// ---------------------------------------------------------------------- SET REWARD RECIPIENT

// ValidateBasic performs some sanity checks on the SetRewardRecipientEvent
func (m *SetRewardRecipientEvent) ValidateBasic() error {
	// TODO: More checks can be added

	// Error if the receiver is nil
	if m == nil {
		return errors.New("SetRewardRecipientEvent is nil")
	}

	return nil
}

// ToMsgSetWithdrawAddress is a convenient function for getting a MsgSetWithdrawAddress from the event.
func (m *SetRewardRecipientEvent) ToMsgSetWithdrawAddress() *distributiontypes.MsgSetWithdrawAddress {
	return &distributiontypes.MsgSetWithdrawAddress{
		DelegatorAddress: m.Delegator,
		WithdrawAddress:  m.RewardRecipient,
	}
}

// Messages converts the event to a set of messages encoded as Anys, typically to be included in an SDK transaction.
// In this case we only get one message.
func (m *SetRewardRecipientEvent) Messages(_ codec.BinaryCodec, _, _ string) ([]*codectypes.Any, error) {
	return NewAnysWithValue(m.ToMsgSetWithdrawAddress())
}

// Signer returns the address that signed the transaction resulting in this event.
func (m *SetRewardRecipientEvent) Signer() string {
	return m.Delegator
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
func (m *AuthorizeEvent) Messages(cdc codec.BinaryCodec, _, _ string) ([]*codectypes.Any, error) {

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

// Signer returns the address that signed the transaction resulting in this event.
func (m *AuthorizeEvent) Signer() string {
	return m.Sender
}
