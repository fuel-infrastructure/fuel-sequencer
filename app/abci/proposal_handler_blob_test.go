package abci_test

import (
	"context"
	"math"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
	"github.com/fuel-infrastructure/blob-storage/pkg/store"
	"github.com/golang/mock/gomock"

	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	blobtypes "github.com/fuel-infrastructure/fuel-sequencer/x/blob/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *AppTestSuite) TestPrepareProposalHandler_BlobFunctionality() {
	ctx := context.TODO()
	// Create test blob data and transactions
	data1 := []byte("test blob data 1")
	data2 := []byte("test blob data 2")
	data3 := []byte("test blob data 3")

	key1 := store.NewKey(data1)
	key2 := store.NewKey(data2)
	key3 := store.NewKey(data3)

	// Store blobs 1 and 2 in the keeper (available)
	s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data1))
	s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data2))
	// Blob 3 is not stored (unavailable)

	// Create blob transactions
	blobTx1 := s.CreateEncodedBlobTx(key1, 100, "test-topic-1", 1, "test-sender-1")
	blobTx2 := s.CreateEncodedBlobTx(key2, 200, "test-topic-2", 2, "test-sender-2")
	blobTx3 := s.CreateEncodedBlobTx(key3, 300, "test-topic-3", 3, "test-sender-3") // unavailable blob

	// Create regular dummy transactions
	encodedDummyTxs := s.CreateEncodedDummyTxs(3, 500)

	// Create mixed transaction set
	mixedTxs := [][]byte{
		encodedDummyTxs[0], // regular tx
		blobTx1,            // available blob tx
		encodedDummyTxs[1], // regular tx
		blobTx2,            // available blob tx
		blobTx3,            // unavailable blob tx
		encodedDummyTxs[2], // regular tx
	}

	// Create expected filtered transactions (blob tx3 should be filtered out)
	expectedFilteredTxs := [][]byte{
		encodedDummyTxs[0], // regular tx
		encodedDummyTxs[1], // regular tx
		encodedDummyTxs[2], // regular tx
		blobTx1,            // available blob tx
		blobTx2,            // available blob tx
	}

	// Create test cases
	testCases := []struct {
		name                         string
		requestPrepareProposal       *abcitypes.RequestPrepareProposal
		expQueryBlockEventsCalled    int
		expQueryBlockEventsReq       *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet          apptesting.MockQueryBlockEventsResponse
		maxBlockGas                  int64
		supplyDeltaPeriod            uint64
		ethereumProxyContractAddress string
		injectedEventTxMaxBytes      uint64
		sequencerTxsAllocation       sdkmath.LegacyDec
		expErrMsg                    string
		expRes                       *abcitypes.ResponsePrepareProposal
	}{
		{
			name:                      "filters out blob transactions with unavailable blobs",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        mixedTxs,
				Height:     1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					[][]byte{}, // MsgIndex without events
					expectedFilteredTxs...,
				),
			},
		},
		{
			name:                      "includes blob transactions with available blobs",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        [][]byte{blobTx1, blobTx2}, // only available blob txs
				Height:     1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					[][]byte{}, // MsgIndex without events
					[][]byte{blobTx1, blobTx2}...,
				),
			},
		},
		{
			name:                      "handles mixed blob and regular transactions correctly",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        mixedTxs,
				Height:     1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					[][]byte{}, // MsgIndex without events
					expectedFilteredTxs...,
				),
			},
		},
		{
			name:                      "filters all blob transactions when none are available",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        [][]byte{blobTx3}, // only unavailable blob tx
				Height:     1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: [][]byte{}, // MsgIndex without events, no blob txs
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Re-store blobs for each test
			s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data1))
			s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data2))

			// Set bridge module params
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					InjectedEventTxMaxBytes:      tc.injectedEventTxMaxBytes,
					SequencerTxsAllocation:       tc.sequencerTxsAllocation,
				},
			)
			s.Require().NoError(err)

			// Set the MaxBlockGas and MaxTxBytes
			prepareProposalHandlerCtx := s.Ctx().WithConsensusParams(
				comettypes.ConsensusParams{
					Block: &comettypes.BlockParams{
						MaxBytes: tc.requestPrepareProposal.MaxTxBytes,
						MaxGas:   tc.maxBlockGas,
					},
				},
			)

			// Set sidecar mock
			ctrl := gomock.NewController(s.T())
			defer ctrl.Finish()
			sidecarClientMock := sidecartestutil.NewMockAppSidecarClient(ctrl)
			if tc.expQueryBlockEventsCalled > 0 {
				sidecarClientMock.EXPECT().GetBlockEvents(
					gomock.Eq(prepareProposalHandlerCtx), gomock.Eq(tc.expQueryBlockEventsReq),
				).Return(
					tc.queryBlockEventsRet.Response,
					tc.queryBlockEventsRet.Error,
				).Times(tc.expQueryBlockEventsCalled)
			} else {
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

			s.Require().Equal(tc.expRes.Txs, res.Txs[1:]) // slice to skip the MsgIndex tx
		})
	}
}

