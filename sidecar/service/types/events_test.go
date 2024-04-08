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

	expect := crypto.Keccak256Hash([]byte("SendToSequencerEvent(address,uint256,string,uint256)")).Hex()
	actual := types.SendToSequencerEventHashFn
	require.Equal(t, expect, actual)

	expect = crypto.Keccak256Hash([]byte("AuthorizeEvent(address,bytes)")).Hex()
	actual = types.AuthorizeEventHashFn
	require.Equal(t, expect, actual)
}

func TestParsedEvent_Equal(t *testing.T) {
	var nilSendToSequencerEvent *types.SendToSequencerEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name          string
		event1        types.ParsedEvent
		event2        types.ParsedEvent
		expectedEqual bool
	}{
		{
			name:          "SendToSequencerEvent - equal events - both nil",
			event1:        nilSendToSequencerEvent,
			event2:        nilSendToSequencerEvent,
			expectedEqual: true,
		},
		{
			name:   "SendToSequencerEvent - equal events - both not nil",
			event1: testtypes.TestSendToSequencerEvent1,
			event2: &types.SendToSequencerEvent{
				From:     testtypes.TestFrom1,
				Amount:   testtypes.TestAmount1,
				To:       testtypes.TestTo1,
				Duration: testtypes.TestDuration1,
			},
			expectedEqual: true,
		},
		{
			name:          "SendToSequencerEvent - unequal events - one is nil the other is not",
			event1:        testtypes.TestSendToSequencerEvent1,
			event2:        nilSendToSequencerEvent,
			expectedEqual: false,
		},
		{
			name:          "SendToSequencerEvent - unequal events - both not nil",
			event1:        testtypes.TestSendToSequencerEvent1,
			event2:        testtypes.TestSendToSequencerEvent2,
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
				From:    testtypes.TestFrom1,
				Message: testutils.MustHexDecodeString(testtypes.TestMessage1),
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
	var nilSendToSequencerEvent *types.SendToSequencerEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name      string
		event     types.ParsedEvent
		expErrMsg string
	}{
		{
			name:  "SendToSequencerEvent - valid - to not empty",
			event: testtypes.TestSendToSequencerEvent1,
		},
		{
			name:  "SendToSequencerEvent - valid - to empty",
			event: testtypes.TestSendToSequencerEvent2,
		},
		{
			name:      "SendToSequencerEvent - nil receiver - error",
			event:     nilSendToSequencerEvent,
			expErrMsg: "SendToSequencerEvent is nil",
		},
		{
			name: "SendToSequencerEvent - invalid from - error",
			event: &types.SendToSequencerEvent{
				From:     "invalid-from",
				Amount:   testtypes.TestAmount1,
				To:       testtypes.TestTo1,
				Duration: testtypes.TestDuration1,
			},
			expErrMsg: "from is not a valid hex address",
		},
		{
			name: "SendToSequencerEvent - invalid to - error",
			event: &types.SendToSequencerEvent{
				From:     testtypes.TestFrom1,
				Amount:   testtypes.TestAmount1,
				To:       "invalid-to",
				Duration: testtypes.TestDuration1,
			},
			expErrMsg: "to is not a valid Bech32 address",
		},
		{
			name: "SendToSequencerEvent - invalid duration - error",
			event: &types.SendToSequencerEvent{
				From:     testtypes.TestFrom1,
				Amount:   testtypes.TestAmount1,
				To:       testtypes.TestTo1,
				Duration: "0.23523",
			},
			expErrMsg: "could not convert duration to a valid sdk.Int",
		},
		{
			name: "SendToSequencerEvent - amount is zero - error",
			event: &types.SendToSequencerEvent{
				From:     testtypes.TestFrom1,
				Amount:   "0",
				To:       testtypes.TestTo1,
				Duration: testtypes.TestDuration1,
			},
			expErrMsg: "amount must be bigger than zero",
		},
		{
			name: "SendToSequencerEvent - amount is float - error",
			event: &types.SendToSequencerEvent{
				From:     testtypes.TestFrom1,
				Amount:   "0.4356346",
				To:       testtypes.TestTo1,
				Duration: testtypes.TestDuration1,
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
			name: "AuthorizeEvent - invalid from - error",
			event: &types.AuthorizeEvent{
				From:    "invalid-from",
				Message: testutils.MustHexDecodeString(testtypes.TestMessage1),
			},
			expErrMsg: "from is not a valid hex address",
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
