package abci_test

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	abcitypes "github.com/cometbft/cometbft/abci/types"
	comettypes "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/golang/mock/gomock"

	"github.com/fuel-infrastructure/fuel-sequencer/app/abci"
	"github.com/fuel-infrastructure/fuel-sequencer/app/apptesting"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	sidecartestutil "github.com/fuel-infrastructure/fuel-sequencer/sidecar/testutil"
	testutils "github.com/fuel-infrastructure/fuel-sequencer/testutil"
	testtypes "github.com/fuel-infrastructure/fuel-sequencer/testutil/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

func (s *AppTestSuite) TestPrepareProposalHandler() {
	// Dummy Txs consume at most 1000 units of gas each. Injected Txs don't consume any gas
	totalTxsGas := int64(3000)
	encodedDummyTxs := s.CreateEncodedDummyTxs(3, 1000)

	encodedMsgIndexWithEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndex)
	encodedMsgIndexWithFourEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithFourEvents)
	encodedMsgIndexPartialBlock := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexPartial)
	encodedMsgIndexPartialBlockWith3Events := s.GetMsgIndexWithEventsEncoder(testtypes.TestMsgIndexPartial2)
	encodedMsgIndexWithoutEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithoutEvents)
	encodedMsgIndexSidecarErr := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexSidecarErr)
	encodedMsgIndexWithSkippedEvent := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithSkippedEvent)

	msgIndexSequence := uint64(1)
	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx(msgIndexSequence + 1)

	totalTxsBytesWithEventsAndSupplyDelta := utils.TxsSize(
		append(
			encodedMsgIndexWithEvents(true),
			msgSupplyDeltaTx,
			encodedDummyTxs[0],
			encodedDummyTxs[1],
			encodedDummyTxs[2],
		),
	)
	totalTxsBytesWithFourEventsAndSupplyDeltaOnly := utils.TxsSize(
		append(
			encodedMsgIndexWithFourEvents(true),
			msgSupplyDeltaTx,
		),
	)
	totalTxsBytesWithEventsSupplyDeltaAndFiveDummyTxs := utils.TxsSize(
		append(
			encodedMsgIndexWithEvents(true),
			msgSupplyDeltaTx,
			encodedDummyTxs[0],
			encodedDummyTxs[1],
			encodedDummyTxs[2],
			encodedDummyTxs[0],
			encodedDummyTxs[1],
		),
	)

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
		sequencerTxsAllocation         sdkmath.LegacyDec
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
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(true),
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(false),
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithoutEvents(false),
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexSidecarErr(false),
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
				// Set to the size of a partial MsgIndex tx, so that the last event does not fit. We need to add +2
				// since the generated MsgIndex will have NewEthereumBlock set to true at first until the proposer
				// updates it. When NewEthereumBlock is set to true, it consumes 2 bytes, otherwise it does not consume
				// anything.
				MaxTxBytes: int64(utils.TxsSize(encodedMsgIndexPartialBlock(false))) + 2, //nolint:gosec // Safe conversion, result is small

				Txs:    encodedDummyTxs,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Set to zero so that we test functionality without having to consider allocating some block space for
			// Sequencer-native transactions.
			sequencerTxsAllocation: sdkmath.LegacyZeroDec(),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: encodedMsgIndexPartialBlock(false), // partial MsgIndex tx
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg:                    "could not get Ethereum event index offset from state",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - Unrecognized event",
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg:                    "failed to generate MsgIndex and event txs",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - Invalid event field value",
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg:                    "failed to generate MsgIndex and event txs",
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
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
				MaxTxBytes: int64(utils.TxSize(msgSupplyDeltaTx)), //nolint:gosec // Safe conversion, size is small
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,

			expErrMsg: "failed to calculate number of events with max bytes 0",
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
				MaxTxBytes: int64(utils.TxsSize(append(encodedMsgIndexWithEvents(true), msgSupplyDeltaTx))), //nolint:gosec // Safe conversion, size is small
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// We need to set SequencerTxsAllocation to zero so that Sequencer-native transactions are ignored.
			sequencerTxsAllocation: sdkmath.LegacyZeroDec(),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(encodedMsgIndexWithEvents(true), msgSupplyDeltaTx),
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
				MaxTxBytes: int64(totalTxsBytesWithEventsAndSupplyDelta - 1), //nolint:gosec // Safe conversion, size is small
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Set a zero percentage so that we do not prioritize Sequencer-native transactions.
			sequencerTxsAllocation: sdkmath.LegacyZeroDec(),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
				),
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
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas - 1, // Set to total - 1 so that the last transaction is omitted
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
				),
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
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg: "failed to trim event txs from head: insufficient no of events, expected at " +
				"least 4 got 3",
		},
		{
			name:                      "returns error if invalid deposit",
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg:                    "failed to generate MsgIndex and event txs",
		},
		{
			name:                      "skipped event if big authorize found - maxBytes exceeded",
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
			injectedEventTxMaxBytes:      1, // To ensure authorize event is too big
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithSkippedEvent(false),
					s.EncodeMsgSkippedEventTx(
						abci.NewFailedToEncodeEventAsRawTxBytesError(
							sidecartypes.NewGeneratedRawTxBytesExceededMaxBytesError(150, 1),
							testtypes.TestEventsAuthorizeOnly[0],
						).Error(),
						1, // same as height of Ethereum block to query
						testtypes.TestEventsAuthorizeOnly[0].LogIndex,
						testtypes.TestEventsAuthorizeOnly[0].TxIndex,
						testtypes.TestEventsAuthorizeOnly[0].TxHash,
						2, // starting from 1 with msgIndex, then this event
					),
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name:                      "returns error if invalid authorize",
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expErrMsg:                    "failed to generate MsgIndex and event txs",
		},
		{
			name:                      "returns skipped tx if authorize is not authenticated",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseInvalidAuthorizeWithBadAuth, Error: nil,
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
			sequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithSkippedEvent(false),
					s.EncodeMsgSkippedEventTx(
						"unauthorized event",
						1, // same as height
						testtypes.TestEventsAuthorizeWithBadAuth[0].LogIndex,
						testtypes.TestEventsAuthorizeWithBadAuth[0].TxIndex,
						testtypes.TestEventsAuthorizeWithBadAuth[0].TxHash,
						2, // msgIndex, supply delta, then this event
					),
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name:                      "heavy usage - returns all event & sequencer-native txs if right on limits",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: int64(totalTxsBytesWithEventsAndSupplyDelta), //nolint:gosec // Safe conversion, size is small
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion // supply delta height
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Calculate the percentage of block space that dummy txs will take to compute the correct limit.
			sequencerTxsAllocation: sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(utils.TxsSize(
					append(
						[][]byte{},
						encodedDummyTxs[0],
						encodedDummyTxs[1],
						encodedDummyTxs[2],
					),
				), 10),
			).Quo(sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(totalTxsBytesWithEventsAndSupplyDelta, 10),
			)),
			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name: "heavy usage - allocates limited space for sequencer-native txs if too many " +
				"events",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},

			// 4 block events
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseWithFourEvents, Error: nil,
			},

			requestPrepareProposal: &abcitypes.RequestPrepareProposal{

				// We set the maximum size such that we are able to fit in four events to demonstrate that even though
				// we can fit all events, some limited block space is reserve for Sequencer-native transactions.
				MaxTxBytes: int64(totalTxsBytesWithFourEventsAndSupplyDeltaOnly), //nolint:gosec // Safe conversion, size is small

				// We add another transaction to demonstrate that it gets trimmed because there is not enough space.
				Txs: append(encodedDummyTxs, encodedDummyTxs[0]),

				Height: int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion // supply delta height
			},
			maxBlockGas:                  4000, // Enough gas for 4 dummy transactions so that Gas is not a variable
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Calculate the percentage of block space that dummy txs and supply delta will take to compute the correct
			// limit.
			sequencerTxsAllocation: sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(utils.TxsSize(
					append(
						[][]byte{},
						encodedDummyTxs[0],
						encodedDummyTxs[1],
						encodedDummyTxs[2],
					),
				), 10),
			).Quo(sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(totalTxsBytesWithFourEventsAndSupplyDeltaOnly, 10),
			)),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexPartialBlockWith3Events(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name: "block space - sequencerTxsAllocation will get exceeded if there is more space for Sequencer-native" +
				" txs",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: int64(totalTxsBytesWithEventsAndSupplyDelta), //nolint:gosec // Safe conversion, size is small
				Txs:        encodedDummyTxs,
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion // supply delta height
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Allocation is set such that we can fit 2 dummy txs and supply delta. This showcases that the third dummy
			// transaction is still included in the block.
			sequencerTxsAllocation: sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(utils.TxsSize(
					append(
						[][]byte{},
						msgSupplyDeltaTx,
						encodedDummyTxs[0],
						encodedDummyTxs[1],
					),
				), 10),
			).Quo(sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(totalTxsBytesWithEventsAndSupplyDelta, 10),
			)),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithEvents(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
					encodedDummyTxs[1],
					encodedDummyTxs[2],
				),
			},
		},
		{
			name: "block space - if number of Sequencer-native txs is low, event transactions " +
				"are given more space",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},

			// We will return 4 events from the sidecar
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseWithFourEvents, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{

				// MaxTxBytes is set to the size of three events + supply delta + 5 dummy transactions to demonstrate
				// that event number 4 is still included in the block even though there is fixed block space reserved
				// for when the bridge is under heavy usage.
				MaxTxBytes: int64(totalTxsBytesWithEventsSupplyDeltaAndFiveDummyTxs), //nolint:gosec // Safe conversion, size is small
				Txs:        [][]byte{encodedDummyTxs[0]},                             // Just 1 dummy transaction
				Height:     int64(testtypes.TestSupplyDeltaPeriod * 2),               //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion // supply delta height
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,

			// Allocation is set such that we can fit 5 dummy txs and supply delta. This is done to demonstrate the
			// dynamically adjusting block space in favor of event transactions when there is enough space.
			sequencerTxsAllocation: sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(utils.TxsSize(
					append(
						[][]byte{},
						msgSupplyDeltaTx,
						encodedDummyTxs[0],
						encodedDummyTxs[1],
						encodedDummyTxs[2],
						encodedDummyTxs[0],
						encodedDummyTxs[1],
					),
				), 10),
			).Quo(sdkmath.LegacyMustNewDecFromStr(
				strconv.FormatUint(totalTxsBytesWithEventsSupplyDeltaAndFiveDummyTxs, 10),
			)),

			expRes: &abcitypes.ResponsePrepareProposal{
				Txs: append(
					encodedMsgIndexWithFourEvents(true),
					msgSupplyDeltaTx,
					encodedDummyTxs[0],
				),
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

			s.Require().Equal(tc.expRes, res, fmt.Sprintf("{%X}\n!=\n{%X}", tc.expRes, res))
		})
	}
}

