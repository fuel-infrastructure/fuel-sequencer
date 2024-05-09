package types_test

import (
	"fmt"
	"testing"

	"github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalParsedEvent(t *testing.T) {
	testCases := []struct {
		name          string
		event         *types.Event
		expectedEvent types.ParsedEvent
		expErrMsg     string
	}{
		{
			name:          "DepositEvent",
			event:         testtypes.TestEvent1,
			expectedEvent: testtypes.TestDepositEvent3,
		},
		{
			name:          "AuthorizeEvent",
			event:         testtypes.TestEvent2,
			expectedEvent: testtypes.TestAuthorizeEvent3,
		},
		{
			name: "Unknown event type",
			event: &types.Event{
				EventType:       "invalid-event",
				Data:            []byte{},
				ContractAddress: testtypes.TestEthereumProxyContractAddress,
			},
			expErrMsg: "unknown event type: invalid-event",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event, err := tc.event.UnmarshalParsedEvent()
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			require.True(t, event.Equal(tc.expectedEvent))
		})
	}
}

func TestEvent_Equal(t *testing.T) {
	var nilEvent *types.Event = nil

	testCases := []struct {
		name          string
		event1        *types.Event
		event2        *types.Event
		expectedEqual bool
		expErrMsg     string
	}{
		{
			name:          "Equal events - both nil",
			event1:        nilEvent,
			event2:        nilEvent,
			expectedEqual: true,
		},
		{
			name:   "Equal events - Deposit",
			event1: testtypes.TestEvent1,
			event2: testutils.MustGetSidecarEventFromParsedEvent(
				testtypes.TestDepositEvent3, testtypes.TestEthereumProxyContractAddress,
			),
			expectedEqual: true,
		},
		{
			name:   "Equal events - AuthorizeEvent",
			event1: testtypes.TestEvent2,
			event2: testutils.MustGetSidecarEventFromParsedEvent(
				testtypes.TestAuthorizeEvent3, testtypes.TestEthereumProxyContractAddress,
			),
			expectedEqual: true,
		},
		{
			name:          "Unequal events - one is nil the other is not",
			event1:        testtypes.TestEvent1,
			event2:        nilEvent,
			expectedEqual: false,
		},
		{
			name:          "Unequal events - events with different types",
			event1:        testtypes.TestEvent1,
			event2:        testtypes.TestEvent2,
			expectedEqual: false,
		},
		{
			name:   "Unequal events - events belonging to a different contract",
			event1: testtypes.TestEvent1,
			event2: testutils.MustGetSidecarEventFromParsedEvent(
				testtypes.TestDepositEvent3, "different-contract-address",
			),
			expectedEqual: false,
		},
		{
			name: "Error - event1 cannot be unmarshalled",
			event1: &types.Event{
				EventType:       types.AuthorizeEventName,
				Data:            []byte("invalid-data"),
				ContractAddress: testtypes.TestEthereumProxyContractAddress,
			},
			event2:    testtypes.TestEvent2,
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", types.AuthorizeEventName),
		},
		{
			name:   "Error - event2 cannot be unmarshalled",
			event1: testtypes.TestEvent2,
			event2: &types.Event{
				EventType:       types.AuthorizeEventName,
				Data:            []byte("invalid-data"),
				ContractAddress: testtypes.TestEthereumProxyContractAddress,
			},
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", types.AuthorizeEventName),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualEqual, err := tc.event1.Equal(tc.event2)

			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			if tc.expectedEqual {
				require.True(t, actualEqual)
			} else {
				require.False(t, actualEqual)
			}
		})
	}
}

func TestEvent_ValidateBasic(t *testing.T) {
	var nilEvent *types.Event = nil

	testCases := []struct {
		name      string
		event     *types.Event
		expErrMsg string
	}{
		{
			name:  "Valid event - DepositEvent",
			event: testtypes.TestEvent1,
		},
		{
			name:  "Valid event - AuthorizeEvent",
			event: testtypes.TestEvent2,
		},
		{
			name:      "Invalid event - nil",
			event:     nilEvent,
			expErrMsg: "event is nil",
		},
		{
			name: "Invalid event - contract address not in the right format",

			// Ethereum addresses are strictly 40 chars long. Adding more characters should make ValidateBasic error.
			event: testutils.MustGetSidecarEventFromParsedEvent(
				testtypes.TestAuthorizeEvent3, testtypes.TestEthereumProxyContractAddress+"Ab",
			),
			expErrMsg: "contract_address is not a valid hex address",
		},
		{
			name: "Invalid event - DepositEvent cannot be unmarshalled",
			event: &types.Event{
				EventType:       types.DepositEventName,
				Data:            []byte("invalid-data"),
				ContractAddress: testtypes.TestEthereumProxyContractAddress,
			},
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", types.DepositEventName),
		},
		{
			name: "Invalid event - AuthorizeEvent cannot be unmarshalled",
			event: &types.Event{
				EventType:       types.AuthorizeEventName,
				Data:            []byte("invalid-data"),
				ContractAddress: testtypes.TestEthereumProxyContractAddress,
			},
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", types.AuthorizeEventName),
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
