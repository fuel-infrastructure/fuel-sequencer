package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestParsedEvent_ValidateBasic(t *testing.T) {
	var nilDepositEvent *sidecartypes.DepositEvent = nil
	var nilDelegateEvent *sidecartypes.DelegateEvent = nil
	var nilRedelegateEvent *sidecartypes.RedelegateEvent = nil
	var nilClaimRewardsEvent *sidecartypes.ClaimRewardsEvent = nil
	var nilUnbondEvent *sidecartypes.UnbondEvent = nil
	var nilWithdrawEvent *sidecartypes.WithdrawEvent = nil
	var nilTransferEvent *sidecartypes.TransferEvent = nil
	var nilVoteEvent *sidecartypes.VoteEvent = nil
	var nilSetRewardRecipientEvent *sidecartypes.SetRewardRecipientEvent = nil
	var nilAuthorizeEvent *sidecartypes.AuthorizeEvent = nil

	testCases := []struct {
		name      string
		event     sidecartypes.ParsedEvent
		expErrMsg string
	}{
		{
			name:  "DepositEvent - valid - recipient is a Hex address",
			event: testtypes.TestDepositEvent1,
		},
		{
			name:  "DepositEvent - valid - recipient is a Sequencer address",
			event: testtypes.TestDepositEvent9,
		},
		{
			name:  "DepositEvent - valid - recipient is the null address",
			event: testtypes.TestDepositEvent2,
		},
		{
			name:      "DepositEvent - nil receiver - error",
			event:     nilDepositEvent,
			expErrMsg: "DepositEvent is nil",
		},
		{
			name: "DepositEvent - invalid depositor - NO ERROR",
			event: &sidecartypes.DepositEvent{
				Depositor: "invalid-depositor",
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "DepositEvent - invalid recipient - NO ERROR",
			event: &sidecartypes.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: "invalid-recipient",
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "DepositEvent - invalid lockup - error",
			event: &sidecartypes.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    "0.23523",
			},
			expErrMsg: "could not convert lockup to a valid sdk.Int",
		},
		{
			name: "DepositEvent - amount is zero - error",
			event: &sidecartypes.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0",
				Lockup:    testtypes.TestLockup1,
			},
			expErrMsg: "amount must be bigger than zero",
		},
		{
			name: "DepositEvent - amount is float - error",
			event: &sidecartypes.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0.4356346",
				Lockup:    testtypes.TestLockup1,
			},
			expErrMsg: "could not convert amount to a valid sdk.Int",
		},
		{
			name:  "AuthorizeEvent - valid",
			event: testtypes.TestAuthorizeEvent1,
		},
		{
			name:      "AuthorizeEvent - nil receiver - error",
			event:     nilAuthorizeEvent,
			expErrMsg: "AuthorizeEvent is nil",
		},
		{
			name: "AuthorizeEvent - invalid sender - error",
			event: &sidecartypes.AuthorizeEvent{
				Sender: "invalid-sender",
				Data:   testutils.MustHexDecodeString(testtypes.TestData1),
			},
			expErrMsg: "sender is not a valid hex address",
		},
		{
			name:      "DelegateEvent - nil receiver - error",
			event:     nilDelegateEvent,
			expErrMsg: "DelegateEvent is nil",
		},
		{
			name:      "RedelegateEvent - nil receiver - error",
			event:     nilRedelegateEvent,
			expErrMsg: "RedelegateEvent is nil",
		},
		{
			name:      "ClaimRewardsEvent - nil receiver - error",
			event:     nilClaimRewardsEvent,
			expErrMsg: "ClaimRewardsEvent is nil",
		},
		{
			name:      "UnbondEvent - nil receiver - error",
			event:     nilUnbondEvent,
			expErrMsg: "UnbondEvent is nil",
		},
		{
			name:      "WithdrawEvent - nil receiver - error",
			event:     nilWithdrawEvent,
			expErrMsg: "WithdrawEvent is nil",
		},
		{
			name:      "TransferEvent - nil receiver - error",
			event:     nilTransferEvent,
			expErrMsg: "TransferEvent is nil",
		},
		{
			name:      "VoteEvent - nil receiver - error",
			event:     nilVoteEvent,
			expErrMsg: "VoteEvent is nil",
		},
		{
			name:      "SetRewardRecipientEvent - nil receiver - error",
			event:     nilSetRewardRecipientEvent,
			expErrMsg: "SetRewardRecipientEvent is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.ValidateBasic()
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestParsedEvent_Messages(t *testing.T) {

	testCases := []struct {
		name       string
		event      sidecartypes.ParsedEvent
		getExpMsgs func() []*codectypes.Any
		expErrMsg  string
	}{
		{
			name:  "DepositEvent",
			event: testtypes.TestDepositEvent1,
			getExpMsgs: func() []*codectypes.Any {
				anys, err := sidecartypes.NewAnysWithValue(
					&bridgetypes.MsgDepositFromEthereum{
						Authority: testtypes.TestGovernanceAddress,
						Depositor: testtypes.TestDepositEvent1.Depositor,
						Recipient: testtypes.TestDepositEvent1.Recipient,
						Amount:    testtypes.TestDepositEvent1.Amount,
						Lockup:    testtypes.TestDepositEvent1.Lockup,
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "DelegateEvent",
			event: testtypes.TestDelegateEvent,
			getExpMsgs: func() []*codectypes.Any {
				amount, ok := sdkmath.NewIntFromString(testtypes.TestDelegateEvent.Amount)
				require.True(t, ok)

				anys, err := sidecartypes.NewAnysWithValue(
					&stakingtypes.MsgDelegate{
						DelegatorAddress: testtypes.TestDelegateEvent.Delegator,
						ValidatorAddress: testtypes.TestDelegateEvent.Validator,
						Amount:           sdk.NewCoin(testtypes.TestToken, amount),
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "RedelegateEvent",
			event: testtypes.TestRedelegateEvent,
			getExpMsgs: func() []*codectypes.Any {
				amount, ok := sdkmath.NewIntFromString(testtypes.TestRedelegateEvent.Amount)
				require.True(t, ok)

				anys, err := sidecartypes.NewAnysWithValue(
					&stakingtypes.MsgBeginRedelegate{
						DelegatorAddress:    testtypes.TestRedelegateEvent.Delegator,
						ValidatorSrcAddress: testtypes.TestRedelegateEvent.SrcValidator,
						ValidatorDstAddress: testtypes.TestRedelegateEvent.DstValidator,
						Amount:              sdk.NewCoin(testtypes.TestToken, amount),
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "ClaimRewardsEvent",
			event: testtypes.TestClaimRewardsEvent,
			getExpMsgs: func() []*codectypes.Any {
				anys, err := sidecartypes.NewAnysWithValue(
					&distributiontypes.MsgWithdrawDelegatorReward{
						DelegatorAddress: testtypes.TestClaimRewardsEvent.Delegator,
						ValidatorAddress: testtypes.TestClaimRewardsEvent.Validator,
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "UnbondEvent",
			event: testtypes.TestUnbondEvent,
			getExpMsgs: func() []*codectypes.Any {
				amount, ok := sdkmath.NewIntFromString(testtypes.TestUnbondEvent.Amount)
				require.True(t, ok)

				anys, err := sidecartypes.NewAnysWithValue(
					&stakingtypes.MsgUndelegate{
						DelegatorAddress: testtypes.TestUnbondEvent.Delegator,
						ValidatorAddress: testtypes.TestUnbondEvent.Validator,
						Amount:           sdk.NewCoin(testtypes.TestToken, amount),
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "WithdrawEvent",
			event: testtypes.TestWithdrawEvent,
			getExpMsgs: func() []*codectypes.Any {
				amount, ok := sdkmath.NewIntFromString(testtypes.TestWithdrawEvent.Amount)
				require.True(t, ok)

				anys, err := sidecartypes.NewAnysWithValue(
					&bridgetypes.MsgWithdrawToEthereum{
						From:   testtypes.TestWithdrawEvent.From,
						To:     testtypes.TestWithdrawEvent.To,
						Amount: sdk.NewCoin(testtypes.TestToken, amount),
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "TransferEvent",
			event: testtypes.TestTransferEvent,
			getExpMsgs: func() []*codectypes.Any {
				amount, ok := sdkmath.NewIntFromString(testtypes.TestTransferEvent.Amount)
				require.True(t, ok)

				anys, err := sidecartypes.NewAnysWithValue(
					&banktypes.MsgSend{
						FromAddress: testtypes.TestTransferEvent.Sender,
						ToAddress:   testtypes.TestTransferEvent.Recipient,
						Amount:      sdk.NewCoins(sdk.NewCoin(testtypes.TestToken, amount)),
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "VoteEvent",
			event: testtypes.TestVoteEvent,
			getExpMsgs: func() []*codectypes.Any {
				anys, err := sidecartypes.NewAnysWithValue(
					&govtypesv1.MsgVote{
						ProposalId: testtypes.TestVoteEvent.ProposalId,
						Voter:      testtypes.TestVoteEvent.Voter,
						Option:     govtypesv1.VoteOption(testtypes.TestVoteEvent.Option),
						Metadata:   testtypes.TestVoteEvent.Metadata,
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "SetRewardRecipientEvent",
			event: testtypes.TestSetRewardRecipientEvent,
			getExpMsgs: func() []*codectypes.Any {
				anys, err := sidecartypes.NewAnysWithValue(
					&distributiontypes.MsgSetWithdrawAddress{
						DelegatorAddress: testtypes.TestSetRewardRecipientEvent.Delegator,
						WithdrawAddress:  testtypes.TestSetRewardRecipientEvent.RewardRecipient,
					})
				require.NoError(t, err)
				return anys
			},
		},
		{
			name:  "AuthorizeEvent",
			event: testtypes.TestAuthorizeEvent1,
			getExpMsgs: func() []*codectypes.Any {

				var authorizeTx bridgetypes.AuthorizeTx
				err := testtypes.TestCdc.Unmarshal(testtypes.TestAuthorizeEvent1.Data, &authorizeTx)
				require.NoError(t, err)

				return authorizeTx.Messages
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msgs, err := tc.event.Messages(testtypes.TestCdc, testtypes.TestGovernanceAddress, testtypes.TestToken)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			expMsgs := tc.getExpMsgs()
			require.Equal(t, expMsgs, msgs)
		})
	}
}

func TestParsedEvent_Signer(t *testing.T) {

	testCases := []struct {
		name      string
		event     sidecartypes.ParsedEvent
		expSigner string
	}{
		{
			name:      "DepositEvent",
			event:     testtypes.TestDepositEvent1,
			expSigner: testtypes.TestDepositEvent1.Depositor,
		},
		{
			name:      "DelegateEvent",
			event:     testtypes.TestDelegateEvent,
			expSigner: testtypes.TestDelegateEvent.Delegator,
		},
		{
			name:      "RedelegateEvent",
			event:     testtypes.TestRedelegateEvent,
			expSigner: testtypes.TestRedelegateEvent.Delegator,
		},
		{
			name:      "ClaimRewardsEvent",
			event:     testtypes.TestClaimRewardsEvent,
			expSigner: testtypes.TestClaimRewardsEvent.Delegator,
		},
		{
			name:      "UnbondEvent",
			event:     testtypes.TestUnbondEvent,
			expSigner: testtypes.TestUnbondEvent.Delegator,
		},
		{
			name:      "WithdrawEvent",
			event:     testtypes.TestWithdrawEvent,
			expSigner: testtypes.TestWithdrawEvent.From,
		},
		{
			name:      "TransferEvent",
			event:     testtypes.TestTransferEvent,
			expSigner: testtypes.TestTransferEvent.Sender,
		},
		{
			name:      "VoteEvent",
			event:     testtypes.TestVoteEvent,
			expSigner: testtypes.TestVoteEvent.Voter,
		},
		{
			name:      "SetRewardRecipientEvent",
			event:     testtypes.TestSetRewardRecipientEvent,
			expSigner: testtypes.TestSetRewardRecipientEvent.Delegator,
		},
		{
			name:      "AuthorizeEvent",
			event:     testtypes.TestAuthorizeEvent1,
			expSigner: testtypes.TestDelegateEvent.Delegator,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expSigner, tc.event.Signer())
		})
	}
}
