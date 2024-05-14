package types_test

import (
	"math"
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
		eventTx1      *testtypes.TestEthEventsTxWithEvents
		eventTx2      *testtypes.TestEthEventsTxWithEvents
		expectedEqual bool
		expErrMsg     string
	}{
		{
			name: "Equal events tx - both nil",
			eventTx1: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: nilEthEventsTx,
				Events:      nil,
			},
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: nilEthEventsTx,
				Events:      nil,
			},
		},
		{
			name:     "Equal events tx - not nil",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: &types.EthEventsTx{
					Authority:         testtypes.TestGovernanceAddress,
					NumInjectedEvents: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:  true,
					BlockNumber:       1,
				},
				Events: testtypes.TestEvents,
			},
		},
		{
			name:     "Unequal events tx - one is nil the other is not",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: nilEthEventsTx,
				Events:      nil,
			},
			expErrMsg: "nil (false) != (true)",
		},
		{
			name:     "Unequal events tx - events list is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: &types.EthEventsTx{
					Authority:         testtypes.TestGovernanceAddress,
					NumInjectedEvents: 2,
					NewEthereumBlock:  true,
					BlockNumber:       1,
				},
				Events: []*sidecartypes.Event{testtypes.TestEvent1, testtypes.TestEvent2},
			},
			expErrMsg: "number of injected events (3) != (2)",
		},
		{
			name:     "Unequal events tx - NewEthereumBlock is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: &types.EthEventsTx{
					Authority:         testtypes.TestGovernanceAddress,
					NumInjectedEvents: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:  false,
					BlockNumber:       1,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "new Ethereum block (true) != (false)",
		},
		{
			name:     "Unequal events tx - BlockNumber is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: &types.EthEventsTx{
					Authority:         testtypes.TestGovernanceAddress,
					NumInjectedEvents: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:  true,
					BlockNumber:       0,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "block number (1) != (0)",
		},
		{
			name:     "Unequal events tx - Authority is different",
			eventTx1: &testtypes.TestEthEventsTx,
			eventTx2: &testtypes.TestEthEventsTxWithEvents{
				EthEventsTx: &types.EthEventsTx{
					Authority:         "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4",
					NumInjectedEvents: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:  true,
					BlockNumber:       1,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "authority (fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3) != (fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			eventTxs1 := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, tc.eventTx1.Events)
			eventTxs2 := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, tc.eventTx2.Events)

			err := tc.eventTx1.EthEventsTx.Equal(tc.eventTx2.EthEventsTx, eventTxs1, eventTxs2)

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
			eventTx:          testtypes.TestEthEventsTx.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid full tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestEthEventsTx.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid partial tx at right height",
			eventTx:          testtypes.TestEthEventsTxPartial.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid partial tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestEthEventsTxPartial.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid empty tx at right height",
			eventTx:          testtypes.TestEthEventsTxWithoutEvents.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid NoNewBlock tx at right height",
			eventTx:          testtypes.TestEthEventsTxNoNewBlock.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		// Invalid transactions with no events when there's a non-zero offset
		{
			name:             "invalid tx with no events when there's a non-zero offset",
			eventTx:          testtypes.TestEthEventsTxWithoutEvents.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		{
			name:             "invalid tx with no new block when there's a non-zero offset",
			eventTx:          testtypes.TestEthEventsTxNoNewBlock.EthEventsTx,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		// Invalid transactions with wrong height
		{
			name:             "invalid tx at height in the future",
			eventTx:          testtypes.TestEthEventsTx.EthEventsTx,
			lastBlockSynced:  previousBlock - 1,
			eventIndexOffset: 0,
			expErrMsg:        "expected block number 0, got 1 in EthEventsTx",
		},
		{
			name:             "invalid tx at height in the past",
			eventTx:          testtypes.TestEthEventsTx.EthEventsTx,
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

func TestCorrelationBetweenNumberOfEventsWithMaxBytesAndRawTxBytes(t *testing.T) {

	tx := testtypes.TestEthEventsTx
	txRawBytes, err := tx.RawTxBytes()
	require.NoError(t, err)

	events := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, testtypes.TestEthEventsTx.Events)
	eventsSize := testtypes.MustGetSizeFromEvents(testtypes.TestCdc, testtypes.TestEthEventsTx.Events)

	totalSize := len(txRawBytes) + eventsSize

	numEvents, err := tx.NumberOfEventsWithMaxBytes(events, uint64(totalSize))
	require.NoError(t, err)
	require.EqualValues(t, 3, numEvents) // just enough bytes

	numEvents, err = tx.NumberOfEventsWithMaxBytes(events, uint64(totalSize-1))
	require.NoError(t, err)
	require.EqualValues(t, 2, numEvents) // just under enough
}

func TestEthEventsTx_NumberOfEventsWithMaxBytes(t *testing.T) {

	tx := testtypes.TestEthEventsTx.EthEventsTx
	txRawBytes, err := tx.RawTxBytes()
	require.NoError(t, err)

	events := testtypes.TestEthEventsTx.Events
	rawEvents := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, events)
	eventsSize := testtypes.MustGetSizeFromEvents(testtypes.TestCdc, events)

	txAndEventsSize := len(txRawBytes) + eventsSize

	testCases := []struct {
		name           string
		eventTx        *testtypes.TestEthEventsTxWithEvents
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
			maxBytes:       uint64(txAndEventsSize),
			expNumOfEvents: len(testtypes.TestEthEventsTx.Events),
		},
		{
			name:           "just under exact size fits n-1 events",
			eventTx:        &testtypes.TestEthEventsTx,
			maxBytes:       uint64(txAndEventsSize) - 1,
			expNumOfEvents: len(testtypes.TestEthEventsTx.Events) - 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			numOfEvents, err := tc.eventTx.NumberOfEventsWithMaxBytes(rawEvents, tc.maxBytes)

			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expNumOfEvents, numOfEvents)
		})
	}
}