func (s *AppTestSuite) TestProcessProposalHandler_BlobValidation() {
	ctx := context.TODO()
	// Create test blob data
	data := []byte("test blob data for validation")
	key := store.NewKey(data)

	// Store blob in keeper
	s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data))

	// Create valid blob transaction
	validBlobTx := s.CreateEncodedBlobTx(key, 100, "test-topic", 1, "test-sender")

	// Create blob transaction with unavailable blob
	unavailableKey := store.NewKey([]byte("unavailable data"))
	unavailableBlobTx := s.CreateEncodedBlobTx(unavailableKey, 100, "test-topic", 1, "test-sender")

	// Create regular dummy transactions
	encodedDummyTxs := s.CreateEncodedDummyTxs(2, 500)

	// encodedMsgIndexWithEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndex)
	encodedMsgIndexWithoutEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithoutEvents)

	// Create test cases
	testCases := []struct {
		name                         string
		requestProcessProposal       *abcitypes.RequestProcessProposal
		expQueryBlockEventsCalled    int
		expQueryBlockEventsReq       *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet          apptesting.MockQueryBlockEventsResponse
		maxBlockGas                  int64
		supplyDeltaPeriod            uint64
		ethereumProxyContractAddress string
		injectedEventTxMaxBytes      uint64
		expErrMsg                    string
	}{
		{
			name:                      "accepts block with valid blob transactions",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs: append(encodedMsgIndexWithoutEvents(true),
					validBlobTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
				),
				Height: 1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
		},
		{
			name:                      "rejects block with blob transaction having unavailable blob",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs: append(encodedMsgIndexWithoutEvents(true),
					unavailableBlobTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
				),
				Height: 1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg:                    "blob not available in local pool",
		},
		{
			name:                      "accepts block with mixed valid and invalid blob transactions",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: &sidecartypes.QueryBlockEventsResponse{}, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs: append(encodedMsgIndexWithoutEvents(true),
					validBlobTx,
					encodedDummyTxs[0],
					validBlobTx,
				),
				Height: 1,
			},
			maxBlockGas:                  3000,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Re-store blob for each test
			s.Require().NoError(s.App.BlobKeeper.Insert(ctx, data))

			// Set bridge module params
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					InjectedEventTxMaxBytes:      tc.injectedEventTxMaxBytes,
				},
			)
			s.Require().NoError(err)

			// Set the MaxBlockGas and MaxTxBytes
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
				sidecarClientMock.EXPECT().GetBlockEvents(
					gomock.Eq(processProposalHandlerCtx), gomock.Eq(tc.expQueryBlockEventsReq),
				).Return(
					tc.queryBlockEventsRet.Response,
					tc.queryBlockEventsRet.Error,
				).Times(tc.expQueryBlockEventsCalled)
			} else {
				sidecarClientMock.EXPECT().GetBlockEvents(gomock.Any(), gomock.Any()).Times(0)
			}

			// Execute ProcessProposalHandler
			propHandler := s.GetTestProposalHandler(sidecarClientMock)
			res, err := propHandler.ProcessProposalHandler()(processProposalHandlerCtx, tc.requestProcessProposal)

			if len(tc.expErrMsg) > 0 {
				s.Require().Error(err)
				s.Require().ErrorContains(err, tc.expErrMsg)
				s.Require().Equal(&abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}, res)
				return
			}
			s.Require().NoError(err)
			s.Require().Equal(&abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}, res)
		})
	}
}

// CreateEncodedBlobTx creates an encoded blob transaction for testing
func (s *AppTestSuite) CreateEncodedBlobTx(
	key store.Key, size uint64, topic string, nonce uint64, sender string,
) []byte {
	// Create blob message
	blobMsg := &blobtypes.MsgBlobMetadataTx{
		Hash:   key.String(),
		Size_:  size,
		Topic:  topic,
		Nonce:  nonce,
		Sender: sender,
	}

	// Create Any from message
	blobMsgAny, err := types.NewAnyWithValue(blobMsg)
	if err != nil {
		panic("could not construct any from MsgBlobMetadataTx")
	}

	// Construct Tx Body with the message
	txBodyBz, err := proto.Marshal(&txtypes.TxBody{
		Messages: []*types.Any{blobMsgAny},
	})
	if err != nil {
		panic("could not construct TxBody")
	}

	// Construct Auth Info with Fee to avoid nil pointer panics
	authInfoBz, err := proto.Marshal(&txtypes.AuthInfo{
		SignerInfos: []*txtypes.SignerInfo{
			{Sequence: 1},
		},
		Fee: &txtypes.Fee{
			GasLimit: 0,
		},
	})
	if err != nil {
		panic("could not construct AuthInfo")
	}

	// Construct final Tx
	txRawBz, err := proto.Marshal(&txtypes.TxRaw{
		BodyBytes:     txBodyBz,
		AuthInfoBytes: authInfoBz,
		Signatures:    nil,
	})
	if err != nil {
		panic("could not construct TxRaw bytes")
	}

	return txRawBz
}

func (s *AppTestSuite) TestBlobTransactionCreation() {
	// Test creating and encoding blob transactions
	data := []byte("test blob data")
	key := store.NewKey(data)

	// Create blob transaction
	blobTx := s.CreateEncodedBlobTx(key, 100, "test-topic", 1, "test-sender")
	s.Require().NotNil(blobTx)
	s.Require().Greater(len(blobTx), 0)

	// Test decoding the transaction using the app's codec
	var tx txtypes.TxRaw
	err := proto.Unmarshal(blobTx, &tx)
	s.Require().NoError(err)
	s.Require().NotNil(tx.BodyBytes)

	// Decode the body to get messages
	var body txtypes.TxBody
	err = proto.Unmarshal(tx.BodyBytes, &body)
	s.Require().NoError(err)
	s.Require().Len(body.Messages, 1)

	// Decode the message
	var msg blobtypes.MsgBlobMetadataTx
	err = proto.Unmarshal(body.Messages[0].Value, &msg)
	s.Require().NoError(err)

	// Verify the message fields
	s.Require().Equal(key.String(), msg.Hash)
	s.Require().Equal(uint64(100), msg.GetSize_())
	s.Require().Equal("test-topic", msg.Topic)
	s.Require().Equal(uint64(1), msg.Nonce)
	s.Require().Equal("test-sender", msg.Sender)
}
