package keeper_test

import (
	"reflect"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	"github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgSkippedEventTx_SingleTransaction() {
	msgSkippedEventTx := *testtypes.TestMsgSkippedEventTx
	msgSkippedEventTx2 := *testtypes.TestMsgSkippedEventTx2
	eventSkippedEventTx := types.EventSkippedEventTx{
		ReasonForSkip:  testtypes.TestMsgSkippedEventTx.ReasonForSkip,
		EthBlockNumber: testtypes.TestMsgSkippedEventTx.EthBlockNumber,
		EthLogIndex:    testtypes.TestMsgSkippedEventTx.EthLogIndex,
		EthTxIndex:     testtypes.TestMsgSkippedEventTx.EthTxIndex,
		EthTxHash:      testtypes.TestMsgSkippedEventTx.EthTxHash,
	}
	eventSkippedEventTx2 := types.EventSkippedEventTx{
		ReasonForSkip:  testtypes.TestMsgSkippedEventTx2.ReasonForSkip,
		EthBlockNumber: testtypes.TestMsgSkippedEventTx2.EthBlockNumber,
		EthLogIndex:    testtypes.TestMsgSkippedEventTx2.EthLogIndex,
		EthTxIndex:     testtypes.TestMsgSkippedEventTx2.EthTxIndex,
		EthTxHash:      testtypes.TestMsgSkippedEventTx2.EthTxHash,
	}

	testBlockTime := time.Now().Round(0)
	heightToAvoidSupplyDelta := int64(999)
	//nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
	heightForSupplyDelta := int64(testtypes.TestSupplyDeltaPeriod)

	testCases := []struct {
		name                   string
		supplyDeltaPeriod      uint64
		msgs                   []types.MsgSkippedEventTx
		blockHeight            int64
		expectMsgSkippedEvents []*types.EventSkippedEventTx
	}{
		{
			name:                   "No skipped events",
			supplyDeltaPeriod:      testtypes.TestSupplyDeltaPeriod,
			msgs:                   []types.MsgSkippedEventTx{msgSkippedEventTx},
			blockHeight:            heightToAvoidSupplyDelta,
			expectMsgSkippedEvents: []*types.EventSkippedEventTx{},
		},
		{
			name:              "Single skipped event",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			msgs:              []types.MsgSkippedEventTx{msgSkippedEventTx},
			blockHeight:       heightToAvoidSupplyDelta,
			expectMsgSkippedEvents: []*types.EventSkippedEventTx{
				&eventSkippedEventTx,
			},
		},
		{
			name:              "Single skipped event with different block height, with supply delta period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			msgs:              []types.MsgSkippedEventTx{msgSkippedEventTx},
			blockHeight:       heightForSupplyDelta,
			expectMsgSkippedEvents: []*types.EventSkippedEventTx{
				&eventSkippedEventTx,
			},
		},
		{
			name:              "Multiple skipped events",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			msgs:              []types.MsgSkippedEventTx{msgSkippedEventTx, msgSkippedEventTx2},
			blockHeight:       heightToAvoidSupplyDelta,
			expectMsgSkippedEvents: []*types.EventSkippedEventTx{
				&eventSkippedEventTx,
				&eventSkippedEventTx2,
			},
		},
		{
			name:              "Multiple skipped events with supply delta period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			msgs:              []types.MsgSkippedEventTx{msgSkippedEventTx, msgSkippedEventTx2},
			blockHeight:       heightForSupplyDelta,
			expectMsgSkippedEvents: []*types.EventSkippedEventTx{
				&eventSkippedEventTx,
				&eventSkippedEventTx2,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set the supply delta period
			params := s.App.BridgeKeeper.GetParams(s.Ctx())
			params.SupplyDeltaPeriod = tc.supplyDeltaPeriod
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), params)
			s.Require().NoError(err)

			// Get the message server
			msgServer := keeper.NewMsgServerImpl(s.App.BridgeKeeper)

			msgSkippedEventTxCtx := s.Ctx().WithBlockTime(testBlockTime).WithBlockHeight(tc.blockHeight)
			for _, msg := range tc.msgs {
				_, err = msgServer.SkippedEventTx(msgSkippedEventTxCtx, &msg)
				s.Require().NoError(err)
			}

			// Check MsgSkippedEventTx
			if tc.expectMsgSkippedEvents == nil {
				s.AssertEventEmitted(msgSkippedEventTxCtx, sdk.EventTypeMessage, 0) // not emitted
			} else {
				var latestExpectedEvent sdk.Event
				receivedEvents := msgSkippedEventTxCtx.EventManager().Events()

				for _, expectMsgSkippedEvent := range tc.expectMsgSkippedEvents {
					expectEvent, err := sdk.TypedEventToEvent(expectMsgSkippedEvent)
					s.Require().NoError(err)

					latestExpectedEvent = expectEvent

					found := false
					for _, event := range receivedEvents {
						if found = reflect.DeepEqual(expectEvent, event); found {
							break
						}
					}
					if !found {
						s.FailNow("expected event not found",
							"event received", latestExpectedEvent,
							"expected events", tc.expectMsgSkippedEvents)
					}
				}
			}
		})
	}
}
