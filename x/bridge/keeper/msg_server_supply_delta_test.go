package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgSupplyDelta() {
	testCases := []struct {
		name              string
		supplyDeltaPeriod uint64
		lastEthereumNonce sdkmath.Int
		supplyDeltaInfo   bridgetypes.SupplyDeltaInfo
		msg               *bridgetypes.MsgSupplyDelta
		expErrMsg         string
	}{
		{
			name:              "valid MsgSupplyDelta",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
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

			// Set the last Ethereum nonce
			s.App.BridgeKeeper.SetLastEthereumNonce(s.Ctx(), tc.lastEthereumNonce)

			// Set the supply delta info
			s.App.BridgeKeeper.SetSupplyDeltaInfo(s.Ctx(), tc.supplyDeltaInfo)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}
			s.Require().NoError(err)
		})
	}
}
