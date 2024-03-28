package abci

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/gogoproto/proto"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type FuelSequencerProposalHandler struct {
	logger       log.Logger
	valStore     baseapp.ValidatorStore         // to get the current validators' pubkeys
	txSelector   TxSelector                     // a utility for checking whether a tx can be included in the proposal
	txVerifier   baseapp.ProposalTxVerifier     // a utility for transaction verification
	sidecar      sidecarclient.AppSidecarClient // a client to query the Sidecar service
	bridgeKeeper bridgekeeper.Keeper            // Bridge keeper
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	logger log.Logger,
	valStore baseapp.ValidatorStore,
	txVerifier baseapp.ProposalTxVerifier,
	sidecar sidecarclient.AppSidecarClient,
	bridgeKeeper bridgekeeper.Keeper,
) *FuelSequencerProposalHandler {
	return &FuelSequencerProposalHandler{
		logger:       logger,
		valStore:     valStore,
		txVerifier:   txVerifier,
		txSelector:   NewFuelSequencerTxSelector(),
		sidecar:      sidecar,
		bridgeKeeper: bridgeKeeper,
	}
}

// PrepareProposalHandler defines the logic that is executed by the block proposer when they are crafting a new block.
// This function needs to satisfy the following rules:
//
// 1. Can be non-deterministic
// 2. Transaction size cannot exceed RequestPrepareProposal.MaxTxBytes
// 3. Block gas cannot exceed BlockParams.MaxGas
//
// Notes:
//
// 1. Here we are assuming that the NoOp mempool is to be used, meaning that txs requested from CometBFT will simply
// be returned and not verified. It is recommended that the ProcessProposalHandler implements any verifications that
// PrepareProposalHandler implements, therefore, for the NoOp mempool, ProcessProposalHandler shouldn't implement any
// verification checks. In the case that a different mempool is implemented we must perform extra tx verification checks
// in PrepareProposalHandler and ProcessProposalHandler because we are no longer relying on CometBFT, and thus we must
// ensure that we are including valid transactions.
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L199-L303
//
// 2. Any error raised by PrepareProposalHandler is caught in baseapp/abci.go, resulting into the req.Txs to be returned
// to CometBFT.
// Reference: https://github.com/cosmos/cosmos-sdk/blob/a248d05f70f4ad7b8ff7b521e3d23086867d07dc/baseapp/abci.go#L447-L451
func (h *FuelSequencerProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {

		supplyDeltaPeriod := h.bridgeKeeper.GetParams(ctx).SupplyDeltaPeriod
		if supplyDeltaPeriod == 0 {
			return nil, errors.New("SupplyDeltaPeriod cannot be zero")
		}

		// Inject MsgSupplyDeltaTx if expected at current height
		injectMsgSupplyDelta := uint64(req.Height)%supplyDeltaPeriod == 0
		supplyDeltaBytesSize := int64(0)
		if injectMsgSupplyDelta {
			supplyDeltaBytes, err := h.generateMsgSupplyDeltaTx()
			if err != nil {
				return nil, fmt.Errorf("failed to generate msg supply delta tx: %w", err)
			}

			// Calculate size of supply delta message
			supplyDeltaBytesSize = int64(len(supplyDeltaBytes))

			// Set MsgSupplyDeltaTx as first transaction to precede over user initiated MsgSupplyDelta
			req.Txs = append([][]byte{supplyDeltaBytes}, req.Txs...)
		}

		lastEthereumBlockSynced, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		if !found {
			return nil, errors.New("could not get last Ethereum block synced from state")
		}
		ethereumEventIndexOffset, found := h.bridgeKeeper.GetEthereumEventIndexOffset(ctx)
		if !found {
			return nil, errors.New("could not get Ethereum event index offset from state")
		}

		// Query the events of the next Ethereum block
		ethBlockToQuery := lastEthereumBlockSynced.Add(sdkmath.OneInt())
		response, err := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: ethBlockToQuery.String()},
		)

		ethEventsTx, err := h.generateEthEventsTx(response, ethBlockToQuery, err)
		if err != nil {
			return nil, fmt.Errorf("failed to generate eth events tx: %w", err)
		}

		// Trim events from head to skip the events that were already processed.
		err = ethEventsTx.TrimEventsFromHead(ethereumEventIndexOffset.Uint64())
		if err != nil {
			return nil, fmt.Errorf("failed to trim eth events tx head: %w", err)
		}

		// Trim events from tail to fit the block size allocated for events.
		maxBytesForEvents := uint64(req.MaxTxBytes - supplyDeltaBytesSize)
		maxNumberOfEvents := uint64(ethEventsTx.NumberOfEventsWithMaxBytes(maxBytesForEvents))
		trimmed, err := ethEventsTx.KeepEventsFromHead(maxNumberOfEvents)
		if err != nil {
			return nil, fmt.Errorf("failed to trim eth events tx tail: %w", err)
		}
		if trimmed > 0 {
			ctx.Logger().Info(fmt.Sprintf(
				"Skipped %d events because only %d could fit with max bytes %d",
				trimmed, maxNumberOfEvents, maxBytesForEvents,
			))
		}

		ethEventsTxBz, err := ethEventsTx.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to encode injected eth events tx: %w", err)
		}

		// Current implementation assumes txs that cannot be decoded into sdk.Tx are appended at entry 0.
		req.Txs = append([][]byte{ethEventsTxBz}, req.Txs...)

		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		// This is needed to clear variables tracked by the TxSelector
		defer h.txSelector.Clear()

		// EthEventsTx will not satisfy sdk.Tx, so we cannot handle it the same way as we handle other transactions.
		// Note: Here we are assuming that EthEventsTx is injected at index zero and that we will always have such a tx.
		success := h.txSelector.SelectNonSDKTxForProposal(ctx, maxBytesForEvents, req.Txs[0])
		if !success {
			// Given that we trimmed the events list earlier on, we expect the EthEventsTx to be selected successfully.
			// If this is not the case, an empty block is generated and expected to be rejected by ProcessProposal. If a
			// MsgSupplyDelta Tx has been injected in a previous step, this will be disregarded as well.
			//
			// NOTE: When an error is returned the baseApp sends req.Txs to CometBFT, therefore we have to remove any
			// injected EthEventsTx and MsgSupplyDelta Tx to ensure these do not make their way into the block.
			req.Txs = [][]byte{}
			return nil, errors.New("failed to add eth events transaction to block proposal")
		}

		// Since we are assuming a NoOp mempool we simply return the transactions requested from CometBFT, which, by
		// default, should be in FIFO order. Note, we still need to ensure the transactions returned respect
		// req.MaxTxBytes and blockParams.MaxGas
		for index, txBz := range req.Txs[1:] {
			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {

				// We will be assuming that all transactions given to PrepareProposal can be properly decoded.
				// As a result, blocks will get rejected by ProcessProposal if PrepareProposal can't decode a tx.
				return nil, err
			}

			stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)

			// Given that we trimmed the events list earlier on, we expect the MsgSupplyDelta Tx to fit in the block.
			// If this is not the case, there must be something wrong either with the size of EthEventsTx or the
			// MsgSupplyDelta Tx. We want to fail in both of these cases.
			if index == 0 && injectMsgSupplyDelta && len(h.txSelector.SelectedTxs(ctx)) != 2 {
				req.Txs = [][]byte{}
				return nil, errors.New("failed to add message supply delta transaction to block proposal")
			}

			// If we are at full capacity stop adding transactions
			if stop {
				break
			}
		}

		h.logger.Debug("prepared proposal", "txs", len(h.txSelector.SelectedTxs(ctx)))

		return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
	}
}

