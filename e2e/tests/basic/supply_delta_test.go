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

		// Ensure that MsgSupplyDelta injected, and only at the right heights
		for block := int64(1); block < searchTillBlock; block++ {
			found := s.SearchForEventInBlockResults(s.Ctx(), supplyDeltaEvent.Type, block)
			if block%supplyDeltaPeriod == 0 {
				s.Require().True(found)
			} else {
				s.Require().False(found)
			}
		}
	})
}
