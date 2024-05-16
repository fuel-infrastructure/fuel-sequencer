package keeper_test

import (
	"time"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgIndex_SingleTransaction() {
	encodedMsgIndexWithEvents := *testtypes.TestMsgIndex.MsgIndex
	encodedMsgIndexPartialBlock := *testtypes.TestMsgIndexPartial.MsgIndex
	encodedMsgIndexWithoutEvents := *testtypes.TestMsgIndexWithoutEvents.MsgIndex

	msgIndexWithWrongBlock := *testtypes.TestMsgIndex.MsgIndex
	msgIndexWithWrongBlock.BlockNumber = 99
	encodedMsgIndexWithWrongBlock := msgIndexWithWrongBlock

	testBlockTime := time.Now().Round(0)
	heightToAvoidSupplyDelta := int64(999)
	heightForSupplyDelta := int64(types.DefaultSupplyDeltaPeriod)

	testCases := []struct {
		name                            string
		setIndex                        *types.Index
		msg                             types.MsgIndex
		blockHeight                     int64
		expectNewBlock                  bool
		expectEthereumEventsIndexOffset uint64
		expectNumInjectedTxsTotal       uint64
		expectErrMsg                    string
	}{
		{
			name:         "MsgIndex with wrong block number => error",
			msg:          encodedMsgIndexWithWrongBlock,
			blockHeight:  heightToAvoidSupplyDelta,
			expectErrMsg: "expected block number 1, got 99 in MsgIndex",
		},
		{
			name:                            "MsgIndex with events => new block and offset stays at zero",
			msg:                             encodedMsgIndexWithEvents,
			blockHeight:                     heightToAvoidSupplyDelta,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
			expectNumInjectedTxsTotal:       encodedMsgIndexWithEvents.NumInjectedTxs,
		},
		{
			name:                            "MsgIndex without events => new block and offset stays at zero",
			msg:                             encodedMsgIndexWithoutEvents,
			blockHeight:                     heightToAvoidSupplyDelta,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
			expectNumInjectedTxsTotal:       encodedMsgIndexWithoutEvents.NumInjectedTxs,
		},
		{
			name:                            "MsgIndex with partial events => no new block but offset updated",
			msg:                             encodedMsgIndexPartialBlock,
			blockHeight:                     heightToAvoidSupplyDelta,
			expectNewBlock:                  false,
			expectEthereumEventsIndexOffset: encodedMsgIndexPartialBlock.NumInjectedTxs,
			expectNumInjectedTxsTotal:       encodedMsgIndexPartialBlock.NumInjectedTxs,
		},
		{
			name:                            "MsgIndex with partial events => no new block but offset updated",
			msg:                             encodedMsgIndexPartialBlock,
			blockHeight:                     heightToAvoidSupplyDelta,
			expectNewBlock:                  false,
			expectEthereumEventsIndexOffset: encodedMsgIndexPartialBlock.NumInjectedTxs,
			expectNumInjectedTxsTotal:       encodedMsgIndexPartialBlock.NumInjectedTxs,
		},
		{
			name:                            "MsgIndex with events at supply delta height => supply delta considered",
			msg:                             encodedMsgIndexWithEvents,
			blockHeight:                     heightForSupplyDelta,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
			expectNumInjectedTxsTotal:       encodedMsgIndexWithEvents.NumInjectedTxs + 1, // +1 for MsgSupplyDelta
		},
		{
			name:         "MsgIndex with invalid authority => failed",
			msg:          types.MsgIndex{Authority: "fuelsequencer17w0adeg64ky0daxwd2ugyuneellmjgnx5dpmtz"},
			expectErrMsg: "invalid authority",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			// Set Index
			if tc.setIndex != nil {
				s.App.BridgeKeeper.SetIndex(s.Ctx(), *tc.setIndex)
			}

			_, err := msgServer.Index(s.Ctx().WithBlockTime(testBlockTime).WithBlockHeight(tc.blockHeight), &tc.msg)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			// Verify that Index is in state
			index, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
			s.Require().True(found)
			s.Require().EqualValues(index.NumInjectedTxsTotal, tc.expectNumInjectedTxsTotal)
			s.Require().EqualValues(index.NumInjectedTxsAnte, 0)

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

func (s *KeeperTestSuite) TestMsgIndex_Combinations() {

	msgWithEvents1 := *testtypes.TestMsgIndex.MsgIndex
	msgPartialBlock1 := *testtypes.TestMsgIndexPartial.MsgIndex
	msgWithoutEvents1 := *testtypes.TestMsgIndexWithoutEvents.MsgIndex
	msgNoNewBlock1 := *testtypes.TestMsgIndexNoNewBlock.MsgIndex

	// Set of events with block number set to 2
	msgWithEvents2 := *testtypes.TestMsgIndex.MsgIndex
	msgPartialBlock2 := *testtypes.TestMsgIndexPartial.MsgIndex
	msgWithoutEvents2 := *testtypes.TestMsgIndexWithoutEvents.MsgIndex
	msgNoNewBlock2 := *testtypes.TestMsgIndexNoNewBlock.MsgIndex
	msgWithEvents2.BlockNumber = 2
	msgPartialBlock2.BlockNumber = 2
	msgWithoutEvents2.BlockNumber = 2
	msgNoNewBlock2.BlockNumber = 2

	// Ensure that original block numbers are as expected, and unchanged
	s.Require().EqualValues(msgWithEvents1.BlockNumber, 1)
	s.Require().EqualValues(msgPartialBlock1.BlockNumber, 1)
	s.Require().EqualValues(msgWithoutEvents1.BlockNumber, 1)
	s.Require().EqualValues(msgNoNewBlock1.BlockNumber, 1)

	numberOfEventsInPartialTx := testtypes.TestMsgIndexPartial.NumInjectedTxs

	testBlockTime1 := time.Now().Round(0)
	testBlockTime2 := time.Now().Round(0)

	testCases := []struct {
		name                            string
		msg                             []types.MsgIndex
		blockHeight                     uint64
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
			msg: []types.MsgIndex{
				msgPartialBlock1,
				msgWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + Partial => 1,1 synced and 0,N offset",
			msg: []types.MsgIndex{
				msgWithEvents1,
				msgPartialBlock2, // from new block
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
			msg: []types.MsgIndex{
				msgPartialBlock1,
				msgNoNewBlock1, // from same block
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
			msg: []types.MsgIndex{
				msgNoNewBlock1,
				msgPartialBlock1, // from same block
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
			msg: []types.MsgIndex{
				msgPartialBlock1,
				msgWithoutEvents1, // from same block
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
			msg: []types.MsgIndex{
				msgWithoutEvents1,
				msgPartialBlock2, // from new block
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
			msg: []types.MsgIndex{
				msgNoNewBlock1,
				msgWithoutEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "NoEvents + NoNewBlock => 1,1 synced and 0,0 offset",
			msg: []types.MsgIndex{
				msgWithoutEvents1,
				msgNoNewBlock2, // from new block
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
			msg: []types.MsgIndex{
				msgNoNewBlock1,
				msgWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + NoNewBlock => 1,1 synced and 0,0 offset",
			msg: []types.MsgIndex{
				msgWithEvents1,
				msgNoNewBlock2, // from new block
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
			msg: []types.MsgIndex{
				msgWithoutEvents1,
				msgWithEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		{
			name: "Full + NoEvents => 1,2 synced and 0,0 offset",
			msg: []types.MsgIndex{
				msgWithEvents1,
				msgWithoutEvents2, // from new block
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
			msg: []types.MsgIndex{
				msgWithEvents1,
				msgWithEvents2, // from new block
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
			msg: []types.MsgIndex{
				msgPartialBlock1,
				msgPartialBlock1, // from same block
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
			msg: []types.MsgIndex{
				msgNoNewBlock1,
				msgNoNewBlock1, // from same block
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
			msg: []types.MsgIndex{
				msgWithoutEvents1,
				msgWithoutEvents2, // from new block
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

			for i := 0; i < len(tc.msg); i++ {

				heightToAvoidSupplyDelta := int64(999)
				indexCtx := s.Ctx().WithBlockTime(tc.blockTime[i]).WithBlockHeight(heightToAvoidSupplyDelta)
				_, err := msgServer.Index(indexCtx, &tc.msg[i])
				if tc.expectErrMsg != nil && tc.expectErrMsg[i] != "" {
					s.Require().ErrorContains(err, tc.expectErrMsg[i])
					return
				}
				s.Require().NoError(err)

				expectLastEthereumBlockSynced := tc.expectLastEthereumBlockSynced[i]
				expectEthereumEventsIndexOffset := tc.expectEthereumEventsIndexOffset[i]
				expectLastEthBlockUpdateTime := tc.expectLastEthBlockUpdateTime[i]
				expectBlockTime := tc.expectBlockTime[i]

				// Verify that Index is in state
				index, found := s.App.BridgeKeeper.GetIndex(s.Ctx())
				s.Require().True(found)
				s.Require().EqualValues(index.NumInjectedTxsTotal, tc.msg[i].NumInjectedTxs)
				s.Require().EqualValues(index.NumInjectedTxsAnte, 0)

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
				s.App.BridgeKeeper.RemoveIndex(s.Ctx())
			}
		})
	}
}