// ProcessProposalHandler defines the logic which confirms the validity of a block proposal. This logic is executed
// by all validators, and it needs to satisfy the following:
//
// 1. Must be deterministic
// 2. Block gas cannot exceed BlockParams.MaxGas
//
// Notes:
//
// 1. Here we are assuming that the NoOp mempool is to be used, meaning that PrepareProposal may include some txs
// that might fail verification. The logic implemented here attempts to perform the exact verifications performed by the
// PrepareProposalHandler as suggested by the Cosmos SDK documentation. Unfortunately, we cannot verify whether
// transaction sizes exceed the limit because MaxTxBytes is not part of abci.RequestProcessProposal. That being said, it
// seems that CometBFT validates the Tx sizes returned by the application https://github.com/cometbft/cometbft/blob/719b64156aaa3cb89add29d053439060f8e420dd/spec/consensus/creating-proposal.md?plain=1#L41
// It is also important to highlight that the default implementation of ProcessProposalHandler doesn't perform any
// verifications for NoOp mempools, however, we have included a block gas max limit check to follow recommendations
// stated in the documentation.
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L305-L352
//
// 2. Any error raised by ProcessProposalHandler is caught in baseapp/abci.go, resulting in the application to reject
// the block proposal.
// Reference: https://github.com/cosmos/cosmos-sdk/blob/a248d05f70f4ad7b8ff7b521e3d23086867d07dc/baseapp/abci.go#L541-L545
func (h *FuelSequencerProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		// Expect that there is at least one transaction (EthEventsTx must be there)
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"block proposal doesn't have any transactions: first tx expected to be an eth events tx",
			)
		}

		// Expect that the first transaction is always the EthEventsTx
		var injectedEthEventsTx bridgetypes.EthEventsTx
		if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"first transaction expected to be an eth events tx: %w", err,
			)
		}

		lastEthereumBlockSynced, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		if !found {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"could not get last Ethereum block synced from state",
			)
		}
		ethereumEventIndexOffset, found := h.bridgeKeeper.GetEthereumEventIndexOffset(ctx)
		if !found {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"could not get Ethereum event index offset from state",
			)
		}

		// Query the events of the next Ethereum block
		ethBlockToQuery := lastEthereumBlockSynced.Add(sdkmath.OneInt())
		response, err := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: ethBlockToQuery.String()},
		)

		// Generate the EthEventsTx that should be included at index 0 in the block proposal
		ethEventsTx, err := h.generateEthEventsTx(response, ethBlockToQuery, err)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to generate eth events tx: %w", err,
			)
		}

		// Trim events from head to skip the events that were already processed.
		err = ethEventsTx.TrimEventsFromHead(ethereumEventIndexOffset.Uint64())
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim eth events tx head: %w", err,
			)
		}

		// Trim events from tail to fit the block size allocated for events. Unlike the PrepareProposal step, here we do
		// not have access to the max block size, so instead we assume that the proposer proposed an optimised block.
		// TODO: consider adding access to max block size instead of assuming the optimal number of events were proposed
		maxNumberOfEvents := uint64(len(injectedEthEventsTx.Events))
		trimmed, err := ethEventsTx.KeepEventsFromHead(maxNumberOfEvents)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim eth events tx tail: %w", err,
			)
		}
		if trimmed > 0 {
			ctx.Logger().Info(fmt.Sprintf(
				"Skipped %d events because only %d were received from the proposer",
				trimmed, maxNumberOfEvents,
			))
		}

		// Reject block if the sequencer should not proceed with block generation
		if !ethEventsTx.AdvanceSequencer {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"generated eth events tx implies block rejection",
			)
		}

		// Reject block if injected EthEventsTx does not match the one generated by the validator verifying the block
		// proposal
		equal, err := injectedEthEventsTx.Equal(ethEventsTx)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"eth events tx equality check failed: %w", err,
			)
		}
		if !equal {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"generated eth events tx does not match the one in the block proposal",
			)
		}

		supplyDeltaPeriod := h.bridgeKeeper.GetParams(ctx).SupplyDeltaPeriod
		if supplyDeltaPeriod == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"SupplyDeltaPeriod cannot be zero",
			)
		}

		// Check that MsgSupplyDelta was injected correctly if expected
		expectMsgSupplyDelta := uint64(req.Height)%supplyDeltaPeriod == 0
		if expectMsgSupplyDelta {
			err := h.verifyInjectedMsgSupplyDeltaTx(req.Txs)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
					"failed to verify injected MsgSupplyDeltaTx: %w", err,
				)
			}
		}

		var totalTxGas uint64

		var maxBlockGas int64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = b.MaxGas
		}

		// NOTE: Eth events transactions typically don't implement sdk.Tx, therefore, they can be skipped since no
		// gas is consumed. Here we are assuming that eth events txs are always injected at index zero.
		for _, txBytes := range req.Txs[1:] {
			tx, err := h.txVerifier.TxDecode(txBytes)
			if err != nil {

				// This should not occur as PrepareProposal should get transactions that can be decoded properly, but,
				// block proposal rejection is done just in case.
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
			}

			// Confirm that the block's max gas limit is not exceeded
			if maxBlockGas > 0 {
				gasTx, ok := tx.(baseapp.GasTx)
				if ok {
					totalTxGas += gasTx.GetGas()
				}

				if totalTxGas > uint64(maxBlockGas) {
					return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
						"block gas limit exceeded",
					)
				}
			}
		}

		h.logger.Debug("processing proposal", "height", req.Height, "num_txs", len(req.Txs))

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// generateEthEventsTx generates an EthEventsTx based on the response of the sidecar. It returns an error if the events
// returned from the sidecar don't pass validation.
func (h *FuelSequencerProposalHandler) generateEthEventsTx(
	sidecarResponse *sidecartypes.QueryBlockEventsResponse,
	blockNumber sdkmath.Int,
	sidecarErr error,
) (*bridgetypes.EthEventsTx, error) {
	// If sidecar response is nil set the events to nil to avoid null pointer dereference. Context: Sidecar returns nil
	// when it errors.
	var events []*sidecartypes.Event
	if sidecarResponse != nil {
		events = sidecarResponse.Events
	}

	newEthereumBlock := true
	advanceSequencer := true
	if sidecarErr != nil {
		// If the sidecar errored newEthereumBlock must be set to false as this means that the sidecar failed to query
		// the next Ethereum block, therefore, we can't account for it
		newEthereumBlock = false

		// Set advanceSequencer to false if sidecar error is not due to block generation
		if !strings.Contains(sidecarErr.Error(), sidecartypes.ErrBlockDoesNotExist) {
			advanceSequencer = false
		}
	}

	ethEventsTx := bridgetypes.EthEventsTx{
		Events:           events,
		NewEthereumBlock: newEthereumBlock,
		AdvanceSequencer: advanceSequencer,
		BlockNumber:      blockNumber,
	}
	if err := ethEventsTx.ValidateBasic(); err != nil {
		return nil, err
	}

	return &ethEventsTx, nil
}

