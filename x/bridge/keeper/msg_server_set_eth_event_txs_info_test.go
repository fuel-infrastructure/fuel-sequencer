package keeper_test

import (
	"time"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestSetEthEventTxsIndex_SingleTransaction() {
	encodedEthEventsTxWithEvents := *testtypes.TestEthEventsTx.EthEventsTx
	encodedEthEventsTxPartialBlock := *testtypes.TestEthEventsTxPartial.EthEventsTx
	encodedEthEventsTxWithoutEvents := *testtypes.TestEthEventsTxWithoutEvents.EthEventsTx

	ethEventsTxWithWrongBlock := *testtypes.TestEthEventsTx.EthEventsTx
	ethEventsTxWithWrongBlock.BlockNumber = 99
	encodedEthEventsTxWithWrongBlock := ethEventsTxWithWrongBlock

	testBlockTime := time.Now().Round(0)

	testCases := []struct {
		name                            string
		setEthEventsTxIndex             *types.EthEventsTxIndex
		ethEventsTx                     types.EthEventsTx
		expectNewBlock                  bool
		expectEthereumEventsIndexOffset uint64
		expectErrMsg                    string
	}{
		{
			name:         "EthEventsTx with wrong block number => error",
			ethEventsTx:  encodedEthEventsTxWithWrongBlock,
			expectErrMsg: "expected block number 1, got 99 in EthEventsTx",
		},
		{
			name:                            "EthEventsTx with events => new block and offset stays at zero",
			ethEventsTx:                     encodedEthEventsTxWithEvents,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
		},
		{
			name:                            "EthEventsTx without events => new block and offset stays at zero",
			ethEventsTx:                     encodedEthEventsTxWithoutEvents,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
		},
		{
			name:                            "EthEventsTx with partial events => no new block but offset updated",
			ethEventsTx:                     encodedEthEventsTxPartialBlock,
			expectNewBlock:                  false,
			expectEthereumEventsIndexOffset: testtypes.TestEthEventsTxPartial.NumInjectedEvents,
		},
		{
			name:                            "EthEventsTx with partial events => no new block but offset updated",
			ethEventsTx:                     encodedEthEventsTxPartialBlock,
			expectNewBlock:                  false,
			expectEthereumEventsIndexOffset: testtypes.TestEthEventsTxPartial.NumInjectedEvents,
		},
		{
			name: "failure if EthEventsTxIndex already exists",
			setEthEventsTxIndex: &types.EthEventsTxIndex{
				NumUnhandledEventTxs: 2,
			},
			ethEventsTx:  encodedEthEventsTxWithEvents,
			expectErrMsg: "MsgSetEthEventTxsIndex was already processed",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			// Set EthEventsTxIndex
			if tc.setEthEventsTxIndex != nil {
				s.App.BridgeKeeper.SetEthEventsTxIndex(s.Ctx(), *tc.setEthEventsTxIndex)
			}

			_, err := msgServer.SetEthEventTxsIndex(s.Ctx().WithBlockTime(testBlockTime), &tc.ethEventsTx)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			// Verify that EthEventsTxIndex is in state
			ethEventsTxIndex, found := s.App.BridgeKeeper.GetEthEventsTxIndex(s.Ctx())
			s.Require().True(found)
			s.Require().Equal(ethEventsTxIndex.NumUnhandledEventTxs, tc.ethEventsTx.NumInjectedEvents)

			// Check LastEthereumBlockSynced
			lastBlock, found := s.App.BridgeKeeper.GetLastEthereumBlockSynced(s.Ctx())
			s.Require().True(found)
			if tc.expectNewBlock {
				s.Require().EqualValues(1, lastBlock)
			} else {
				s.Require().EqualValues(0, lastBlock)
			}

			// Check EthereumEventIndexOffset
			indexOffset, found := s.App.BridgeKeeper.GetEthereumEventIndexOffset(s.Ctx())
			s.Require().True(found)
			s.Require().EqualValues(indexOffset, tc.expectEthereumEventsIndexOffset)

			// Check LastEthBlockUpdateTime
			lastEthBlockUpdateTime, found := s.App.BridgeKeeper.GetLastEthBlockUpdateTime(s.Ctx())
			if tc.expectNewBlock {
				s.Require().True(found)
				s.Require().Equal(testBlockTime, lastEthBlockUpdateTime)
			} else {
				s.Require().False(found)
				s.Require().Equal(time.Time{}, lastEthBlockUpdateTime)
			}
		})
	}
}

