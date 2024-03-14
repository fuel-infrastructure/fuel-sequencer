package types_test

import (
	"fmt"
	"testing"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestEthEventsTx_Equal(t *testing.T) {
	var nilEthEventsTx *types.EthEventsTx = nil

	testCases := []struct {
		name          string
		eventTx1      *types.EthEventsTx
		eventTx2      *types.EthEventsTx
		expectedEqual bool
		expErrMsg     string
	}{
		{
			name:          "Equal events tx - both nil",
			eventTx1:      nilEthEventsTx,
			eventTx2:      nilEthEventsTx,
			expectedEqual: true,
		},
		{
			name:     "Equal events tx - not nil",
			eventTx1: testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: true,
			},
			expectedEqual: true,
		},
		{
			name:          "Unequal events tx - one is nil the other is not",
			eventTx1:      testtypes.TestEthEventsTx,
			eventTx2:      nilEthEventsTx,
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - events list is different",
			eventTx1: testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           []*sidecartypes.Event{testtypes.TestEvent1, testtypes.TestEvent2},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - AdvanceSequencer is different",
			eventTx1: testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: false,
				NewEthereumBlock: true,
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - NewEthereumBlock is different",
			eventTx1: testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: false,
			},
			expectedEqual: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualEqual, err := tc.eventTx1.Equal(tc.eventTx2)

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

func TestEthEventsTx_ValidateBasic(t *testing.T) {
	var nilEthEventsTx *types.EthEventsTx = nil

	testCases := []struct {
		name      string
		eventTx   *types.EthEventsTx
		expErrMsg string
	}{
		{
			name:    "Valid events tx",
			eventTx: testtypes.TestEthEventsTx,
		},
		{
			name:      "Invalid events tx - nil",
			eventTx:   nilEthEventsTx,
			expErrMsg: "EthEventsTx is nil",
		},
		{
			name: "Invalid events tx - invalid event in list",
			eventTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					{
						EventType: sidecartypes.SendToSequencerEventName,
						Data:      []byte("invalid-data"),
					},
					testtypes.TestEvent1,
					testtypes.TestEvent2,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
			},
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", sidecartypes.SendToSequencerEventName),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.eventTx.ValidateBasic()
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
