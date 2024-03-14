package abci_test

import (
	"errors"
	"fmt"
	"math"
	"time"

	abcitypes "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
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
	totalTxsGas := int64(3000) // Dummy Txs consume at most 1000 units of gas each. EthEventsTx doesn't consume any gas
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

	totalBytesTxWithEventsAndDummyTxs := int64(calculateTotalTxBytes(
		[][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]},
	))

	testCases := []struct {
		name                          string
		removeLastEthereumBlockSynced bool
		expQueryBlockEventsCalled     int
		expQueryBlockEventsReq        *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet           apptesting.TestQueryBlockEventsRet
		requestPrepareProposal        *abcitypes.RequestPrepareProposal
		maxBlockGas                   int64
		expErrMsg                     string
		expRes                        *abcitypes.ResponsePrepareProposal
	}{
		{
			name:                          "returns injected eth tx and other txs if block events found",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         math.MaxInt64,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2]},
			},
		},
		{
			name:                          "returns injected eth tx and other txs if no block events found",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         math.MaxInt64,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
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
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: nil, Error: fmt.Errorf("%s 1", sidecartypes.ErrBlockDoesNotExist),
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         math.MaxInt64,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
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
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: nil, Error: errors.New("block not yet processed 1"),
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         math.MaxInt64,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
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
			name:                          "returns error if LastEthereumBlockSynced not found",
			removeLastEthereumBlockSynced: true,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.TestQueryBlockEventsRet{},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         0,
				Txs:                nil,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "could not get last Ethereum block synced from state",
		},
		{
			name:                          "returns error if EthEventsTx cannot be generated",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: &sidecartypes.QueryBlockEventsResponse{Events: []*sidecartypes.Event{
					{EventType: "invalid-event", Data: nil},
				}},
				Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         0,
				Txs:                nil,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "failed to generate eth events tx",
		},
		{
			name:                          "returns error if EthEventsTx cannot be selected by the TxSelector",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         0, // Set to zero to make sure there is no capacity for first transaction
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "failed to add eth events transaction to block proposal",
		},
		{
			name:                          "returns selected Txs if MaxTxBytes exceeded",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				// Set to expected size - 1 to omit last tx
				MaxTxBytes:         totalBytesTxWithEventsAndDummyTxs - 1,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1]},
			},
		},
		{
			name:                          "returns selected Txs if MaxBlockGas exceeded",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes:         math.MaxInt64,
				Txs:                encodedDummyTxs,
				LocalLastCommit:    abcitypes.ExtendedCommitInfo{},
				Misbehavior:        nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas - 1, // Set to total - 1 so that the last transaction is omitted
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1]},
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

	totalTxsGas := int64(3000) // Dummy Txs consume at most 1000 units of gas each. EthEventsTx doesn't consume any gas
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

	validTxsWithEvents := [][]byte{
		encodedEthEventsTxWithEvents, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
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
		queryBlockEventsRet           apptesting.TestQueryBlockEventsRet
		requestProcessProposal        *abcitypes.RequestProcessProposal
		maxBlockGas                   int64
		expErrMsg                     string
	}{
		{
			name:                          "accepts block if matches EthEventsTx with events",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithEvents,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
		},
		{
			name:                          "accepts block if matches EthEventsTx without events",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithoutEvents,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
		},
		{
			name:                          "accepts block if matches EthEventsTx indicating no Ethereum block",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: nil, Error: fmt.Errorf("%s 1", sidecartypes.ErrBlockDoesNotExist),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsNoNewBlock,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
		},
		{
			name:                          "returns error if zero transactions in req.Txs",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.TestQueryBlockEventsRet{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                nil,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "block proposal doesn't have any transactions: first tx expected to be an eth events tx",
		},
		{
			name:                          "returns error if first tx no an EthEventsTx",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.TestQueryBlockEventsRet{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                encodedDummyTxs,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "first transaction expected to be an eth events tx",
		},
		{
			name:                          "returns error if LastEthereumBlockSynced not found",
			removeLastEthereumBlockSynced: true,
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.TestQueryBlockEventsRet{},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithEvents,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "could not get last Ethereum block synced from state",
		},
		{
			name:                          "returns error if EthEventsTx cannot be generated",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: &sidecartypes.QueryBlockEventsResponse{Events: []*sidecartypes.Event{
					{EventType: "invalid-event", Data: nil},
				}},
				Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithEvents, // Problem is with validator not the proposer
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "failed to generate eth events tx",
		},
		{
			name:                          "returns error if sidecar errors (AdvanceSequencer false)",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: nil,
				Error:    errors.New("block not yet processed 1"),
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithEvents, // Problem is with validator not the proposer
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "generated eth events tx implies block rejection",
		},
		{
			name:                          "returns error if generated EthEventsTx not equal to block proposers'",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithoutEvents,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas,
			expErrMsg:   "generated eth events tx does not match the one in the block proposal",
		},
		{
			name:                          "returns error if block exceeds MaxBlockGas",
			removeLastEthereumBlockSynced: false,
			expQueryBlockEventsCalled:     1,
			expQueryBlockEventsReq:        &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.TestQueryBlockEventsRet{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:                validTxsWithEvents,
				ProposedLastCommit: abcitypes.CommitInfo{},
				Misbehavior:        nil,
				Hash:               nil,
				Height:             0,
				Time:               time.Time{},
				NextValidatorsHash: nil,
				ProposerAddress:    nil,
			},
			maxBlockGas: totalTxsGas - 1, // Set to total - 1 so that MaxBlockGas is exceeded
			expErrMsg:   "block gas limit exceeded",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Remove LastEthereumBlockSynced if not required by test
			if tc.removeLastEthereumBlockSynced {
				s.App.BridgeKeeper.RemoveLastEthereumBlockSynced(s.Ctx())
			}

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
