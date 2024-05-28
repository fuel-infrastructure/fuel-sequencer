package abci_test

import (
	"errors"
	"fmt"
	"math"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
	"github.com/golang/mock/gomock"
)

func calculateTotalTxBytes(txs [][]byte) uint64 {
	var totalTxBytes uint64

	for _, txBz := range txs {
		totalTxBytes += uint64(len(txBz))
	}

	return totalTxBytes
}

func (s *AppTestSuite) TestPrepareProposalHandler() {
	totalTxsGas := int64(3000) // Dummy Txs consume at most 1000 units of gas each. Injected Txs don't consume any gas
	encodedDummyTxs := s.CreateEncodedDummyTxs(3, 1000)

	encodedMsgIndexWithEvents := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndex)
	encodedMsgIndexPartialBlock := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexPartial)
	encodedMsgIndexWithoutEvents := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexWithoutEvents)
	encodedMsgIndexSidecarErr := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexSidecarErr)

	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx()

	totalTxsBytesWithEventsAndSupplyDelta := int64(calculateTotalTxBytes(
		append(
			encodedMsgIndexWithEvents,
			msgSupplyDeltaTx,
			encodedDummyTxs[0],
			encodedDummyTxs[1],
			encodedDummyTxs[2],
		),
	))

	four := uint64(4)

	testCases := []struct {
		name                           string
		removeLastEthereumBlockSynced  bool
		removeEthereumEventIndexOffset bool
		setEthereumEventIndexOffset    *uint64
		expQueryBlockEventsCalled      int
		expQueryBlockEventsReq         *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet            apptesting.MockQueryBlockEventsResponse
		requestPrepareProposal         *abcitypes.RequestPrepareProposal
		maxBlockGas                    int64
		supplyDeltaPeriod              uint64
		ethereumProxyContractAddress   string
		injectedEventTxMaxBytes        uint64
		expErrMsg                      string
		expRes                         *abcitypes.ResponsePrepareProposal
	}{
		{
			name:                      "returns supply delta tx if expected at height and other txs",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents,
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name:                      "returns injected txs and other txs if block events found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]),
			},
		},
		{
			name:                      "returns injected txs and other txs if no block events found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithoutEvents,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name:                      "returns injected txs and other txs if sidecar errors",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil, Error: errors.New("block not yet processed 1"),
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexSidecarErr,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name:                      "returns partial injected txs if not enough space in the block",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to the size of a partial eth tx, so that the last event does not fit. We need to add +2 since the
				// generated MsgIndex will have NewEthereumBlock set to true at first until the proposer updates it.
				// When NewEthereumBlock is set to true, it consumes 2 bytes, otherwise it does not consume anything.
				MaxTxBytes: int64(calculateTotalTxBytes(encodedMsgIndexPartialBlock)) + 2,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: encodedMsgIndexPartialBlock, // partial eth tx
			},
		},
		{
			name:                          "returns error if LastEthereumBlockSynced not found",
			removeLastEthereumBlockSynced: true,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "could not get last Ethereum block synced from state",
		},
		{
			name:                           "returns error if EthereumEventIndexOffset not found",
			removeEthereumEventIndexOffset: true,
			expQueryBlockEventsCalled:      0,
			expQueryBlockEventsReq:         nil,
			queryBlockEventsRet:            apptesting.MockQueryBlockEventsResponse{},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "could not get Ethereum event index offset from state",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - ValidateBasic error",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{Events: []*sidecartypes.Event{
					{EventType: "invalid-event", Data: nil},
				}},
				Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to generate MsgIndex",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - ValidateStateful error",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{
					Events: []*sidecartypes.Event{
						testutils.MustGetSidecarEventFromParsedEvent(
							testtypes.TestAuthorizeEvent1, "invalid-ethereum-proxy-contract-address",
						),
					},
				},
				Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to generate MsgIndex",
		},
		{
			name:                      "returns error if not enough block space for at least one MsgIndex event",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0, // Set to zero to make sure there is no capacity for the MsgIndex events
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to calculate number of events with max bytes 0",
		},
		{
			name:                      "returns error if none of the MsgIndex events fit in the block",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to the size of MsgSupplyDeltaTx so that MsgIndex does not fit
				MaxTxBytes: int64(calculateTotalTxBytes([][]byte{msgSupplyDeltaTx})),
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to calculate number of events with max bytes 0",
		},
		{
			name:                      "returns only supply delta and MsgIndex if it's just enough block size",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to EXACTLY the size of MsgSupplyDelta transaction plus MsgIndex
				MaxTxBytes: int64(calculateTotalTxBytes(append(encodedMsgIndexWithEvents, msgSupplyDeltaTx))),
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithEvents, msgSupplyDeltaTx),
			},
		},
		{
			name:                      "returns selected Txs if MaxTxBytes exceeded",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to expected size - 1 to omit last tx
				MaxTxBytes: totalTxsBytesWithEventsAndSupplyDelta - 1,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]),
			},
		},
		{
			name:                      "returns selected Txs if MaxBlockGas exceeded",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas - 1, // Set to total - 1 so that the last transaction is omitted
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]),
			},
		},
		{
			name:                        "returns error if Ethereum event index offset larger than events list",
			setEthereumEventIndexOffset: &four,
			expQueryBlockEventsCalled:   1,
			expQueryBlockEventsReq:      &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: "failed to trim event txs from head: insufficient no of events, expected at " +
				"least 4 got 3",
		},
		{
			name:                      "returns error if invalid deposit found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseInvalidDeposit, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to generate MsgIndex",
		},
		{
			name:                      "skips event if invalid authorize found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseInvalidAuthorize, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]),
			},
		},
		{
			name:                      "returns error if big deposit found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseDepositOnly, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      1, // To ensure deposit event is too big
			expErrMsg:                    "generated bytes exceed max size",
		},
		{
			name:                      "skips event if big authorize found",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseAuthorizeOnly, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      1, // To ensure authorize event is big enough
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Remove LastEthereumBlockSynced if not required by test
			if tc.removeLastEthereumBlockSynced {
				s.App.BridgeKeeper.RemoveLastEthereumBlockSynced(s.Ctx())
			}
			// Remove EthereumEventIndexOffset if not required by test
			if tc.removeEthereumEventIndexOffset {
				s.App.BridgeKeeper.RemoveEthereumEventIndexOffset(s.Ctx())
			}
			// Set EthereumEventIndexOffset if the test requires it changed
			if tc.setEthereumEventIndexOffset != nil {
				s.App.BridgeKeeper.SetEthereumEventIndexOffset(s.Ctx(), *tc.setEthereumEventIndexOffset)
			}

			// Set bridge module params
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					AuthorizeMessagesAllowed:     []string{bridgetypes.AllowAllAuthorizeMessages},
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					InjectedEventTxMaxBytes:      tc.injectedEventTxMaxBytes,
				},
			)
			s.Require().NoError(err)

			// Set the MaxBlockGas
			prepareProposalHandlerCtx := s.Ctx().WithConsensusParams(
				comettypes.ConsensusParams{
					Block: &comettypes.BlockParams{
						MaxBytes: 0,
						MaxGas:   tc.maxBlockGas,
					},
				},
			)

			// Set sidecar mock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)
			if tc.expQueryBlockEventsCalled > 0 {
				// If we expect queryBlockEvents to be called, then we must expect the function to be called with
				// certain parameters for expQueryBlockEventsCalled times
				sidecarClientMock.EXPECT().GetBlockEvents(
					gomock.Eq(prepareProposalHandlerCtx), gomock.Eq(tc.expQueryBlockEventsReq),
				).Return(
					tc.queryBlockEventsRet.Response,
					tc.queryBlockEventsRet.Error,
				).Times(tc.expQueryBlockEventsCalled)
			} else {
				// If queryBlockEvents is expected to not be called, then we must expect the function to be called 0
				// times with any parameters
				sidecarClientMock.EXPECT().GetBlockEvents(gomock.Any(), gomock.Any()).Times(0)
			}

			// Execute PrepareProposalHandler
			propHandler := s.GetTestProposalHandler(sidecarClientMock)
			res, err := propHandler.PrepareProposalHandler()(prepareProposalHandlerCtx, tc.requestPrepareProposal)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				s.Require().Nil(res)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(tc.expRes, res)
		})
	}
}

