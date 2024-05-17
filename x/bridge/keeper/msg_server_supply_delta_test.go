package keeper_test

import (
	"cosmossdk.io/errors"
	sdk "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *KeeperTestSuite) TestMsgSupplyDelta() {
	testCases := []struct {
		name              string
		supplyDeltaPeriod uint64
		lastEthereumNonce sdk.Int
		supplyDeltaInfo   bridgetypes.SupplyDeltaInfo
		chainHeight       int64
		msg               *bridgetypes.MsgSupplyDelta
		expErrMsg         string
	}{
		{
			name:              "valid MsgSupplyDelta - block height equals period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod), // height % period == 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
		},
		{
			name:              "valid MsgSupplyDelta - block height not equal period",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod * 4), // height % period == 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
		},
		{
			name:              "invalid MsgSupplyDelta - block height not at the right height",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod + 1), // height % period != 0
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expErrMsg: "MsgSupplyDelta not expected at height 101",
		},
		{
			name:              "invalid MsgSupplyDelta - invalid authority",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
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
			supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
			chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expErrMsg: errors.Wrapf(
				bridgetypes.ErrInvalidSupplyDeltaPeriod, "SupplyDeltaPeriod cannot be zero",
			).Error(),
		},
		{
			name:              "invalid MsgSupplyDelta - SupplyDelta is zero",
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			lastEthereumNonce: testtypes.TestLastEthereumNonce,
			supplyDeltaInfo: bridgetypes.SupplyDeltaInfo{
				LastSupply: testtypes.TestLastSupply,
				Delta:      sdk.ZeroInt(),
				Offset:     sdk.ZeroInt(),
			},
			chainHeight: int64(testtypes.TestSupplyDeltaPeriod),
			msg: &bridgetypes.MsgSupplyDelta{
				Authority: testtypes.TestGovernanceAddress,
			},
			expErrMsg: errors.Wrapf(
				bridgetypes.ErrInvalidSupplyDeltaValue, "cannot report 0 supply delta to Ethereum",
			).Error(),
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

			// Confirm that the LastEthereumNonce has been incremented by 1
			actualNonce, found := s.App.BridgeKeeper.GetLastEthereumNonce(s.Ctx())
			expectedNonce := tc.lastEthereumNonce.Add(sdk.OneInt())
			s.Require().True(found)
			s.Require().Equal(expectedNonce, actualNonce)

			// Confirm that SupplyDeltaInfo has been reset to the correct value
			actualSupplyDeltaInfo := s.App.BridgeKeeper.MustGetSupplyDeltaInfo(s.Ctx())
			expectedSupplyDeltaInfo := bridgetypes.SupplyDeltaInfo{
				LastSupply: tc.supplyDeltaInfo.LastSupply,
				Delta:      sdk.ZeroInt(),
				Offset:     sdk.ZeroInt(),
			}
			s.Require().Equal(expectedSupplyDeltaInfo, actualSupplyDeltaInfo)

			// Confirm that we received the expected response
			expectedReportedDelta := tc.supplyDeltaInfo.Delta.Add(tc.supplyDeltaInfo.Offset)
			expectedResponse := &bridgetypes.MsgSupplyDeltaResponse{
				Nonce:       expectedNonce,
				SupplyDelta: expectedReportedDelta,
			}
			s.Require().Equal(expectedResponse, response)

			// Check events emitted
			s.AssertEventEmitted(msgSupplyDeltaCtx, proto.MessageName(&bridgetypes.EventSupplyDeltaReported{}), 1)
		})
	}
}
