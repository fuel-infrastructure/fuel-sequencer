package types_test

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func TestParsedEvent_ValidateBasic(t *testing.T) {
	var nilDepositEvent *types.DepositEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name      string
		event     types.ParsedEvent
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
			event: &types.DepositEvent{
				Depositor: "invalid-depositor",
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "DepositEvent - invalid recipient - NO ERROR",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: "invalid-recipient",
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestLockup1,
			},
		},
		{
			name: "DepositEvent - invalid lockup - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    "0.23523",
			},
			expErrMsg: "could not convert lockup to a valid sdk.Int",
		},
		{
			name: "DepositEvent - amount is zero - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0",
				Lockup:    testtypes.TestLockup1,
			},
			expErrMsg: "amount must be bigger than zero",
		},
		{
			name: "DepositEvent - amount is float - error",
			event: &types.DepositEvent{
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
			event: &types.AuthorizeEvent{
				Sender: "invalid-sender",
				Data:   testutils.MustHexDecodeString(testtypes.TestData1),
			},
			expErrMsg: "sender is not a valid hex address",
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
		event      types.ParsedEvent
		getExpMsgs func() []*codectypes.Any
		expErrMsg  string
	}{
		{
			name:  "DepositEvent",
			event: testtypes.TestDepositEvent1,
			getExpMsgs: func() []*codectypes.Any {
				anys, err := types.NewAnysWithValue(
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
			msgs, err := tc.event.Messages(testtypes.TestCdc, testtypes.TestGovernanceAddress)
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
