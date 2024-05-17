package keeper_test

import (
	"cosmossdk.io/errors"
	sdk "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgSupplyDelta() {

	testSupplyDeltaInfo := testtypes.TestSupplyDeltaInfo
	testSupplyDelta := testSupplyDeltaInfo.Delta.Add(testSupplyDeltaInfo.Offset)

	testSupplyDeltaInfoZeros := bridgetypes.SupplyDeltaInfo{
		LastSupply: testSupplyDeltaInfo.LastSupply,
		Delta:      sdk.ZeroInt(),
		Offset:     sdk.ZeroInt(),
	}

	testCases := []struct {
		name                 string
		supplyDeltaPeriod    uint64
		lastEthereumNonce    sdk.Int
		supplyDeltaInfo      bridgetypes.SupplyDeltaInfo
		chainHeight          int64
		msg                  *bridgetypes.MsgSupplyDelta
		expErrMsg            string
		expSupplyDeltaInfo   bridgetypes.SupplyDeltaInfo
		expLastEthereumNonce sdk.Int
		expResponse          *bridgetypes.MsgSupplyDeltaResponse
		expEventEmitted      bool
	}{
		{
			name:              "valid MsgSupplyDelta - block height equals period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod), // height % period == 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expSupplyDeltaInfo:   testSupplyDeltaInfoZeros,                  // reset to zero
			expLastEthereumNonce: testtypes.TestLastEthereumNonce.AddRaw(1), // incremented
			expResponse: &bridgetypes.MsgSupplyDeltaResponse{
				Nonce:       testtypes.TestLastEthereumNonce.AddRaw(1), // incremented
				SupplyDelta: testSupplyDelta,
			},
			expEventEmitted: true,
		},
		{
			name:              "valid MsgSupplyDelta - block height not equal period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod * 4), // height % period == 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expSupplyDeltaInfo:   testSupplyDeltaInfoZeros,                  // reset to zero
			expLastEthereumNonce: testtypes.TestLastEthereumNonce.AddRaw(1), // incremented
			expResponse: &bridgetypes.MsgSupplyDeltaResponse{
				Nonce:       testtypes.TestLastEthereumNonce.AddRaw(1), // incremented
				SupplyDelta: testSupplyDelta,
			},
			expEventEmitted: true,
		},
		{
			name:              "invalid MsgSupplyDelta - block height not at the right height",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod + 1), // height % period != 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expSupplyDeltaInfo:   testSupplyDeltaInfo,             // unchanged
			expLastEthereumNonce: testtypes.TestLastEthereumNonce, // unchanged
			expResponse:          &bridgetypes.MsgSupplyDeltaResponse{},
			expEventEmitted:      false,
		},
		{
			name:              "invalid MsgSupplyDelta - invalid authority",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: "invalid address",
			},
			expErrMsg: errors.Wrapf(
				bridgetypes.ErrInvalidSigner,
				"invalid authority; expected %s, got %s",
				testtypes.TestGovernanceAddress,
				"invalid address",
			).Error(),
		},
		{
			name:              "invalid MsgSupplyDelta - SupplyDeltaPeriod is zero",
			supplyDeltaPeriod: 0,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expSupplyDeltaInfo:   testSupplyDeltaInfo,             // unchanged
			expLastEthereumNonce: testtypes.TestLastEthereumNonce, // unchanged
			expResponse:          &bridgetypes.MsgSupplyDeltaResponse{},
			expEventEmitted:      false,
		},
		{
			name:              "invalid MsgSupplyDelta - SupplyDelta is zero",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testSupplyDeltaInfoZeros, // zeros
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expSupplyDeltaInfo:   testSupplyDeltaInfoZeros,        // unchanged
			expLastEthereumNonce: testtypes.TestLastEthereumNonce, // unchanged
			expResponse:          &bridgetypes.MsgSupplyDeltaResponse{},
			expEventEmitted:      false,
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

			// Fast-forward the chain so that MsgSupplyDelta is executed at the desired height
			msgSupplyDeltaCtx := s.Ctx().WithBlockHeight(tc.chainHeight)

			// Execute MsgSupplyDelta
			response, err := s.GetMsgServer().SupplyDelta(msgSupplyDeltaCtx, tc.msg)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				return
			}
			s.Require().NoError(err)

			// Confirm the value of LastEthereumNonce
			actualNonce, found := s.App.BridgeKeeper.GetLastEthereumNonce(s.Ctx())
			s.Require().True(found)
			s.Require().EqualValues(tc.expLastEthereumNonce, actualNonce)

			// Confirm the value of SupplyDeltaInfo
			actualSupplyDeltaInfo := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
			s.Require().EqualValues(tc.expSupplyDeltaInfo, actualSupplyDeltaInfo)

			// Confirm that we received the expected response
			s.Require().EqualValues(tc.expResponse, response)

			// Check events emitted
			event := proto.MessageName(&bridgetypes.EventSupplyDeltaReported{})
			if tc.expEventEmitted {
				s.AssertEventEmitted(msgSupplyDeltaCtx, event, 1)
			} else {
				s.AssertEventEmitted(msgSupplyDeltaCtx, event, 0)
			}
		})
	}
}
