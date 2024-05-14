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

	encodedEthEventsTxWithEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTx)
	encodedEthEventsTxPartialBlock := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxPartial)
	encodedEthEventsTxWithoutEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxWithoutEvents)
	encodedEthEventsTxSidecarErr := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxSidecarErr)

	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx()

	totalTxsBytesWithEventsAndSupplyDelta := int64(calculateTotalTxBytes(
		[][]byte{
			encodedEthEventsTxWithEvents,
			msgSupplyDeltaTx,
			encodedDummyTxs[0],
			encodedDummyTxs[1],
			encodedDummyTxs[2],
		},
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{
					encodedEthEventsTxWithEvents,
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				},
			},
		},
		{
			name:                      "returns injected eth tx and other txs if block events found",
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]},
			},
		},
		{
			name:                      "returns injected eth tx and other txs if no block events found",
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{
					encodedEthEventsTxWithoutEvents,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				},
			},
		},
		{
			name:                      "returns injected eth tx and other txs if sidecar errors",
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{
					encodedEthEventsTxSidecarErr,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				},
			},
		},
		{
			name:                      "returns partial injected eth tx if not enough space in the block",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to the size of a partial eth tx, so that the last event does not fit. We need to add +2 since the
				// generated EthEventsTx will have NewEthereumBlock set to true at first until the proposer updates it.
				// When NewEthereumBlock is set to true, it consumes 2 bytes, otherwise it does not consume anything.
				MaxTxBytes: int64(calculateTotalTxBytes([][]byte{encodedEthEventsTxPartialBlock})) + 2,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{
					encodedEthEventsTxPartialBlock, // partial eth tx
				},
			},
		},
		{
			name:                      "returns error if SupplyDeltaPeriod is zero",
			expQueryBlockEventsCalled: 0,
			expQueryBlockEventsReq:    nil,
			queryBlockEventsRet:       apptesting.MockQueryBlockEventsResponse{},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            uint64(0),
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg:                    "SupplyDeltaPeriod cannot be zero",
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
			expErrMsg:                    "could not get Ethereum event index offset from state",
		},
		{
			name:                      "returns error if EthEventsTx cannot be generated - ValidateBasic error",
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
			expErrMsg:                    "failed to generate eth events tx",
		},
		{
			name:                      "returns error if EthEventsTx cannot be generated - ValidateStateful error",
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
			expErrMsg:                    "failed to generate eth events tx",
		},
		{
			name:                      "returns error if not enough block space for at least one EthEventsTx event",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0, // Set to zero to make sure there is no capacity for the EthEventsTx events
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg: "failed to trim eth events tx tail: cannot trim all 3 events from " +
				"EthEventsTx",
		},
		{
			name:                      "returns error if none of the EthEventsTx events fit in the block",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to the size of MsgSupplyDeltaTx so that EthEventsTx does not fit
				MaxTxBytes: int64(calculateTotalTxBytes([][]byte{msgSupplyDeltaTx})),
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg: "failed to trim eth events tx tail: cannot trim all 3 events from " +
				"EthEventsTx",
		},
		{
			name:                      "returns only supply delta and EthEventsTx if it's just enough block size",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to EXACTLY the size of MsgSupplyDelta transaction plus EthEventsTx
				MaxTxBytes: int64(calculateTotalTxBytes([][]byte{msgSupplyDeltaTx, encodedEthEventsTxWithEvents})),
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, msgSupplyDeltaTx},
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]},
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
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]},
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
			expErrMsg: "failed to trim eth events tx head: insufficient no of events, expected at " +
				"least 4 got 3",
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

			// Set SupplyDeltaPeriod and EthereumProxyContractAddress
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
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

	encodedEthEventsTxWithEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTx)
	encodedEthEventsTxWithDifferentEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxWithDifferentEvents)
	encodedEthEventsTxPartialBlock := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxPartial)
	encodedEthEventsTxWithEventsReduced := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxReduced)
	encodedEthEventsTxWithoutEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxWithoutEvents)
	encodedEthEventsTxSidecarErr := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxSidecarErr)

	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx()

	validTxsWithEvents := [][]byte{
		encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsWithDifferentEvents := [][]byte{
		encodedEthEventsTxWithDifferentEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsPartialBlock := [][]byte{
		encodedEthEventsTxPartialBlock,
	}
	validTxsWithEventsReduced := [][]byte{
		encodedEthEventsTxWithEventsReduced, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsWithEventsAndSupplyDelta := [][]byte{
		encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsWithoutEvents := [][]byte{
		encodedEthEventsTxWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsSidecarErr := [][]byte{
		encodedEthEventsTxSidecarErr, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}

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
		expErrMsg                      string
	}{
		{
			name:                      "accepts block if match EthEventsTx with events & supply delta if expected",
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
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches EthEventsTx with events",
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
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches partial EthEventsTx with events",
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
			maxBlockGas:                  totalTxsGas,
		},
		{
			name:                      "accepts block if matches EthEventsTx without events",
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
			expErrMsg: "block proposal doesn't have any transactions: first tx expected to be " +
				"an eth events tx",
		},
		{
			name:                      "returns error if first tx no an EthEventsTx",
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
			expErrMsg:                    "first transaction expected to be an eth events tx",
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
			expErrMsg:                    "could not get Ethereum event index offset from state",
		},
		{
			name:                      "returns error if EthEventsTx cannot be generated - ValidateBasic error",
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
			expErrMsg:                    "failed to generate eth events tx",
		},
		{
			name:                      "returns error if EthEventsTx cannot be generated - ValidateStateful error",
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
			expErrMsg:                    "failed to generate eth events tx",
		},
		{
			name:                          "accepts block if matches EthEventsTx indicating error",
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
			expErrMsg:                    "insufficient no of events, expected at least 3 got 0",
		},
		{
			name:                          "returns error if generated EthEventsTx not equal to block proposer's",
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
			expErrMsg:                    "generated eth events tx does not match the one in the block proposal",
		},
		{
			name: "returns error if generated EthEventsTx not equal to block proposer's " +
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
			expErrMsg: "failed to trim eth events tx tail: cannot trim all 3 events from " +
				"EthEventsTx events",
		},
		{
			name: "returns error if generated EthEventsTx not equal to block proposer's " +
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
			expErrMsg:                    "insufficient no of events, expected at least 3 got 0",
		},
		{
			name: "returns error if generated EthEventsTx not equal to block proposer's " +
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
			expErrMsg:                    "insufficient no of events, expected at least 3 got 2",
		},
		{
			name: "returns error if generated EthEventsTx not equal to block proposer's " +
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
			expErrMsg:                    "generated eth events tx does not match the one in the block proposal",
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
			expErrMsg:                    "block gas limit exceeded",
		},
		{
			name:                      "returns error if SupplyDeltaPeriod is zero",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            0,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg:                    "SupplyDeltaPeriod cannot be zero",
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
				Txs:    validTxsWithEvents[:1],
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg:                    "expected at least two transactions in block proposal",
		},
		{
			name:                      "returns error if incorrect msg injected in tx at index 1",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// MsgSupplyDelta not injected even though expected in height
				Txs:    validTxsWithEvents,
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			expErrMsg: fmt.Errorf(
				"incorrect msg type url in transaction at index 1; expected %s got %s",
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
			maxBlockGas:                  totalTxsGas,
			expErrMsg: "failed to trim eth events tx head: insufficient no of events, expected " +
				"at least 4 got 3",
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

			// Set SupplyDeltaPeriod and EthereumProxyContractAddress
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					MaxEthBlockUpdateDelay:       testMaxEthBlockUpdateDelay,
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

func (s *AppTestSuite) TestPreBlockerEthEventsTxHandling_SingleTransaction() {
	encodedEthEventsTxWithEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTx)
	encodedEthEventsTxPartialBlock := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxPartial)
	encodedEthEventsTxWithoutEvents := s.EncodeEthEventsTx(&testtypes.TestEthEventsTxWithoutEvents)

	ethEventsTxWithWrongBlock := testtypes.TestEthEventsTx
	ethEventsTxWithWrongBlock.BlockNumber = 99
	encodedEthEventsTxWithWrongBlock := s.EncodeEthEventsTx(&ethEventsTxWithWrongBlock)

	testBlockTime := time.Now().Round(0)

	testCases := []struct {
		name                            string
		requestTxs                      [][]byte
		expectEvents                    bool
		expectNewBlock                  bool
		expectEthereumEventsIndexOffset uint64
		expectErrMsg                    string
	}{
		{
			name:         "no EthEventsTx => error",
			requestTxs:   [][]byte{},
			expectErrMsg: "expected eth events transaction to be injected",
		},
		{
			name:         "EthEventsTx with wrong block number => error",
			requestTxs:   [][]byte{encodedEthEventsTxWithWrongBlock},
			expectErrMsg: "expected block number 1, got 99 in EthEventsTx",
		},
		{
			name:                            "EthEventsTx with events => new block and offset stays at zero",
			requestTxs:                      [][]byte{encodedEthEventsTxWithEvents},
			expectEvents:                    true,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
		},
		{
			name:                            "EthEventsTx without events => new block and offset stays at zero",
			requestTxs:                      [][]byte{encodedEthEventsTxWithoutEvents},
			expectEvents:                    false,
			expectNewBlock:                  true,
			expectEthereumEventsIndexOffset: 0,
		},
		{
			name:                            "EthEventsTx with partial events => no new block but offset updated",
			requestTxs:                      [][]byte{encodedEthEventsTxPartialBlock},
			expectEvents:                    true,
			expectNewBlock:                  false,
			expectEthereumEventsIndexOffset: uint64(len(testtypes.TestEthEventsTxPartial.Events)),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Simulate calling PreBlocker with the provided transactions
			req := &abcitypes.RequestFinalizeBlock{Txs: tc.requestTxs, Time: testBlockTime}

			// Set sidecar mock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)

			propHandler := s.GetTestProposalHandler(sidecarClientMock)
			_, err := propHandler.PreBlocker(s.Ctx(), req)
			if tc.expectErrMsg != "" {
				s.Require().ErrorContains(err, tc.expectErrMsg)
				return
			}
			s.Require().NoError(err)

			// Verify the EthEventsTx in state
			storedTx, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), uint64(1))
			if tc.expectEvents {
				s.Require().True(found)
				s.Require().NotEmpty(storedTx.Events)
			} else {
				s.Require().False(found)
			}

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

func (s *AppTestSuite) TestPreBlockerEthEventsTxHandling_Combinations() {

	ethEventsTxWithEvents1 := testtypes.TestEthEventsTx
	ethEventsTxPartialBlock1 := testtypes.TestEthEventsTxPartial
	ethEventsTxWithoutEvents1 := testtypes.TestEthEventsTxWithoutEvents
	ethEventsTxNoNewBlock1 := testtypes.TestEthEventsTxNoNewBlock

	// Set of events with block number set to 2
	ethEventsTxWithEvents2 := testtypes.TestEthEventsTx
	ethEventsTxPartialBlock2 := testtypes.TestEthEventsTxPartial
	ethEventsTxWithoutEvents2 := testtypes.TestEthEventsTxWithoutEvents
	ethEventsTxNoNewBlock2 := testtypes.TestEthEventsTxNoNewBlock
	ethEventsTxWithEvents2.BlockNumber = 2
	ethEventsTxPartialBlock2.BlockNumber = 2
	ethEventsTxWithoutEvents2.BlockNumber = 2
	ethEventsTxNoNewBlock2.BlockNumber = 2

	// Ensure that original block numbers are as expected, and unchanged
	s.Require().EqualValues(ethEventsTxWithEvents1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxPartialBlock1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxWithoutEvents1.BlockNumber, 1)
	s.Require().EqualValues(ethEventsTxNoNewBlock1.BlockNumber, 1)

	numberOfEventsInPartialTx := uint64(len(testtypes.TestEthEventsTxPartial.Events))

	testBlockTime1 := time.Now().Round(0)
	testBlockTime2 := time.Now().Round(0)

	testCases := []struct {
		name                            string
		ethEventsTx                     []bridgetypes.EthEventsTx
		blockTime                       []time.Time
		expectEvents                    []bool
		expectLastEthereumBlockSynced   []uint64
		expectEthereumEventsIndexOffset []uint64
		expectLastEthBlockUpdateTime    []bool
		expectBlockTime                 []time.Time
		expectErrMsg                    []string
	}{
		// ---------------------------- Combinations of Partial and Full
		{
			name: "Partial + Full => 0,1 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, true},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + Partial => 1,1 synced and 0,N offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxPartialBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, true},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Partial and NoNewBlock
		{
			name: "Partial + NoNewBlock => ERR because we expect at least one new event",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxNoNewBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true},
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
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxPartialBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, true},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of Partial and NoEvents
		{
			name: "Partial + NoEvents => ERR because we expect at least one new event",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxWithoutEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true},
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
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxPartialBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, true},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, numberOfEventsInPartialTx},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoNewBlock and NoEvents
		{
			name: "NoNewBlock + NoEvents => 0,1 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxWithoutEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, false},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "NoEvents + NoNewBlock => 1,1 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxNoNewBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, false},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoNewBlock and Full
		{
			name: "NoNewBlock + Full => 0,1 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxWithEvents1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, true},
			expectLastEthereumBlockSynced:   []uint64{0, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{false, true},
		},
		{
			name: "Full + NoNewBlock => 1,1 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxNoNewBlock2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, false},
			expectLastEthereumBlockSynced:   []uint64{1, 1},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime1},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of NoEvents and Full
		{
			name: "NoEvents + Full => 1,2 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxWithEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, true},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		{
			name: "Full + NoEvents => 1,2 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxWithoutEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, false},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Full and Full
		{
			name: "Full + Full => 1,2 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithEvents1,
				ethEventsTxWithEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, true},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
		// ---------------------------- Combinations of Partial and Partial
		{
			name: "Partial + Partial => 0,0 synced and N,2N offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxPartialBlock1,
				ethEventsTxPartialBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{true, true},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{numberOfEventsInPartialTx, numberOfEventsInPartialTx * 2},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of NoNewBlock and NoNewBlock
		{
			name: "NoNewBlock + NoNewBlock => 0,0 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxNoNewBlock1,
				ethEventsTxNoNewBlock1, // from same block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, false},
			expectLastEthereumBlockSynced:   []uint64{0, 0},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{time.Time{}, time.Time{}},
			expectLastEthBlockUpdateTime:    []bool{false, false},
		},
		// ---------------------------- Combinations of NoEvents and NoEvents
		{
			name: "NoEvents + NoEvents => 1,2 synced and 0,0 offset",
			ethEventsTx: []bridgetypes.EthEventsTx{
				ethEventsTxWithoutEvents1,
				ethEventsTxWithoutEvents2, // from new block
			},
			blockTime:                       []time.Time{testBlockTime1, testBlockTime2},
			expectEvents:                    []bool{false, false},
			expectLastEthereumBlockSynced:   []uint64{1, 2},
			expectEthereumEventsIndexOffset: []uint64{0, 0},
			expectBlockTime:                 []time.Time{testBlockTime1, testBlockTime2},
			expectLastEthBlockUpdateTime:    []bool{true, true},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set sidecar mock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)

			propHandler := s.GetTestProposalHandler(sidecarClientMock)

			for i := 0; i < len(tc.ethEventsTx); i++ {

				ethEventsTx := tc.ethEventsTx[i]
				blockTime := tc.blockTime[i]
				req := &abcitypes.RequestFinalizeBlock{
					Txs:  [][]byte{s.EncodeEthEventsTx(&ethEventsTx)},
					Time: blockTime,
				}

				_, err := propHandler.PreBlocker(s.Ctx(), req)
				if tc.expectErrMsg != nil && tc.expectErrMsg[i] != "" {
					s.Require().ErrorContains(err, tc.expectErrMsg[i])
					return
				}
				s.Require().NoError(err)

				expectEvents := tc.expectEvents[i]
				expectLastEthereumBlockSynced := tc.expectLastEthereumBlockSynced[i]
				expectEthereumEventsIndexOffset := tc.expectEthereumEventsIndexOffset[i]
				expectLastEthBlockUpdateTime := tc.expectLastEthBlockUpdateTime[i]
				expectBlockTime := tc.expectBlockTime[i]

				// Verify the EthEventsTx in state at the EthEventsTx block number
				storedTx, found := s.App.BridgeKeeper.GetEthEventsTx(s.Ctx(), ethEventsTx.BlockNumber)
				if expectEvents {
					s.Require().True(found)
					s.Require().NotEmpty(storedTx.Events)
				} else {
					s.Require().False(found)
				}

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
				s.App.BridgeKeeper.RemoveEthEventsTx(s.Ctx(), ethEventsTx.BlockNumber)
			}
		})
	}
}
