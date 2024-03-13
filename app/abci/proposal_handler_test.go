package abci_test

import (
	"errors"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	"github.com/golang/mock/gomock"
)

func (s *AppTestSuite) TestMsgSupplyDelta() {
	//s.App.
	//testCases := []struct {
	//	name              string
	//	supplyDeltaPeriod uint64
	//	lastEthereumNonce sdk.Int
	//	supplyDeltaInfo   bridgetypes.SupplyDeltaInfo
	//	chainHeight       int64
	//	msg               *bridgetypes.MsgSupplyDelta
	//	expErrMsg         string
	//}{
	//	{
	//		name:              "valid MsgSupplyDelta - block height equals period",
	//		supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
	//		lastEthereumNonce: testtypes.TestLastEthereumNonce,
	//		supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
	//		chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
	//		msg: &bridgetypes.MsgSupplyDelta{
	//			Authority: testtypes.TestGovernanceAddress,
	//		},
	//	},
	//	{
	//		name:              "valid MsgSupplyDelta - block height not equal period",
	//		supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
	//		lastEthereumNonce: testtypes.TestLastEthereumNonce,
	//		supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
	//		chainHeight:       int64(testtypes.TestSupplyDeltaPeriod * 4),
	//		msg: &bridgetypes.MsgSupplyDelta{
	//			Authority: testtypes.TestGovernanceAddress,
	//		},
	//	},
	//	{
	//		name:              "invalid MsgSupplyDelta - invalid authority",
	//		supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
	//		lastEthereumNonce: testtypes.TestLastEthereumNonce,
	//		supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
	//		chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
	//		msg: &bridgetypes.MsgSupplyDelta{
	//			Authority: "invalid address",
	//		},
	//		expErrMsg: errors.Wrapf(
	//			bridgetypes.ErrInvalidSigner,
	//			"invalid authority; expected %s, got %s",
	//			testtypes.TestGovernanceAddress,
	//			"invalid address",
	//		).Error(),
	//	},
	//	{
	//		name:              "invalid MsgSupplyDelta - SupplyDeltaPeriod is zero",
	//		supplyDeltaPeriod: 0,
	//		lastEthereumNonce: testtypes.TestLastEthereumNonce,
	//		supplyDeltaInfo:   testtypes.TestSupplyDeltaInfo,
	//		chainHeight:       int64(testtypes.TestSupplyDeltaPeriod),
	//		msg: &bridgetypes.MsgSupplyDelta{
	//			Authority: testtypes.TestGovernanceAddress,
	//		},
	//		expErrMsg: errors.Wrapf(
	//			bridgetypes.ErrInvalidSupplyDeltaPeriod, "SupplyDeltaPeriod cannot be zero",
	//		).Error(),
	//	},
	//	{
	//		name:              "invalid MsgSupplyDelta - SupplyDelta is zero",
	//		supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
	//		lastEthereumNonce: testtypes.TestLastEthereumNonce,
	//		supplyDeltaInfo: bridgetypes.SupplyDeltaInfo{
	//			LastSupply: testtypes.TestLastSupply,
	//			Delta:      sdk.ZeroInt(),
	//			Offset:     sdk.ZeroInt(),
	//		},
	//		chainHeight: int64(testtypes.TestSupplyDeltaPeriod),
	//		msg: &bridgetypes.MsgSupplyDelta{
	//			Authority: testtypes.TestGovernanceAddress,
	//		},
	//		expErrMsg: errors.Wrapf(
	//			bridgetypes.ErrInvalidSupplyDeltaValue, "cannot report 0 supply delta to Ethereum",
	//		).Error(),
	//	},
	//}

	s.SetupTest()
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)
	sidecarClientMock.EXPECT().GetBlockEvents(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("test error"))
	propHandler := s.GetTestProposalHandler(sidecarClientMock)
	propHandler.PrepareProposalHandler()(s.Ctx(), &abcitypes.RequestPrepareProposal{
		MaxTxBytes:         0,
		Txs:                nil,
		LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
		Misbehavior:        nil,
		Height:             0,
		Time:               time.Time{},
		NextValidatorsHash: nil,
		ProposerAddress:    nil,
	})
}
