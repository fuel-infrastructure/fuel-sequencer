package types_test

import (
	"math"
	"testing"

	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestMsgIndex_Equal(t *testing.T) {
	var nilMsgIndex *types.MsgIndex = nil

	testCases := []struct {
		name          string
		eventTx1      *testtypes.TestMsgIndexWithEvents
		eventTx2      *testtypes.TestMsgIndexWithEvents
		expectedEqual bool
		expErrMsg     string
	}{
		{
			name: "Equal MsgIndex - both nil",
			eventTx1: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: nilMsgIndex,
				Events:   nil,
			},
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: nilMsgIndex,
				Events:   nil,
			},
		},
		{
			name:     "Equal MsgIndex - not nil",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: &types.MsgIndex{
					Authority:           testtypes.TestGovernanceAddress,
					NumInjectedEventTxs: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:    true,
					BlockNumber:         1,
				},
				Events: testtypes.TestEvents,
			},
		},
		{
			name:     "Unequal MsgIndex - one is nil the other is not",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: nilMsgIndex,
				Events:   nil,
			},
			expErrMsg: "nil (false) != (true)",
		},
		{
			name:     "Unequal MsgIndex - number of injected event txs is different",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: &types.MsgIndex{
					Authority:           testtypes.TestGovernanceAddress,
					NumInjectedEventTxs: 2,
					NewEthereumBlock:    true,
					BlockNumber:         1,
				},
				Events: []*sidecartypes.Event{testtypes.TestEvent1, testtypes.TestEvent2},
			},
			expErrMsg: "number of injected event txs (3) != (2)",
		},
		{
			name:     "Unequal MsgIndex - NewEthereumBlock is different",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: &types.MsgIndex{
					Authority:           testtypes.TestGovernanceAddress,
					NumInjectedEventTxs: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:    false,
					BlockNumber:         1,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "new Ethereum block (true) != (false)",
		},
		{
			name:     "Unequal MsgIndex - BlockNumber is different",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: &types.MsgIndex{
					Authority:           testtypes.TestGovernanceAddress,
					NumInjectedEventTxs: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:    true,
					BlockNumber:         0,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "block number (1) != (0)",
		},
		{
			name:     "Unequal MsgIndex - Authority is different",
			eventTx1: &testtypes.TestMsgIndex,
			eventTx2: &testtypes.TestMsgIndexWithEvents{
				MsgIndex: &types.MsgIndex{
					Authority:           "fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4",
					NumInjectedEventTxs: uint64(len(testtypes.TestEvents)),
					NewEthereumBlock:    true,
					BlockNumber:         1,
				},
				Events: testtypes.TestEvents,
			},
			expErrMsg: "authority (fuelsequencer10d07y265gmmuvt4z0w9aw880jnsr700jdjfvk3) != (fuelsequencer1w8rk2mk84wytpxx7ld63kaqpkhmd39m05xlgt4)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			eventTxs1 := testtypes.MustGetEventTxsFromEvents(
				testtypes.TestCdc, testtypes.TestGovernanceAddress, tc.eventTx1.Events,
			)
			eventTxs2 := testtypes.MustGetEventTxsFromEvents(
				testtypes.TestCdc, testtypes.TestGovernanceAddress, tc.eventTx2.Events,
			)

			err := tc.eventTx1.MsgIndex.Equal(tc.eventTx2.MsgIndex, eventTxs1, eventTxs2)

			if len(tc.expErrMsg) > 0 {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgIndex_ValidateBeforeProcessing(t *testing.T) {

	// Calculate the LastEthereumBlockSynced that we expect when submitting TestMsgIndex
	blockNumber := testtypes.TestMsgIndex.BlockNumber
	previousBlock := blockNumber - 1

	testCases := []struct {
		name             string
		eventTx          *types.MsgIndex
		lastBlockSynced  uint64
		eventIndexOffset uint64
		expErrMsg        string
	}{
		// Valid transactions at the right height
		{
			name:             "valid full tx at right height",
			eventTx:          testtypes.TestMsgIndex.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid full tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestMsgIndex.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid partial tx at right height",
			eventTx:          testtypes.TestMsgIndexPartial.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid partial tx at right height even if offset is non-zero",
			eventTx:          testtypes.TestMsgIndexPartial.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
		},
		{
			name:             "valid empty tx at right height",
			eventTx:          testtypes.TestMsgIndexWithoutEvents.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		{
			name:             "valid NoNewBlock tx at right height",
			eventTx:          testtypes.TestMsgIndexNoNewBlock.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 0,
		},
		// Invalid transactions with no events when there's a non-zero offset
		{
			name:             "invalid tx with no events when there's a non-zero offset",
			eventTx:          testtypes.TestMsgIndexWithoutEvents.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		{
			name:             "invalid tx with no new block when there's a non-zero offset",
			eventTx:          testtypes.TestMsgIndexNoNewBlock.MsgIndex,
			lastBlockSynced:  previousBlock,
			eventIndexOffset: 10,
			expErrMsg:        "expected at least 1 new event if offset is non-zero (10)",
		},
		// Invalid transactions with wrong height
		{
			name:             "invalid tx at height in the future",
			eventTx:          testtypes.TestMsgIndex.MsgIndex,
			lastBlockSynced:  previousBlock - 1,
			eventIndexOffset: 0,
			expErrMsg:        "expected block number 0, got 1 in MsgIndex",
		},
		{
			name:             "invalid tx at height in the past",
			eventTx:          testtypes.TestMsgIndex.MsgIndex,
			lastBlockSynced:  previousBlock + 1,
			eventIndexOffset: 0,
			expErrMsg:        "expected block number 2, got 1 in MsgIndex",
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

	tx := testtypes.TestMsgIndex
	txRawBytes, err := tx.RawTxBytes()
	require.NoError(t, err)

	events := testtypes.MustGetEventTxsFromEvents(
		testtypes.TestCdc, testtypes.TestGovernanceAddress, testtypes.TestMsgIndex.Events,
	)
	eventsSize := testtypes.MustGetSizeFromEvents(
		testtypes.TestCdc, testtypes.TestGovernanceAddress, testtypes.TestMsgIndex.Events,
	)

	totalSize := int(utils.TxSize(txRawBytes)) + eventsSize

	numEvents, err := tx.NumberOfEventsWithMaxBytes(events, uint64(totalSize))
	require.NoError(t, err)
	require.EqualValues(t, 3, numEvents) // just enough bytes

	numEvents, err = tx.NumberOfEventsWithMaxBytes(events, uint64(totalSize-1))
	require.NoError(t, err)
	require.EqualValues(t, 2, numEvents) // just under enough
}

func TestMsgIndex_NumberOfEventsWithMaxBytes(t *testing.T) {

	msgIndex := testtypes.TestMsgIndex.MsgIndex
	msgIndexRawBytes, err := msgIndex.RawTxBytes()
	require.NoError(t, err)

	events := testtypes.TestMsgIndex.Events
	eventTxs := testtypes.MustGetEventTxsFromEvents(testtypes.TestCdc, testtypes.TestGovernanceAddress, events)
	eventsSize := testtypes.MustGetSizeFromEvents(testtypes.TestCdc, testtypes.TestGovernanceAddress, events)

	txAndEventsSize := int(utils.TxSize(msgIndexRawBytes)) + eventsSize

	testCases := []struct {
		name           string
		eventTx        *testtypes.TestMsgIndexWithEvents
		maxBytes       uint64
		expNumOfEvents int
		expErrMsg      string
	}{
		{
			name:           "large max bytes fits all events",
			eventTx:        &testtypes.TestMsgIndex,
			maxBytes:       math.MaxInt64,
			expNumOfEvents: len(testtypes.TestMsgIndex.Events),
		},
		{
			name:           "exact size fits all events",
			eventTx:        &testtypes.TestMsgIndex,
			maxBytes:       uint64(txAndEventsSize),
			expNumOfEvents: len(testtypes.TestMsgIndex.Events),
		},
		{
			name:           "just under exact size fits n-1 events",
			eventTx:        &testtypes.TestMsgIndex,
			maxBytes:       uint64(txAndEventsSize) - 1,
			expNumOfEvents: len(testtypes.TestMsgIndex.Events) - 1,
		},
		{
			name:      "size smaller than MsgIndex size errors",
			eventTx:   &testtypes.TestMsgIndex,
			maxBytes:  uint64(utils.TxSize(msgIndexRawBytes)) - 1,
			expErrMsg: "could not fit MsgIndex",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			numOfEvents, err := tc.eventTx.NumberOfEventsWithMaxBytes(eventTxs, tc.maxBytes)

			if tc.expErrMsg != "" {
				require.ErrorContains(t, err, tc.expErrMsg)
				return
			}
			require.NoError(t, err)
			require.EqualValues(t, tc.expNumOfEvents, numOfEvents)
		})
	}
}

func TestMsgIndex_TrimEventsFromHead(t *testing.T) {

	events := testtypes.MustGetEventTxsFromEvents(
		testtypes.TestCdc, testtypes.TestGovernanceAddress, testtypes.TestMsgIndex.Events,
	)

	testCases := []struct {
		name            string
		eventTx         types.MsgIndex // do not use a pointer, since the function modifies in-place
		events          [][]byte
		numEventsToTrim uint64
		expEventTx      types.MsgIndex
		expEvents       [][]byte
		expErrMsg       string
	}{
		{
			name: "trim none => same MsgIndex",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          events,
			numEventsToTrim: 0,
			expEventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			expEvents: events,
		},
		{
			name: "trim one => trimmed MsgIndex",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          events,
			numEventsToTrim: 1,
			expEventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs - 1, // 1 trimmed
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			expEvents: events[1:], // 1 trimmed
		},
		{
			name: "trim all is not possible",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			numEventsToTrim: uint64(len(testtypes.TestMsgIndex.Events)),
			expErrMsg:       "cannot trim all 3 events",
		},
		{
			name: "trim more than all is not possible",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			numEventsToTrim: uint64(len(testtypes.TestMsgIndex.Events)) + 1,
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

func TestMsgIndex_KeepEventsFromHead(t *testing.T) {

	eventTxs := testtypes.MustGetEventTxsFromEvents(
		testtypes.TestCdc, testtypes.TestGovernanceAddress, testtypes.TestMsgIndex.Events,
	)

	testCases := []struct {
		name            string
		eventTx         types.MsgIndex // do not use a pointer, since the function modifies in-place
		events          [][]byte
		numEventsToKeep uint64
		expTrimmed      uint64
		expEventTx      types.MsgIndex
		expEvents       [][]byte
		expErrMsg       string
	}{
		{
			name: "keep all => same MsgIndex",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          eventTxs,
			numEventsToKeep: uint64(len(testtypes.TestMsgIndex.Events)),
			expEventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			expEvents: eventTxs,
		},
		{
			name: "keep all but one => same MsgIndex",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: 3,
				NewEthereumBlock:    true, // will become false
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          eventTxs[:3],
			numEventsToKeep: uint64(len(testtypes.TestMsgIndex.Events)) - 1,
			expTrimmed:      1,
			expEventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: 2,     // 1 trimmed
				NewEthereumBlock:    false, // becomes false
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			expEvents: eventTxs[:2], // 1 trimmed
		},
		{
			name: "keep none is not possible",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          eventTxs,
			numEventsToKeep: 0,
			expErrMsg:       "cannot trim all 3 events",
		},
		{
			name: "keep more than all is not possible",
			eventTx: types.MsgIndex{
				Authority:           testtypes.TestMsgIndex.Authority,
				NumInjectedEventTxs: testtypes.TestMsgIndex.NumInjectedEventTxs,
				NewEthereumBlock:    testtypes.TestMsgIndex.NewEthereumBlock,
				BlockNumber:         testtypes.TestMsgIndex.BlockNumber,
			},
			events:          eventTxs,
			numEventsToKeep: uint64(len(testtypes.TestMsgIndex.Events)) + 1,
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
