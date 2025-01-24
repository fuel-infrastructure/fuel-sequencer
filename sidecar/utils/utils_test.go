package utils

import (
	"encoding/json"
	"strconv"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereumtypes "github.com/ethereum/go-ethereum/core/types"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil/fixtures"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sequencerProxyContractABI(t *testing.T) abi.ABI {
	var contractAbi abi.ABI
	err := contractAbi.UnmarshalJSON([]byte(sidecartypes.SequencerProxyContractABI))
	require.NoError(t, err)
	return contractAbi
}

func eventFromMsg(t *testing.T, msg sdk.Msg) *sidecartypes.Event {
	data, err := AuthorizeTxFromMsg(msg)
	require.NoError(t, err)

	authorizeEvent := sidecartypes.AuthorizeEvent{
		Sender: fixtures.SenderAddress,
		Data:   data,
	}
	authorizeEventBz, err := authorizeEvent.Marshal()
	require.NoError(t, err)

	event := &sidecartypes.Event{
		EventType:       sidecartypes.AuthorizeEventName,
		Data:            authorizeEventBz,
		ContractAddress: fixtures.SequencerProxyContractAddress,
	}
	return event
}

func eventFromDepositEvent(t *testing.T, depositEvent sidecartypes.DepositEvent) *sidecartypes.Event {
	depositEventBz, err := depositEvent.Marshal()
	require.NoError(t, err)

	return &sidecartypes.Event{
		EventType:       sidecartypes.DepositEventName,
		Data:            depositEventBz,
		ContractAddress: fixtures.SequencerProxyContractAddress,
	}
}

func TestExtractLogDataToEvent_AuthorizeTxFromEvent(t *testing.T) {
	sequencerProxyABI := sequencerProxyContractABI(t)
	amountParsed := sdkmath.NewInt(fixtures.Amount)

	testCases := []struct {
		name        string
		logs        string
		getExpEvent func() *sidecartypes.Event
		expErrMsg   string
	}{
		{
			name: "Deposit event via Deposit",
			logs: fixtures.DepositLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.SenderAddress, // sender deposits to themselves
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    "0",
				})
			},
		},
		{
			name: "Deposit event via DepositFor",
			logs: fixtures.DepositForLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.ReceiverAddress,
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    "0",
				})
			},
		},
		{
			name: "Deposit event with lockup",
			logs: fixtures.DepositWithLockupLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromDepositEvent(t, sidecartypes.DepositEvent{
					Depositor: fixtures.SenderAddress,
					Recipient: fixtures.SenderAddress, // sender deposits to themselves
					Amount:    strconv.FormatInt(fixtures.Amount, 10),
					Lockup:    fixtures.VestingDurationSeconds,
				})
			},
		},
		{
			name: "Delegate event",
			logs: fixtures.DelegateLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &stakingtypes.MsgDelegate{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
					Amount:           sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Redelegate event",
			logs: fixtures.RedelegateLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &stakingtypes.MsgBeginRedelegate{
					DelegatorAddress:    fixtures.SenderAddress,
					ValidatorSrcAddress: fixtures.Validator1Address,
					ValidatorDstAddress: fixtures.Validator2Address,
					Amount:              sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "ClaimRewards event",
			logs: fixtures.ClaimRewardsLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &distributiontypes.MsgWithdrawDelegatorReward{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
				})
			},
		},
		{
			name: "Unbond event",
			logs: fixtures.UnbondLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &stakingtypes.MsgUndelegate{
					DelegatorAddress: fixtures.SenderAddress,
					ValidatorAddress: fixtures.Validator1Address,
					Amount:           sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Withdraw event via Withdraw",
			logs: fixtures.WithdrawLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &bridgetypes.MsgWithdrawToEthereum{
					From:   fixtures.SenderAddress,
					To:     fixtures.SenderAddress, // sender withdraws to themselves
					Amount: sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Withdraw event via WithdrawTo",
			logs: fixtures.WithdrawToLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &bridgetypes.MsgWithdrawToEthereum{
					From:   fixtures.SenderAddress,
					To:     fixtures.ReceiverAddress,
					Amount: sdk.NewCoin(fixtures.BridgeDenom, amountParsed),
				})
			},
		},
		{
			name: "Transfer event",
			logs: fixtures.TransferLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &banktypes.MsgSend{
					FromAddress: fixtures.SenderAddress,
					ToAddress:   fixtures.ReceiverAddress,
					Amount:      sdk.NewCoins(sdk.NewCoin(fixtures.BridgeDenom, amountParsed)),
				})
			},
		},
		{
			name: "Vote event",
			logs: fixtures.VoteLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &govtypesv1.MsgVote{
					ProposalId: fixtures.VoteProposalId,
					Voter:      fixtures.SenderAddress,
					Option:     govtypesv1.VoteOption(fixtures.VoteOption),
					Metadata:   fixtures.VoteMetadata,
				})
			},
		},
		{
			name: "SetRewardRecipient event",
			logs: fixtures.SetRewardRecipientLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &distributiontypes.MsgSetWithdrawAddress{
					DelegatorAddress: fixtures.SenderAddress,
					WithdrawAddress:  fixtures.ReceiverAddress,
				})
			},
		},
		{
			name: "Authorize event",
			logs: fixtures.AuthorizeLogs,
			getExpEvent: func() *sidecartypes.Event {
				return eventFromMsg(t, &banktypes.MsgSend{
					FromAddress: fixtures.SenderAddress,
					ToAddress:   fixtures.ReceiverAddress,
					Amount:      sdk.NewCoins(sdk.NewCoin(fixtures.BridgeDenom, amountParsed)),
				})
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedEvent := tc.getExpEvent()

			// Parse log into Log struct
			var parsedLog []ethereumtypes.Log
			err := json.Unmarshal([]byte(tc.logs), &parsedLog)
			require.NoError(t, err)

			// There might be multiple logs. Check that there is exactly one log that matches the expected log.
			matches := 0
			for _, log := range parsedLog {
				event, err := ExtractLogDataToEvent(log, sequencerProxyABI, fixtures.BridgeDenom)
				if tc.expErrMsg != "" {
					require.EqualError(t, err, tc.expErrMsg)
					return
				}
				require.NoError(t, err)

				if assert.ObjectsAreEqual(expectedEvent, event) {
					matches += 1
				}
			}

			require.Equal(t, 1, matches)
		})
	}
}
