package types_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"
)

func TestEventHashFnsAreAsExpected(t *testing.T) {

	expect := crypto.Keccak256Hash([]byte("Deposit(address,address,uint256,uint256)")).Hex()
	actual := types.DepositEventHashFn
	require.Equal(t, expect, actual)

	expect = crypto.Keccak256Hash([]byte("Authorize(address,bytes)")).Hex()
	actual = types.AuthorizeEventHashFn
	require.Equal(t, expect, actual)
}

func TestParsedEvent_Equal(t *testing.T) {
	var nilDepositEvent *types.DepositEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name          string
		event1        types.ParsedEvent
		event2        types.ParsedEvent
		expectedEqual bool
	}{
		{
			name:          "DepositEvent - equal events - both nil",
			event1:        nilDepositEvent,
			event2:        nilDepositEvent,
			expectedEqual: true,
		},
		{
			name:   "DepositEvent - equal events - both not nil",
			event1: testtypes.TestDepositEvent1,
			event2: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestDuration1,
			},
			expectedEqual: true,
		},
		{
			name:          "DepositEvent - unequal events - one is nil the other is not",
			event1:        testtypes.TestDepositEvent1,
			event2:        nilDepositEvent,
			expectedEqual: false,
		},
		{
			name:          "DepositEvent - unequal events - both not nil",
			event1:        testtypes.TestDepositEvent1,
			event2:        testtypes.TestDepositEvent2,
			expectedEqual: false,
		},
		{
			name:          "AuthorizeEvent - equal events - both nil",
			event1:        nilAuthorizeEvent,
			event2:        nilAuthorizeEvent,
			expectedEqual: true,
		},
		{
			name:   "AuthorizeEvent - equal events - both not nil",
			event1: testtypes.TestAuthorizeEvent1,
			event2: &types.AuthorizeEvent{
				Sender: testtypes.TestFrom1,
				Data:   testutils.MustHexDecodeString(testtypes.TestMessage1),
			},
			expectedEqual: true,
		},
		{
			name:          "AuthorizeEvent - unequal events - one is nil the other is not",
			event1:        testtypes.TestAuthorizeEvent1,
			event2:        nilAuthorizeEvent,
			expectedEqual: false,
		},
		{
			name:          "AuthorizeEvent - unequal events - both not nil",
			event1:        testtypes.TestAuthorizeEvent1,
			event2:        testtypes.TestAuthorizeEvent2,
			expectedEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualEqual := tc.event1.Equal(tc.event2)
			if tc.expectedEqual {
				require.True(t, actualEqual)
			} else {
				require.False(t, actualEqual)
			}
		})
	}
}

func TestParsedEvent_ValidateBasic(t *testing.T) {
	var nilDepositEvent *types.DepositEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name      string
		event     types.ParsedEvent
		expErrMsg string
	}{
		{
			name:  "DepositEvent - valid - to is a Hex address",
			event: testtypes.TestDepositEvent1,
		},
		{
			name:  "DepositEvent - valid - to is a Sequencer address",
			event: testtypes.TestDepositEvent9,
		},
		{
			name:  "DepositEvent - valid - to is empty",
			event: testtypes.TestDepositEvent2,
		},
		{
			name:      "DepositEvent - nil receiver - error",
			event:     nilDepositEvent,
			expErrMsg: "DepositEvent is nil",
		},
		{
			name: "DepositEvent - invalid depositor - error",
			event: &types.DepositEvent{
				Depositor: "invalid-depositor",
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestDuration1,
			},
			expErrMsg: "depositor is not a valid hex address",
		},
		{
			name: "DepositEvent - invalid recipient - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: "invalid-recipient",
				Amount:    testtypes.TestAmount1,
				Lockup:    testtypes.TestDuration1,
			},
			expErrMsg: "recipient is not a valid Bech32 or Hex address",
		},
		{
			name: "DepositEvent - invalid duration - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    testtypes.TestAmount1,
				Lockup:    "0.23523",
			},
			expErrMsg: "could not convert duration to a valid sdk.Int",
		},
		{
			name: "DepositEvent - amount is zero - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0",
				Lockup:    testtypes.TestDuration1,
			},
			expErrMsg: "amount must be bigger than zero",
		},
		{
			name: "DepositEvent - amount is float - error",
			event: &types.DepositEvent{
				Depositor: testtypes.TestFrom1,
				Recipient: testtypes.TestTo1,
				Amount:    "0.4356346",
				Lockup:    testtypes.TestDuration1,
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
				Data:   testutils.MustHexDecodeString(testtypes.TestMessage1),
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
