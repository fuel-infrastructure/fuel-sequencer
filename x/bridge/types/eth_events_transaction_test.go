package types_test

import (
	"fmt"
	"math"
	"testing"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
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
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      1,
			},
			expectedEqual: true,
		},
		{
			name:          "Unequal events tx - one is nil the other is not",
			eventTx1:      &testtypes.TestEthEventsTx,
			eventTx2:      nilEthEventsTx,
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - events list is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           []*sidecartypes.Event{testtypes.TestEvent1, testtypes.TestEvent2},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
				BlockNumber:      1,
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - AdvanceSequencer is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: false,
				NewEthereumBlock: true,
				BlockNumber:      1,
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - NewEthereumBlock is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: false,
				BlockNumber:      1,
			},
			expectedEqual: false,
		},
		{
			name:     "Unequal events tx - BlockNumber is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &types.EthEventsTx{
				Events:           testtypes.TestEvents,
				AdvanceSequencer: true,
				NewEthereumBlock: false,
				BlockNumber:      0,
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
			eventTx: &testtypes.TestEthEventsTx,
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
						EventType:       sidecartypes.DepositEventName,
						ContractAddress: testtypes.TestEthereumProxyContractAddress,
						Data:            []byte("invalid-data"),
					},
					testtypes.TestEvent1,
					testtypes.TestEvent2,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
			},
			expErrMsg: fmt.Sprintf("could not unmarshal to %s:", sidecartypes.DepositEventName),
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