func (h *FuelSequencerProposalHandler) generateMsgSupplyDeltaTx() ([]byte, error) {

	// Construct Any from message.
	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(&bridgetypes.MsgSupplyDelta{
		Authority: h.bridgeKeeper.GetAuthority(),
	})
	if err != nil {
		return nil, err
	}

	// Construct Tx Body with the message.
	txBodyBz, err := proto.Marshal(&txtypes.TxBody{
		Messages: []*codectypes.Any{msgSupplyDeltaAny},
	})
	if err != nil {
		return nil, err
	}

	// Construct Auth Info with Fee to avoid nil pointer panics.
	authInfoBz, err := proto.Marshal(&txtypes.AuthInfo{
		Fee: &txtypes.Fee{
			GasLimit: MsgSupplyDeltaGasLimit,
		},
	})
	if err != nil {
		return nil, err
	}

	// Construct final Tx.
	txRawBz, err := proto.Marshal(&txtypes.TxRaw{
		BodyBytes:     txBodyBz,
		AuthInfoBytes: authInfoBz,
		Signatures:    nil,
	})
	if err != nil {
		return nil, err
	}

	return txRawBz, nil
}

// verifyInjectedMsgSupplyDeltaTx is used by ProcessProposal to check whether MsgSupplyDeltaTx was injected properly
func (h *FuelSequencerProposalHandler) verifyInjectedMsgSupplyDeltaTx(txs [][]byte) error {

	// We expect at least two transactions when MsgSupplyDeltaTx is injected, the first being EthEventsTx
	if len(txs) < 2 {
		return errors.New("expected at least two transactions in block proposal")
	}

	// We expect MsgSupplyDeltaTx to be injected at index 1, therefore, try to decode transaction at index 1 to sdk.Tx.
	tx, err := h.txVerifier.TxDecode(txs[1])
	if err != nil {
		return fmt.Errorf("failed to decode transaction at index 1 into sdk.Tx: %w", err)
	}

	// MsgSupplyDeltaTx should contain exactly one message, MsgSupplyDelta
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return errors.New("expected one message in transaction at index 1")
	}
	msg := msgs[0]

	// Confirm that the proper message was encoded
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}) {
		return fmt.Errorf(
			"incorrect msg type url in transaction at index 1; expected %s got %s",
			sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}),
			sdk.MsgTypeURL(msg),
		)
	}

	// Check that the message unmarshals successfully to MsgSupplyDelta
	msgSupplyDelta, ok := msg.(*bridgetypes.MsgSupplyDelta)
	if !ok {
		return errors.New("could not unmarshal message in transaction at index 1 to MsgSupplyDelta")
	}

	// Confirm that MsgSupplyDelta passes all verification checks and error if not
	if msgSupplyDelta.Authority != h.bridgeKeeper.GetAuthority() {
		return fmt.Errorf(
			"incorrect Authority set in MsgSupplyDelta; expected %s got %s",
			h.bridgeKeeper.GetAuthority(),
			msgSupplyDelta.Authority,
		)
	}

	return nil
}