func (s *AppTestSuite) TestProcessProposalHandler() {
	rejectResponse := &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}
	acceptResponse := &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}

	totalTxsGas := int64(3000) // Dummy Txs consume at most 1000 units of gas each. Injected Txs don't consume any gas
	encodedDummyTxs := s.CreateEncodedDummyTxs(3, 1000)

	encodedMsgIndexWithEvents := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndex)
	encodedMsgIndexWithDifferentEvents := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexWithDifferentEvents)
	encodedMsgIndexPartialBlock := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexPartial)
	encodedMsgIndexWithEventsReduced := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexReduced)
	encodedMsgIndexWithoutEvents := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexWithoutEvents)
	encodedMsgIndexSidecarErr := s.EncodeMsgIndexWithEvents(&testtypes.TestMsgIndexSidecarErr)

	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx()

	validTxsWithEvents := append(
		encodedMsgIndexWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithEventsWithMissingSupplyDelta := append(
		encodedMsgIndexWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithDifferentEvents := append(
		encodedMsgIndexWithDifferentEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsPartialBlock := encodedMsgIndexPartialBlock
	validTxsWithEventsReduced := append(
		encodedMsgIndexWithEventsReduced, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithEventsAndSupplyDelta := append(
		encodedMsgIndexWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithoutEvents := append(
		encodedMsgIndexWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsSidecarErr := append(
		encodedMsgIndexSidecarErr, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithLargeAuthorizeSkipped := append(
		encodedMsgIndexWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)

	four := uint64(4)

	testMaxEthBlockUpdateDelay := time.Duration(10)
	testLastEthBlockUpdate := time.Now().Round(0)
	testCurrentTimeDoesNotExceedDelay := testLastEthBlockUpdate.Add(testMaxEthBlockUpdateDelay)
	testCurrentTimeExceedsDelay := testLastEthBlockUpdate.Add(time.Duration(11))

	testCases := []struct {
		name                           string
		removeLastEthereumBlockSynced  bool
		removeEthereumEventIndexOffset bool
		setLastEthBlockUpdateTime      bool
		setEthereumEventIndexOffset    *uint64
		expQueryBlockEventsCalled      int
		expQueryBlockEventsReq         *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet            apptesting.MockQueryBlockEventsResponse
		requestProcessProposal         *abcitypes.RequestProcessProposal
		maxBlockGas                    int64
		supplyDeltaPeriod              uint64
		ethereumProxyContractAddress   string
		injectedEventTxMaxBytes        uint64
		expErrMsg                      string
	}{
		{
			name:                      "accepts block if match MsgIndex with events & supply delta if expected",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEventsAndSupplyDelta,
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches MsgIndex with events",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches partial MsgIndex with events",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, // full response, which will get trimmed
				Error:    nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsPartialBlock,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches MsgIndex without events",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithoutEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                          "accepts block if no new Ethereum block and delay not exceeded",
			removeLastEthereumBlockSynced: false,
			setLastEthBlockUpdateTime:     true,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil,
				Error:    errors.New("block not yet processed 1"),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsSidecarErr,                // Problem is both with validator and the proposer
				Height: 1,                                 // We do not expect MsgSupplyDelta to be injected
				Time:   testCurrentTimeDoesNotExceedDelay, // Block time within syncing delay
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
		},
		{
			name:                          "returns error if no new Ethereum block and delay exceeded",
			removeLastEthereumBlockSynced: false,
			setLastEthBlockUpdateTime:     true,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsSidecarErr,          // Problem is both with validator and the proposer
				Height: 1,                           // We do not expect MsgSupplyDelta to be injected
				Time:   testCurrentTimeExceedsDelay, // Block time exceeds syncing delay
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: fmt.Sprintf(
				"last syncup with Ethereum was at %s; block time: %s; max delay allowed: %s",
				testLastEthBlockUpdate.String(),
				testCurrentTimeExceedsDelay.String(),
				testMaxEthBlockUpdateDelay.String(),
			),
		},
		{
			name:                      "returns error if zero transactions in req.Txs",
			expQueryBlockEventsCalled: 0,
			expQueryBlockEventsReq:    nil,
			queryBlockEventsRet:       apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    nil,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "block proposal doesn't have any transactions: first tx expected to be MsgIndex",
		},
		{
			name:                      "returns error if first tx not a MsgIndex",
			expQueryBlockEventsCalled: 0,
			expQueryBlockEventsReq:    nil,
			queryBlockEventsRet:       apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    encodedDummyTxs,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "first transaction expected to be a valid MsgIndex",
		},
		{
			name:                      "returns error if first tx not a valid sdk.Tx",
			expQueryBlockEventsCalled: 0,
			expQueryBlockEventsReq:    nil,
			queryBlockEventsRet:       apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    [][]byte{[]byte("invalid-MsgSupplyDelta")},
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "first transaction expected to be a valid MsgIndex",
		},
		{
			name:                          "returns error if LastEthereumBlockSynced not found",
			removeLastEthereumBlockSynced: true,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "could not get last Ethereum block synced from state",
		},
		{
			name:                           "returns error if EthereumEventIndexOffset not found",
			removeEthereumEventIndexOffset: true,
			expQueryBlockEventsCalled:      0,
			expQueryBlockEventsReq:         nil,
			queryBlockEventsRet:            apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "could not get Ethereum event index offset from state",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - ValidateBasic error",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{Events: []*sidecartypes.Event{
					{EventType: "invalid-event", Data: nil},
				}},
				Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents, // Problem is with validator not the proposer
				Height: 1,                  // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to generate MsgIndex",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - ValidateStateful error",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{
					Events: []*sidecartypes.Event{
						testutils.MustGetSidecarEventFromParsedEvent(
							testtypes.TestAuthorizeEvent1, "invalid-ethereum-proxy-contract-address",
						),
					},
				},
				Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents, // Problem is with validator not the proposer
				Height: 1,                  // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to generate MsgIndex",
		},
		{
			name:                          "accepts block if matches MsgIndex indicating error",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil,
				Error:    errors.New("block not yet processed 1"),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsSidecarErr, // Problem is both with validator and the proposer
				Height: 1,                  // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
		},
		{
			name:                      "returns error if sidecar errors for validators but not for proposer",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil,
				Error:    errors.New("block not yet processed 1"),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents, // Problem is with validator not the proposer
				Height: 1,                  // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "insufficient no of events, expected at least 3 got 0",
		},
		{
			name:                          "returns error if generated MsgIndex not equal to block proposer's",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithDifferentEvents, // Different events injected by proposer
				Height: 1,                           // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "generated injected txs do not match the ones from the block proposal",
		},
		{
			name: "returns error if generated MsgIndex not equal to block proposer's " +
				"(proposer had events but validators did not)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithoutEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "failed to trim event txs from tail: cannot trim all 3 events",
		},
		{
			name: "returns error if generated MsgIndex not equal to block proposer's " +
				"(proposer had no events but validators did)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestEmptySidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "insufficient no of events, expected at least 3 got 0",
		},
		{
			name: "returns error if generated MsgIndex not equal to block proposer's " +
				"(proposer got more events than the validators)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseReduced, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "insufficient no of events, expected at least 3 got 2",
		},
		{
			name: "returns error if generated MsgIndex not equal to block proposer's " +
				"(validators got more events than the proposer)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEventsReduced,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "generated injected txs do not match the ones from the block proposal",
		},
		{
			name:                      "returns error if block exceeds MaxBlockGas",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas - 1, // Set to total - 1 so that MaxBlockGas is exceeded
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "block gas limit exceeded",
		},
		{
			name:                      "returns error if MsgSupplyDelta expected but not injected",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// MsgSupplyDelta not injected even though expected in height
				Txs:    validTxsWithEventsWithMissingSupplyDelta[:1],
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "expected at least 5 transactions in block proposal",
		},
		{
			name:                      "returns error if incorrect msg injected in tx at index 4",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// MsgSupplyDelta not injected even though expected in height
				Txs:    validTxsWithEventsWithMissingSupplyDelta,
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: fmt.Errorf(
				"failed to parse MsgSupplyDelta at index 4 with error: expected msg type URL %s, got %s",
				sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}),
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
			).Error(),
		},
		{
			name:                        "returns error if Ethereum event index offset larger than events list",
			setEthereumEventIndexOffset: &four, // 4 > len(events)
			expQueryBlockEventsCalled:   1,
			expQueryBlockEventsReq:      &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			maxBlockGas:                  totalTxsGas,
			expErrMsg: "failed to trim event txs from head: insufficient no of events, expected " +
				"at least 4 got 3",
		},
		{
			name:                      "accepts block if match skipped large authorize event",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseAuthorizeOnly, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithLargeAuthorizeSkipped,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      1, // To ensure authorize event is big enough
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "returns error if large deposit event is detected",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseDepositOnly, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// If a big deposit is also matched at PrepareProposal no MsgIndex tx is returned, thus we would
				// error when we try to parse a MsgIndex. To make ProcessProposal error at generateMsgIndexAndEventTxs
				// we need to make use of a different set of txs (one which has a MsgIndex)
				Txs: validTxsWithEvents,

				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      1, // To ensure authorize event is big enough
			maxBlockGas:                  totalTxsGas,
			expErrMsg:                    "generated bytes exceed max size",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Remove LastEthereumBlockSynced if not required by test
			if tc.removeLastEthereumBlockSynced {
				s.App.BridgeKeeper.RemoveLastEthereumBlockSynced(s.Ctx())
			}
			// Remove EthereumEventIndexOffset if not required by test
			if tc.removeEthereumEventIndexOffset {
				s.App.BridgeKeeper.RemoveEthereumEventIndexOffset(s.Ctx())
			}
			// Set EthereumEventIndexOffset if the test requires it changed
			if tc.setEthereumEventIndexOffset != nil {
				s.App.BridgeKeeper.SetEthereumEventIndexOffset(s.Ctx(), *tc.setEthereumEventIndexOffset)
			}
			// Set LastEthBlockUpdateTime if the test requires it
			if tc.setLastEthBlockUpdateTime {
				s.App.BridgeKeeper.SetLastEthBlockUpdateTime(s.Ctx(), testLastEthBlockUpdate)
			}

			// Set bridge module params
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					AuthorizeMessagesAllowed:     []string{bridgetypes.AllowAllAuthorizeMessages},
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					MaxEthBlockUpdateDelay:       testMaxEthBlockUpdateDelay,
					InjectedEventTxMaxBytes:      tc.injectedEventTxMaxBytes,
				},
			)
			s.Require().NoError(err)

			// Set the MaxBlockGas
			processProposalHandlerCtx := s.Ctx().WithConsensusParams(
				comettypes.ConsensusParams{
					Block: &comettypes.BlockParams{
						MaxBytes: 0,
						MaxGas:   tc.maxBlockGas,
					},
				},
			)

			// Set sidecar mock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)
			if tc.expQueryBlockEventsCalled > 0 {
				// If we expect queryBlockEvents to be called, then we must expect the function to be called with
				// certain parameters for expQueryBlockEventsCalled times
				sidecarClientMock.EXPECT().GetBlockEvents(
					gomock.Eq(processProposalHandlerCtx), gomock.Eq(tc.expQueryBlockEventsReq),
				).Return(
					tc.queryBlockEventsRet.Response,
					tc.queryBlockEventsRet.Error,
				).Times(tc.expQueryBlockEventsCalled)
			} else {
				// If queryBlockEvents is expected to not be called, then we must expect the function to be called 0
				// times with any parameters
				sidecarClientMock.EXPECT().GetBlockEvents(gomock.Any(), gomock.Any()).Times(0)
			}

			// Execute ProcessProposalHandler
			propHandler := s.GetTestProposalHandler(sidecarClientMock)
			res, err := propHandler.ProcessProposalHandler()(processProposalHandlerCtx, tc.requestProcessProposal)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				s.Require().Equal(rejectResponse, res)
				return
			}
			s.Require().NoError(err)

			s.Require().Equal(acceptResponse, res)
		})
	}
}