func (s *AppTestSuite) TestPrepareProposalHandler_ReqTxsValueWhenErrorOccurs() {
	// In this test we will check the contents of req.Txs when an error occurs in the PrepareProposalHandler. This is
	// important to check out because the baseApp will be using the contents of req.Txs when an error occurs.

	testCases := []struct {
		name                      string
		expQueryBlockEventsCalled int
		expQueryBlockEventsReq    *sidecartypes.QueryBlockEventsRequest
		queryBlockEventsRet       apptesting.MockQueryBlockEventsResponse
		requestPrepareProposal    *abcitypes.RequestPrepareProposal
		expReqTxs                 [][]byte
	}{
		{
			name:                      "req.Txs is empty if default PrepareProposalHandler errors",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestEmptySidecarResponse, Error: nil,
			},
			requestPrepareProposal: &abcitypes.RequestPrepareProposal{
				MaxTxBytes: math.MaxInt64,
				Txs:        [][]byte{[]byte("badly-encoded-tx")},
				Height:     2, // Not expected to be a supply delta height
			},
			expReqTxs: [][]byte{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest()

			// Set bridge module params
			err := s.App.BridgeKeeper.SetParams(
				s.Ctx(),
				bridgetypes.Params{
					SupplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
					EthereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
					InjectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
					SequencerTxsAllocation:       testtypes.TestSequencerTxsAllocation,
				},
			)
			s.Require().NoError(err)

			// Set the MaxBlockGas and MaxTxBytes
			prepareProposalHandlerCtx := s.Ctx().WithConsensusParams(
				comettypes.ConsensusParams{
					Block: &comettypes.BlockParams{
						MaxBytes: tc.requestPrepareProposal.MaxTxBytes,
						MaxGas:   int64(3000),
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
			_, err = propHandler.PrepareProposalHandler()(prepareProposalHandlerCtx, tc.requestPrepareProposal)

			s.Require().Error(err)
			s.Require().Equal(tc.requestPrepareProposal.Txs, tc.expReqTxs)
		})
	}
}

func (s *AppTestSuite) TestProcessProposalHandler() {
	rejectResponse := &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_REJECT}
	acceptResponse := &abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}

	totalTxsGas := int64(3000) // Dummy Txs consume at most 1000 units of gas each. Injected Txs don't consume any gas
	encodedDummyTxs := s.CreateEncodedDummyTxs(3, 1000)

	encodedMsgIndexWithEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndex)
	encodedMsgIndexWithDifferentEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithDifferentEvents)
	encodedMsgIndexPartialBlock := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexPartial)
	encodedMsgIndexWithEventsReduced := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexReduced)
	encodedMsgIndexWithIncorrectAuthority := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexIncorrectAuthority)
	encodedMsgIndexWithNoNewEthBlock := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithNoNewEthBlock)
	encodedMsgIndexWithDiffBlockNumber := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithDiffBlockNumber)
	encodedMsgIndexWithoutEvents := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithoutEvents)
	encodedMsgIndexWithSkippedEvent := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexWithSkippedEvent)
	encodedMsgIndexSidecarErr := s.GetMsgIndexWithEventsEncoder(&testtypes.TestMsgIndexSidecarErr)

	msgIndexSequence := uint64(1)
	msgSupplyDeltaTx := s.EncodeMsgSupplyDeltaTx(msgIndexSequence + 1)

	validTxsWithEvents := append(
		encodedMsgIndexWithEvents(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithEventsWithMissingSupplyDelta := append(
		encodedMsgIndexWithEvents(true), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
		// Note: tx sequences will still be set as if MsgSupplyDelta was there.
	)
	validTxsWithDifferentEvents := append(
		encodedMsgIndexWithDifferentEvents(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsPartialBlock := encodedMsgIndexPartialBlock(false)
	validTxsWithEventsReduced := append(
		encodedMsgIndexWithEventsReduced(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithIncorrectAuthority := append(
		encodedMsgIndexWithIncorrectAuthority(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithNoNewEthBlock := append(
		encodedMsgIndexWithNoNewEthBlock(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithDiffBlockNumber := append(
		encodedMsgIndexWithDiffBlockNumber(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithEventsAndSupplyDelta := append(
		encodedMsgIndexWithEvents(true), msgSupplyDeltaTx, encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithoutEvents := append(
		encodedMsgIndexWithoutEvents(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsSidecarErr := append(
		encodedMsgIndexSidecarErr(false), encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
	)
	validTxsWithUnauthAuthorizeSkippedTx := func(reasonForSkip string) [][]byte {
		return append(
			encodedMsgIndexWithSkippedEvent(false),
			s.EncodeMsgSkippedEventTx(reasonForSkip,
				1,
				0, 0, "",
				2,
			),
			encodedDummyTxs[0], encodedDummyTxs[1], encodedDummyTxs[2],
		)
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
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
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
			expQueryBlockEventsCalled:     0,
			expQueryBlockEventsReq:        nil,
			queryBlockEventsRet:           apptesting.MockQueryBlockEventsResponse{},
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
			expQueryBlockEventsCalled:      1,
			expQueryBlockEventsReq:         &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
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
			name:                      "returns error if MsgIndex cannot be generated - Unrecognized event",
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
			expErrMsg:                    "failed to generate MsgIndex and event txs",
		},
		{
			name:                      "returns error if MsgIndex cannot be generated - Invalid event field value",
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
			expErrMsg:                    "failed to generate MsgIndex and event txs",
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
			name:                          "returns error if events queried from sidecar not equal to block proposer's",
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
			expErrMsg:                    "generated event txs do not match those from the proposal",
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
			name:                      "errors if generated MsgIndex not equal to proposer's (num event txs mismatch)",
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
			expErrMsg: "generated MsgIndex tx differs from that of the block proposal " +
				"(injected: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3310021801200112060A0218011200) " +
				"(generated: 0A610A5F0A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123A0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B331002200112060A0218011200)",
		},
		{
			name:                      "errors if generated MsgIndex not equal to proposer's (authority mismatch)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithIncorrectAuthority,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: "generated MsgIndex tx differs from that of the block proposal " +
				"(injected: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572317738726B326D6B3834777974707878376C6436336B6171706B686D6433396D3035786C67743410031801200112060A0218011200) " +
				"(generated: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3310031801200112060A0218011200)",
		},
		{
			name:                      "errors if generated MsgIndex not equal to proposer's (new eth block mismatch)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithNoNewEthBlock,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: "generated MsgIndex tx differs from that of the block proposal " +
				"(injected: 0A610A5F0A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123A0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B331003200112060A0218011200) " +
				"(generated: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3310031801200112060A0218011200)",
		},
		{
			name:                      "errors if generated MsgIndex not equal to proposer's (block number mismatch)",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponse, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithDiffBlockNumber,
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: "generated MsgIndex tx differs from that of the block proposal " +
				"(injected: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3310031801206412060A0218011200) " +
				"(generated: 0A630A610A212F6675656C73657175656E6365722E6272696467652E76312E4D7367496E646578123C0A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3310031801200112060A0218011200)",
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
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
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
				Height: int64(testtypes.TestSupplyDeltaPeriod * 2), //nolint:gosec // TestSupplyDeltaPeriod is 100, safe conversion
			},
			maxBlockGas:                  totalTxsGas,
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
			expErrMsg: "generated MsgSupplyDelta tx differs from that of the block proposal " +
				"(injected: 0A390A340A1C2F636F736D6F732E62616E6B2E763162657461312E4D736753656E6412140A09746573742D66726F6D1207746573742D746F12013212070A00120310E8071A00) " +
				"(generated: 0A630A610A272F6675656C73657175656E6365722E6272696467652E76312E4D7367537570706C7944656C746112360A346675656C73657175656E636572313064303779323635676D6D757674347A30773961773838306A6E73723730306A646A66766B3312060A0218021200)",
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
			name:                      "accepts block if with skipped tx for unauthenticated authorize event",
			expQueryBlockEventsCalled: 1,
			expQueryBlockEventsReq:    &sidecartypes.QueryBlockEventsRequest{BlockNumber: "1"},
			queryBlockEventsRet: apptesting.MockQueryBlockEventsResponse{
				Response: testtypes.TestSidecarResponseInvalidAuthorizeWithBadAuth, Error: nil,
			},
			requestProcessProposal: &abcitypes.RequestProcessProposal{
				Txs:    validTxsWithUnauthAuthorizeSkippedTx("unauthorized event"),
				Height: 1, // We do not expect MsgSupplyDelta to be injected
			},
			supplyDeltaPeriod:            testtypes.TestSupplyDeltaPeriod,
			ethereumProxyContractAddress: testtypes.TestEthereumProxyContractAddress,
			injectedEventTxMaxBytes:      testtypes.TestInjectedEventTxMaxBytes,
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
			injectedEventTxMaxBytes:      1, // To ensure deposit event is too big
			maxBlockGas:                  totalTxsGas,
			expErrMsg:                    "generated raw tx bytes exceeded max bytes;",
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
					SupplyDeltaPeriod:            tc.supplyDeltaPeriod,
					EthereumProxyContractAddress: tc.ethereumProxyContractAddress,
					MaxEthBlockUpdateDelay:       testMaxEthBlockUpdateDelay,
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
