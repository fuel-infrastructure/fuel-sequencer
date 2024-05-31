package basic_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *BasicTestSuite) TestMsgSupplyDeltaIsInjected() {
	s.Run("Check that MsgSupplyDelta is getting injected only at the right heights", func() {

		supplyDeltaPeriod := int64(s.QueryBridgeParams(s.Ctx()).SupplyDeltaPeriod)
		supplyDeltaEvent, err := sdk.TypedEventToEvent(&bridgetypes.EventSupplyDeltaReported{})
		s.Require().NoError(err)

		// Wait for at least two MsgSupplyDelta to get injected
		searchTillBlock := supplyDeltaPeriod * 2
		err = s.WaitForSequencerBlocks(s.Ctx(), int(searchTillBlock), time.Minute)
		s.Require().NoError(err)

		// Construct standard MsgSupplyDelta details
		typicalMsgSupplyDelta := bridgetypes.MsgSupplyDelta{Authority: s.GetGovernanceAddress()}
		msgSupplyDeltaTypeUrl := sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{})

		// Ensure that MsgSupplyDelta injected, and only at the right heights
		for block := int64(1); block < searchTillBlock; block++ {
			expectMsgSupplyDelta := block%supplyDeltaPeriod == 0

			// Search for message showing successful injection
			msg, msgFound := s.SearchForMsgInBlock(s.Ctx(), msgSupplyDeltaTypeUrl, block)
			s.Require().Equal(expectMsgSupplyDelta, msgFound)
			if expectMsgSupplyDelta {
				msgSupplyDelta, ok := msg.(*bridgetypes.MsgSupplyDelta)
				s.Require().True(ok)
				s.Require().EqualValues(typicalMsgSupplyDelta, *msgSupplyDelta)
			}

			// Search for event showing successful execution
			_, eventFound := s.SearchForEventInBlockResults(s.Ctx(), supplyDeltaEvent.Type, block)
			s.Require().Equal(expectMsgSupplyDelta, eventFound)
		}
	})
}