// PreBlocker contains logic that should run before any FinalizeBlock logic. FinalizeBlock ignores any byte slices not
// implementing sdk.Tx. As a consequence, any important results originating from PrepareProposal or ProcessProposal not
// implementing sdk.Tx need to be made available to the modules in storage at PreBlocker stage.
func (h *FuelSequencerProposalHandler) PreBlocker(
	ctx sdk.Context, req *abci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	// This check is done for completeness’s sake as we should not expect to run into this scenario
	if len(req.Txs) == 0 {
		return nil, fmt.Errorf("expected eth events transaction to be injected")
	}

	var injectedEthEventsTx bridgetypes.EthEventsTx
	if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
		return nil, fmt.Errorf("failed to decode injected eth events tx: %w", err)
	}

	// Perform some checks on the injected Ethereum events transaction.
	// If any problem is found, this is an indication of a serious bug.
	lastBlockSynced := h.bridgeKeeper.MustGetLastEthereumBlockSynced(ctx)
	eventIndexOffset := h.bridgeKeeper.MustGetEthereumEventIndexOffset(ctx)
	err := injectedEthEventsTx.ValidateBeforeProcessing(lastBlockSynced, eventIndexOffset)
	if err != nil {
		return nil, fmt.Errorf("eth events tx validation failed: %w", err)
	}

	// Set the injected events into state if any.
	if len(injectedEthEventsTx.Events) > 0 {
		h.bridgeKeeper.SetEthEventsTx(ctx, injectedEthEventsTx)
	}

	// Set LastEthereumBlockSynced and reset EthereumEventIndexOffset if we are to increment to a new Ethereum block.
	if injectedEthEventsTx.NewEthereumBlock {
		h.bridgeKeeper.SetLastEthereumBlockSynced(ctx, injectedEthEventsTx.BlockNumber)
		h.bridgeKeeper.ResetEthereumEventIndexOffset(ctx)
	}

	// If no new Ethereum block, but we still received some events, then the block was partially consumed.
	if !injectedEthEventsTx.NewEthereumBlock && len(injectedEthEventsTx.Events) > 0 {
		newOffset := eventIndexOffset.AddRaw(int64(len(injectedEthEventsTx.Events)))
		h.bridgeKeeper.SetEthereumEventIndexOffset(ctx, newOffset)
	}

	h.logger.Debug("finished executing pre-block hook")

	return &sdk.ResponsePreBlock{}, nil
}