func TestEthEventsTx_TrimEventsFromHead(t *testing.T) {

	events := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, testtypes.TestEthEventsTx.Events)

	testCases := []struct {
		name            string
		eventTx         types.EthEventsTx // do not use a pointer, since the function modifies in-place
		events          [][]byte
		numEventsToTrim uint64
		expEventTx      types.EthEventsTx
		expEvents       [][]byte
		expErrMsg       string
	}{
		{
			name: "trim none => same EthEventsTx",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events,
			numEventsToTrim: 0,
			expEventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			expEvents: events,
		},
		{
			name: "trim one => trimmed EthEventsTx",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events,
			numEventsToTrim: 1,
			expEventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents - 1, // 1 trimmed
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			expEvents: events[1:], // 1 trimmed
		},
		{
			name: "trim all is not possible",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			numEventsToTrim: uint64(len(testtypes.TestEthEventsTx.Events)),
			expErrMsg:       "cannot trim all 3 events",
		},
		{
			name: "trim more than all is not possible",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			numEventsToTrim: uint64(len(testtypes.TestEthEventsTx.Events)) + 1,
			expErrMsg:       "insufficient no of events, expected at least 4 got 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			trimmedEvents, err := tc.eventTx.TrimEventsFromHead(tc.events, tc.numEventsToTrim)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)

			err = tc.expEventTx.Equal(&tc.eventTx, tc.expEvents, trimmedEvents)
			require.NoError(t, err)
		})
	}
}

func TestEthEventsTx_KeepEventsFromHead(t *testing.T) {

	events := testtypes.MustGetRawTxBytesFromEvents(testtypes.TestCdc, testtypes.TestEthEventsTx.Events)

	testCases := []struct {
		name            string
		eventTx         types.EthEventsTx // do not use a pointer, since the function modifies in-place
		events          [][]byte
		numEventsToKeep uint64
		expTrimmed      uint64
		expEventTx      types.EthEventsTx
		expEvents       [][]byte
		expErrMsg       string
	}{
		{
			name: "keep all => same EthEventsTx",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events,
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)),
			expEventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			expEvents: events,
		},
		{
			name: "keep all but one => same EthEventsTx",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: 3,
				NewEthereumBlock:  true, // will become false
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events[:3],
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)) - 1,
			expTrimmed:      1,
			expEventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: 2,     // 1 trimmed
				NewEthereumBlock:  false, // becomes false
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			expEvents: events[:2], // 1 trimmed
		},
		{
			name: "keep none is not possible",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events,
			numEventsToKeep: 0,
			expErrMsg:       "cannot trim all 3 events",
		},
		{
			name: "keep more than all is not possible",
			eventTx: types.EthEventsTx{
				Authority:         testtypes.TestEthEventsTx.Authority,
				NumInjectedEvents: testtypes.TestEthEventsTx.NumInjectedEvents,
				NewEthereumBlock:  testtypes.TestEthEventsTx.NewEthereumBlock,
				BlockNumber:       testtypes.TestEthEventsTx.BlockNumber,
			},
			events:          events,
			numEventsToKeep: uint64(len(testtypes.TestEthEventsTx.Events)) + 1,
			expErrMsg:       "insufficient no of events, expected at least 4 got 3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lenBefore := len(tc.events)

			trimmedEvents, trimmed, err := tc.eventTx.KeepEventsFromHead(tc.events, tc.numEventsToKeep)
			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expTrimmed, trimmed)

			lenAfter := len(trimmedEvents)
			require.EqualValues(t, tc.expTrimmed, lenBefore-lenAfter)

			err = tc.expEventTx.Equal(&tc.eventTx, tc.expEvents, trimmedEvents)
			require.NoError(t, err)
		})
	}
}
