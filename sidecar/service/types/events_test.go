package types_test

import (
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"
)

func TestSendToSequencerEvent_Equal(t *testing.T) {
	var nilSendToSequencerEvent *types.SendToSequencerEvent = nil
	var nilAuthorizeEvent *types.AuthorizeEvent = nil

	testCases := []struct {
		name          string
		event1        types.ConcreteEvent
		event2        types.ConcreteEvent
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