func TestEthEventsTx_ValidateStateful(t *testing.T) {
	var nilEthEventsTx *types.EthEventsTx = nil

	testCases := []struct {
		name                         string
		eventTx                      *types.EthEventsTx
		ethereumProxyContractAddress string
		expErrMsg                    string
	}{
		{
			name:                         "Valid events tx",
			eventTx:                      &testtypes.TestEthEventsTx,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
		},
		{
			name:                         "Invalid events tx - nil",
			eventTx:                      nilEthEventsTx,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg:                    "EthEventsTx is nil",
		},
		{
			name: "Invalid events tx - event with unexpected contract address in list",
			eventTx: &types.EthEventsTx{
				Events: []*sidecartypes.Event{
					testtypes.TestEvent1,
					testutils.MustGetSidecarEventFromParsedEvent(
						testtypes.TestDepositEvent2, "invalid-ethereum-proxy-contract-address",
					),
					testtypes.TestEvent2,
				},
				AdvanceSequencer: true,
				NewEthereumBlock: true,
			},
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg: fmt.Sprintf(
				"event contract_address does not match expected ethereum_proxy_contract_address; got %s, expected %s",
				"invalid-ethereum-proxy-contract-address",
				testtypes.TestEthereumProxyContractAddress,
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.eventTx.ValidateStateful(tc.ethereumProxyContractAddress)
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
	previousBlock := blockNumber - 1

	testCases := []struct {
		name             string
		eventTx          *types.EthEventsTx
		lastBlockSynced  uint64
		eventIndexOffset uint64
		expErrMsg        string
	}{
		// Valid transactions at the right height
		{
			name:             "valid full tx at right height",
			eventTx:          &testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid full tx at right height even if offset is non-zero",
			eventTx:          &testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid partial tx at right height",
			eventTx:          &testtypes.TestEthEventsTxPartial,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid partial tx at right height even if offset is non-zero",
			eventTx:          &testtypes.TestEthEventsTxPartial,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid empty tx at right height",
			eventTx:          &testtypes.TestEthEventsTxWithoutEvents,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid NoNewBlock tx at right height",
			eventTx:          &testtypes.TestEthEventsTxNoNewBlock,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		// Invalid transactions with AdvanceSequencer false
		{
			name:             "invalid tx with AdvanceSequencer false",
			eventTx:          &testtypes.TestEthEventsTxSidecarErr,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
			expErrMsg:        "expected AdvanceSequencer to be true",
		},
		// Invalid transactions with no events when there's a non-zero offset
		{
			name:             "invalid tx with no events when there's a non-zero offset",
			eventTx:          &testtypes.TestEthEventsTxWithoutEvents,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		{
			name:             "invalid tx with no new block when there's a non-zero offset",
			eventTx:          &testtypes.TestEthEventsTxNoNewBlock,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		// Invalid transactions with wrong height
		{
			name:             "invalid tx at height in the future",
			eventTx:          &testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock - 1,
			eventIndexOffset: 0,
			expErrMsg:        "expected block number 0, got 1 in EthEventsTx",
		},
		{
			name:             "invalid tx at height in the past",
			eventTx:          &testtypes.TestEthEventsTx,
			lastBlockSynced:  previousBlock + 1,
			eventIndexOffset: 0,
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

// TestCorrelationBetweenNumberOfEventsWithMaxBytesAndSize ensures that if the size of testtypes.TestEthEventsTx
// changes, we get a failed test. If this test fails, it is very important to re-evaluate whether the
// NumberOfEventsWithMaxBytes function is still correctly implemented since this should mirror the Size function.
func TestCorrelationBetweenNumberOfEventsWithMaxBytesAndSize(t *testing.T) {

	tx := testtypes.TestEthEventsTx

	require.EqualValues(t, 527, tx.Size())
	require.EqualValues(t, 3, tx.NumberOfEventsWithMaxBytes(527)) // just enough bytes
	require.EqualValues(t, 2, tx.NumberOfEventsWithMaxBytes(526)) // just under enough
}

// TestCorrelationBetweenSizeAndMarshalling checks that marshalling TestEthEventsTx yields the expected number of bytes.
// This is important because we use the size to determine the number of events to trim, and the transaction selector
// uses size of the marshalled transaction to determine whether to include the transaction in the block. If there is a
// mismatch between these two, we can accidentally over-trim or under-trim the events from EthEventsTx.
func TestCorrelationBetweenSizeAndMarshalling(t *testing.T) {

	tx := testtypes.TestEthEventsTx

	bz, err := tx.Marshal()
	require.NoError(t, err)

	require.EqualValues(t, tx.Size(), len(bz))
}

func TestEthEventsTx_NumberOfEventsWithMaxBytes(t *testing.T) {

	testCases := []struct {
		name           string
		eventTx        *types.EthEventsTx
		maxBytes       uint64
		expNumOfEvents int
		expErrMsg      string
	}{
		{
			name:           "large max bytes fits all events",
			eventTx:        &testtypes.TestEthEventsTx,
			maxBytes:       math.MaxInt64,
			expNumOfEvents: len(testtypes.TestEthEventsTx.Events),
		},
		{
			name:           "exact size fits all events",
			eventTx:        &testtypes.TestEthEventsTx,
			maxBytes:       uint64(testtypes.TestEthEventsTx.Size()),
			expNumOfEvents: len(testtypes.TestEthEventsTx.Events),
		},
		{
			name:           "just under exact size fits n-1 events",
			eventTx:        &testtypes.TestEthEventsTx,
			maxBytes:       uint64(testtypes.TestEthEventsTx.Size()) - 1,
			expNumOfEvents: len(testtypes.TestEthEventsTx.Events) - 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			numOfEvents := tc.eventTx.NumberOfEventsWithMaxBytes(tc.maxBytes)
			require.EqualValues(t, tc.expNumOfEvents, numOfEvents)
		})
	}
}

func TestEthEventsTx_TrimEventsFromHead(t *testing.T) {

	testCases := []struct {
		name            string
		eventTx         types.EthEventsTx // do not use a pointer, since the function modifies in-place
		numEventsToTrim uint64
		expEventTx      types.EthEventsTx
		expErrMsg       string
	}{
		{
			name:            "trim none => same EthEventsTx",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToTrim: 0,
			expEventTx:      testtypes.TestEthEventsTx,
		},
		{
			name: "trim one => trimmed EthEventsTx",
			eventTx: types.EthEventsTx{
				Events:           testtypes.TestEthEventsTx.Events,
				AdvanceSequencer: testtypes.TestEthEventsTx.AdvanceSequencer,
				NewEthereumBlock: testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:      testtypes.TestEthEventsTx.BlockNumber,
			},
			numEventsToTrim: 1,
			expEventTx: types.EthEventsTx{
				Events:           testtypes.TestEthEventsTx.Events[1:], // 1 trimmed
				AdvanceSequencer: testtypes.TestEthEventsTx.AdvanceSequencer,
				NewEthereumBlock: testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:      testtypes.TestEthEventsTx.BlockNumber,
			},
		},
		{
			name:            "trim all is not possible",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToTrim: uint64(len(testtypes.TestEthEventsTx.Events)),
			expErrMsg:       "cannot trim all 3 events from EthEventsTx",
		},
		{
			name:            "trim more than all is not possible",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToTrim: uint64(len(testtypes.TestEthEventsTx.Events)) + 1,
			expErrMsg:       "insufficient no of events, expected at least 4 got 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.eventTx.TrimEventsFromHead(tc.numEventsToTrim)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			equal, err := tc.expEventTx.Equal(&tc.eventTx)
			require.NoError(t, err)
			require.True(t, equal)
		})
	}
}

func TestEthEventsTx_KeepEventsFromHead(t *testing.T) {

	testCases := []struct {
		name            string
		eventTx         types.EthEventsTx // do not use a pointer, since the function modifies in-place
		numEventsToKeep uint64
		expTrimmed      uint64
		expEventTx      types.EthEventsTx
		expErrMsg       string
	}{
		{
			name:            "keep all => same EthEventsTx",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)),
			expEventTx:      testtypes.TestEthEventsTx,
		},
		{
			name: "keep all but one => same EthEventsTx",
			eventTx: types.EthEventsTx{
				Events:           testtypes.TestEthEventsTx.Events[:3],
				AdvanceSequencer: testtypes.TestEthEventsTx.AdvanceSequencer,
				NewEthereumBlock: true, // will become false
				BlockNumber:      testtypes.TestEthEventsTx.BlockNumber,
			},
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)) - 1,
			expTrimmed:      1,
			expEventTx: types.EthEventsTx{
				Events:           testtypes.TestEthEventsTx.Events[:2], // 1 trimmed
				AdvanceSequencer: testtypes.TestEthEventsTx.AdvanceSequencer,
				NewEthereumBlock: false, // becomes false
				BlockNumber:      testtypes.TestEthEventsTx.BlockNumber,
			},
		},
		{
			name:            "keep none is not possible",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToKeep: 0,
			expErrMsg:       "cannot trim all 3 events from EthEventsTx",
		},
		{
			name:            "keep more than all is not possible",
			eventTx:         testtypes.TestEthEventsTx,
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)) + 1,
			expErrMsg:       "insufficient no of events, expected at least 4 got 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lenBefore := len(tc.eventTx.Events)

			trimmed, err := tc.eventTx.KeepEventsFromHead(tc.numEventsToKeep)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expTrimmed, trimmed)

			lenAfter := len(tc.eventTx.Events)
			require.EqualValues(t, tc.expTrimmed, lenBefore-lenAfter)

			equal, err := tc.expEventTx.Equal(&tc.eventTx)
			require.NoError(t, err)
			require.True(t, equal)
		})
	}
}
