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

			require.Equal(t, event, tc.expectedEvent)
		})
	}
}

func TestEvent_Validate(t *testing.T) {
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
			name: "Invalid event - contract address not the correct one",
			event: testutils.MustGetSidecarEventFromParsedEvent(
				testtypes.TestAuthorizeEvent3, "0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97",
			),
			expErrMsg: "event's contract address does not match expected proxy contract address",
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
			err := tc.event.Validate(testtypes.TestEthereumProxyContractAddress)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
