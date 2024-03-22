package abci_test

import (
	"errors"
	"fmt"
	"math"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
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

	encodedEthEventsTxWithEvents := s.EncodeEthEventsTx(testtypes.TestEthEventsTx)
	encodedEthEventsTxWithoutEvents := s.EncodeEthEventsTx(&bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
	})
	encodedEthEventsTxNoNewBlock := s.EncodeEthEventsTx(&bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: false,
	})
	encodedEthEventsTxSidecarErr := s.EncodeEthEventsTx(&bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: false,
		NewEthereumBlock: false,
	})

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

	testCases := []struct {
		name                          string
		removeLastEthereumBlockSynced bool
		expQueryBlockEventsCalled     int
		expQueryBlockEventsReq        *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet           apptesting.MockQueryBlockEventsResponse
		requestPrepareProposal        *abcitypes.RequestPrepareProposal
		maxBlockGas                   int64
		supplyDeltaPeriod             uint64
		expErrMsg                     string
		expRes                        *abcitypes.ResponsePrepareProposal
	}{
		{
			name:                          "returns supply delta tx if expected at height and other txs",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
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
			name:                          "returns injected eth tx and other txs if block events found",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]},
			},
		},
		{
			name:                          "returns injected eth tx and other txs if no block events found",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
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
			name:                          "returns injected eth tx and other txs if no new Ethereum block",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil, Error: fmt.Errorf("%s 1", sidecartypes.ErrBlockDoesNotExist),
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{
					encodedEthEventsTxNoNewBlock,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				},
			},
		},
		{
			name:                          "returns injected eth tx and other txs if sidecar errors",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil, Error: errors.New("block not yet processed 1"),
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
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
			name:                          "returns error if SupplyDeltaPeriod is zero",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0,
				Txs:        nil,
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: uint64(0),
			expErrMsg:         "SupplyDeltaPeriod cannot be zero",
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
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "could not get last Ethereum block synced from state",
		},
		{
			name:                          "returns error if EthEventsTx cannot be generated",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
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
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "failed to generate eth events tx",
		},
		{
			name:                          "returns error if EthEventsTx cannot be selected by the TxSelector",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: 0, // Set to zero to make sure there is no capacity for first transaction
				Txs:        encodedDummyTxs,
				Height:     1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "failed to add eth events transaction to block proposal",
		},
		{
			name:                          "returns error if MsgSupplyDeltaTx cannot be selected by the TxSelector",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to the size of EthEventsTx so that MsgSupplyDeltaTx is not chosen
				MaxTxBytes: int64(calculateTotalTxBytes([][]byte{encodedEthEventsTxWithEvents})),
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "failed to add message supply delta transaction to block proposal",
		},
		{
			name:                          "returns selected Txs if MaxTxBytes exceeded",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to expected size - 1 to omit last tx
				MaxTxBytes: totalTxsBytesWithEventsAndSupplyDelta - 1,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]},
			},
		},
		{
			name:                          "returns selected Txs if MaxBlockGas exceeded",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas - 1, // Set to total - 1 so that the last transaction is omitted
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1]},
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

			// Set SupplyDeltaPeriod
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), bridgetypes.Params{SupplyDeltaPeriod: tc.supplyDeltaPeriod})
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

	encodedEthEventsTxWithEvents := s.EncodeEthEventsTx(testtypes.TestEthEventsTx)
	encodedEthEventsTxWithoutEvents := s.EncodeEthEventsTx(&bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: true,
	})
	encodedEthEventsTxNoNewBlock := s.EncodeEthEventsTx(&bridgetypes.EthEventsTx{
		Events:           []*sidecartypes.Event{},
		AdvanceSequencer: true,
		NewEthereumBlock: false,
	})

	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx()

	validTxsWithEvents := [][]byte{
		encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsWithEventsAndSupplyDelta := [][]byte{
		encodedEthEventsTxWithEvents, msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsWithoutEvents := [][]byte{
		encodedEthEventsTxWithoutEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}
	validTxsNoNewBlock := [][]byte{
		encodedEthEventsTxNoNewBlock, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	}

	testCases := []struct {
		name                          string
		removeLastEthereumBlockSynced bool
		expQueryBlockEventsCalled     int
		expQueryBlockEventsReq        *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet           apptesting.MockQueryBlockEventsResponse
		requestProcessProposal        *abcitypes.RequestProcessProposal
		maxBlockGas                   int64
		supplyDeltaPeriod             uint64
		expErrMsg                     string
	}{
		{
			name:                          "accepts block if match EthEventsTx with events & supply delta if expected",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEventsAndSupplyDelta,
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			maxBlockGas:       totalTxsGas,
		},
		{
			name:                          "accepts block if matches EthEventsTx with events",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			maxBlockGas:       totalTxsGas,
		},
		{
			name:                          "accepts block if matches EthEventsTx without events",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithoutEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			maxBlockGas:       totalTxsGas,
		},
		{
			name:                          "accepts block if matches EthEventsTx indicating no Ethereum block",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil, Error: fmt.Errorf("%s 1", sidecartypes.ErrBlockDoesNotExist),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsNoNewBlock,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			maxBlockGas:       totalTxsGas,
		},
		{
			name:                          "returns error if zero transactions in req.Txs",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    nil,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "block proposal doesn't have any transactions: first tx expected to be an eth events tx",
		},
		{
			name:                          "returns error if first tx no an EthEventsTx",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    encodedDummyTxs,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "first transaction expected to be an eth events tx",
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
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "could not get last Ethereum block synced from state",
		},
		{
			name:                          "returns error if EthEventsTx cannot be generated",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
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
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "failed to generate eth events tx",
		},
		{
			name:                          "returns error if sidecar errors (AdvanceSequencer false)",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: nil,
				Error:    errors.New("block not yet processed 1"),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents, // Problem is with validator not the proposer
				Height: 1,                  // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "generated eth events tx implies block rejection",
		},
		{
			name:                          "returns error if generated EthEventsTx not equal to block proposers'",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithoutEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "generated eth events tx does not match the one in the block proposal",
		},
		{
			name:                          "returns error if block exceeds MaxBlockGas",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas - 1, // Set to total - 1 so that MaxBlockGas is exceeded
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "block gas limit exceeded",
		},
		{
			name:                          "returns error if SupplyDeltaPeriod is zero",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithEvents,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: 0,
			expErrMsg:         "SupplyDeltaPeriod cannot be zero",
		},
		{
			name:                          "returns error if MsgSupplyDelta expected but not injected",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// MsgSupplyDelta not injected even though expected in height
				Txs:    validTxsWithEvents[:1],
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg:         "expected at least two transactions in block proposal",
		},
		{
			name:                          "returns error if incorrect msg injected in tx at index 1",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				// MsgSupplyDelta not injected even though expected in height
				Txs:    validTxsWithEvents,
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2),
			},
			maxBlockGas:       totalTxsGas,
			supplyDeltaPeriod: testtypes.TestSupplyDeltaPeriod,
			expErrMsg: fmt.Errorf(
				"incorrect msg type url in transaction at index 1; expected %s got %s",
				sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}),
				sdk.MsgTypeURL(&banktypes.MsgSend{}),
			).Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Remove LastEthereumBlockSynced if not required by test
			if tc.removeLastEthereumBlockSynced {
				s.App.BridgeKeeper.RemoveLastEthereumBlockSynced(s.Ctx())
			}

			// Set SupplyDeltaPeriod
			err := s.App.BridgeKeeper.SetParams(s.Ctx(), bridgetypes.Params{SupplyDeltaPeriod: tc.supplyDeltaPeriod})
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
