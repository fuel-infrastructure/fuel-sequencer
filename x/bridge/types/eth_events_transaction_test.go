package types_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
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
				BlockNumber:      sdkmath.OneInt(),
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
				BlockNumber:      sdkmath.OneInt(),
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
				BlockNumber:      sdkmath.OneInt(),
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
				BlockNumber:      sdkmath.OneInt(),
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - BlockNumber is different",
			eventTx1: testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: false,
				BlockNumber:      sdkmath.ZeroInt(),
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

func TestEthEventsTx_ValidateBeforeProcessing(t *testing.T) {

	// Calculate the LastEthereumBlockSynced that we expect when submitting TestEthEventsTx
	blockNumber := testtypes.TestEthEventsTx.BlockNumber
	previousBlock := blockNumber.Sub(sdkmath.OneInt())

	testCases := []struct {
		name             string
		eventTx          *types.EthEventsTx
		lastBlockSynced  sdkmath.Int
		eventIndexOffset sdkmath.Int
		expErrMsg        string
	}{
		// Valid transactions at the right height
		{
			name:             "valid full tx at right height",
			eventTx:          testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.ZeroInt(),
		},
		{
			name:             "valid full tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.NewInt(10),
		},
		{
			name:             "valid partial tx at right height",
			eventTx:          testtypes.TestEthEventsTxPartial,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.ZeroInt(),
		},
		{
			name:             "valid partial tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestEthEventsTxPartial,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.NewInt(10),
		},
		{
			name:             "valid empty tx at right height",
			eventTx:          testtypes.TestEthEventsTxWithoutEvents,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.ZeroInt(),
		},
		{
			name:             "valid NoNewBlock tx at right height",
			eventTx:          testtypes.TestEthEventsTxNoNewBlock,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.ZeroInt(),
		},
		// Invalid transactions with AdvanceSequencer false
		{
			name:             "invalid tx with AdvanceSequencer false",
			eventTx:          testtypes.TestEthEventsTxSidecarErr,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.ZeroInt(),
			expErrMsg:        "expected AdvanceSequencer to be true",
		},
		// Invalid transactions with no events when there's a non-zero offset
		{
			name:             "invalid tx with no events when there's a non-zero offset",
			eventTx:          testtypes.TestEthEventsTxWithoutEvents,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.NewInt(10),
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		{
			name:             "invalid tx with no new block when there's a non-zero offset",
			eventTx:          testtypes.TestEthEventsTxNoNewBlock,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: sdkmath.NewInt(10),
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		// Invalid transactions with wrong height
		{
			name:             "invalid tx at height in the future",
			eventTx:          testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock.SubRaw(1),
			eventIndexOffset: sdkmath.ZeroInt(),
			expErrMsg:        "expected block number 0, got 1 in EthEventsTx",
		},
		{
			name:             "invalid tx at height in the past",
			eventTx:          testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock.AddRaw(1),
			eventIndexOffset: sdkmath.ZeroInt(),
			expErrMsg:        "expected block number 2, got 1 in EthEventsTx",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.eventTx.ValidateBeforeProcessing(tc.lastBlockSynced, tc.eventIndexOffset)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