func (s *KeeperTestSuite) TestSetEthEventTxsIndex_Combinations() {

	ethEventsTxWithEvents1 := *testtypes.TestEthEventsTx.EthEventsTx
	ethEventsTxPartialBlock1 := *testtypes.TestEthEventsTxPartial.EthEventsTx
	ethEventsTxWithoutEvents1 := *testtypes.TestEthEventsTxWithoutEvents.EthEventsTx
	ethEventsTxNoNewBlock1 := *testtypes.TestEthEventsTxNoNewBlock.EthEventsTx

	// Set of events with block number set to 2
	ethEventsTxWithEvents2 := *testtypes.TestEthEventsTx.EthEventsTx
	ethEventsTxPartialBlock2 := *testtypes.TestEthEventsTxPartial.EthEventsTx
	ethEventsTxWithoutEvents2 := *testtypes.TestEthEventsTxWithoutEvents.EthEventsTx
	ethEventsTxNoNewBlock2 := *testtypes.TestEthEventsTxNoNewBlock.EthEventsTx
	ethEventsTxWithEvents2.BlockNumber = 2
	ethEventsTxPartialBlock2.BlockNumber = 2
	ethEventsTxWithoutEvents2.BlockNumber = 2
	ethEventsTxNoNewBlock2.BlockNumber = 2

	// Ensure that original block numbers are as expected, and unchanged
	s.Require().EqualValues(ethEventsTxWithEvents1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxPartialBlock1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxWithoutEvents1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxNoNewBlock1.BlockNumber, 1)

	numberOfEventsInPartialTx := testtypes.TestEthEventsTxPartial.NumInjectedEvents

	testBlockTime1 := time.Now().Round(0)
	testBlockTime2 := time.Now().Round(0)

	testCases := []struct {
		name                            string
		ethEventsTx                     []types.EthEventsTx
		blockTime                       []time.Time
		expectLastEthereumBlockSynced   []uint64
		expectEthereumEventsIndexOffset []uint64
		expectErrMsg                    []string
		expectLastEthBlockUpdateTime    []bool
		expectBlockTime                 []time.Time
	}{
		// ---------------------------- Combinations of Partial and Full
		{
			name: "Partial + Full => 0,1 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + Partial => 1,1 synced and 0,N offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxPartialBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Partial and NoNewBlock
		{
			name: "Partial + NoNewBlock => ERR because we expect at least one new event",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxNoNewBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
			expectErrMsg: []string{
				"",
				"expected at least 1 new event if offset is non-zero (2)",
			},
		},
		{
			name: "NoNewBlock + Partial => 0,0 synced and 0,N offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxPartialBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of Partial and NoEvents
		{
			name: "Partial + NoEvents => ERR because we expect at least one new event",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxWithoutEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
			expectErrMsg: []string{
				"",
				"expected at least 1 new event if offset is non-zero (2)",
			},
		},
		{
			name: "NoEvents + Partial => 1,1 synced and 0,N offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxPartialBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoNewBlock and NoEvents
		{
			name: "NoNewBlock + NoEvents => 0,1 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxWithoutEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "NoEvents + NoNewBlock => 1,1 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxNoNewBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoNewBlock and Full
		{
			name: "NoNewBlock + Full => 0,1 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + NoNewBlock => 1,1 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxNoNewBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoEvents and Full
		{
			name: "NoEvents + Full => 1,2 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxWithEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		{
			name: "Full + NoEvents => 1,2 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxWithoutEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Full and Full
		{
			name: "Full + Full => 1,2 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxWithEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Partial and Partial
		{
			name: "Partial + Partial => 0,0 synced and N,2N offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxPartialBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx, numberOfEventsInPartialTx * 2},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of NoNewBlock and NoNewBlock
		{
			name: "NoNewBlock + NoNewBlock => 0,0 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxNoNewBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of NoEvents and NoEvents
		{
			name: "NoEvents + NoEvents => 1,2 synced and 0,0 offset",
			ethEventsTx: []types.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxWithoutEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			for i := 0; i < len(tc.ethEventsTx); i++ {

				_, err := msgServer.SetEthEventTxsIndex(s.Ctx().WithBlockTime(tc.blockTime[i]), &tc.ethEventsTx[i])
				if tc.expectErrMsg != nil && tc.expectErrMsg[i] != "" {
					s.Require().ErrorContains(err, tc.expectErrMsg[i])
					return
				}
				s.Require().NoError(err)

				expectLastEthereumBlockSynced := tc.expectLastEthereumBlockSynced[i]
				expectEthereumEventsIndexOffset := tc.expectEthereumEventsIndexOffset[i]
				expectLastEthBlockUpdateTime := tc.expectLastEthBlockUpdateTime[i]
				expectBlockTime := tc.expectBlockTime[i]

				// Verify that EthEventsTxIndex is in state
				ethEventsTxIndex, found := s.App.BridgeKeeper.GetEthEventsTxIndex(s.Ctx())
				s.Require().True(found)
				s.Require().Equal(ethEventsTxIndex.NumUnhandledEventTxs, tc.ethEventsTx[i].NumInjectedEvents)

				// Check LastEthereumBlockSynced
				lastBlock, found := s.App.BridgeKeeper.GetLastEthereumBlockSynced(s.Ctx())
				s.Require().True(found)
				s.Require().EqualValues(expectLastEthereumBlockSynced, lastBlock)

				// Check EthereumEventIndexOffset
				indexOffset, found := s.App.BridgeKeeper.GetEthereumEventIndexOffset(s.Ctx())
				s.Require().True(found)
				s.Require().EqualValues(indexOffset, expectEthereumEventsIndexOffset)

				// Check LastEthBlockUpdateTime
				lastEthBlockUpdateTime, found := s.App.BridgeKeeper.GetLastEthBlockUpdateTime(s.Ctx())
				s.Require().Equal(expectBlockTime, lastEthBlockUpdateTime)
				if expectLastEthBlockUpdateTime {
					s.Require().True(found)
				} else {
					s.Require().False(found)
				}

				// Simulate EndBlocker consuming the events between each PreBlocker call
				s.App.BridgeKeeper.RemoveEthEventsTxIndex(s.Ctx())
			}
		})
	}
}
