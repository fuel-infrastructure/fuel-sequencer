package abci

import (
	"errors"
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type FuelSequencerProposalHandler struct {
	cdc          codec.Codec // codec
	logger       log.Logger
	valStore     baseapp.ValidatorStore         // to get the current validators' pubkeys
	txSelector   baseapp.TxSelector             // a utility for checking whether a tx can be included in the proposal
	txVerifier   baseapp.ProposalTxVerifier     // a utility for transaction verification
	sidecar      sidecarclient.AppSidecarClient // a client to query the Sidecar service
	bridgeKeeper bridgekeeper.Keeper            // Bridge keeper
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	cdc codec.Codec,
	logger log.Logger,
	valStore baseapp.ValidatorStore,
	txVerifier baseapp.ProposalTxVerifier,
	sidecar sidecarclient.AppSidecarClient,
	bridgeKeeper bridgekeeper.Keeper,
) *FuelSequencerProposalHandler {
	return &FuelSequencerProposalHandler{
		cdc:          cdc,
		logger:       logger,
		valStore:     valStore,
		txVerifier:   txVerifier,
		txSelector:   baseapp.NewDefaultTxSelector(),
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
		bridgeParams := h.bridgeKeeper.GetParams(ctx)

		supplyDeltaPeriod := bridgeParams.SupplyDeltaPeriod
		if supplyDeltaPeriod == 0 {
			return nil, errors.New("SupplyDeltaPeriod cannot be zero")
		}

		blockedAddresses, err := h.bridgeKeeper.GetAllBlockedAddresses(ctx, bridgeParams.AdditionalBlockedAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to get blocked addresses: %w", err)
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

		// Query the events of the next Ethereum block.
		ethBlockToQuery := lastEthereumBlockSynced + 1
		response, sidecarErr := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: strconv.FormatUint(ethBlockToQuery, 10)},
		)
		if sidecarErr != nil {
			ctx.Logger().Warn("observed sidecar error at PrepareProposal", "err", sidecarErr)
			// This error is also passed to generateMsgIndexAndEventTxs to perform dedicated error handling.
		}

		msgIndex, eventTxs, err := h.generateMsgIndexAndEventTxs(
			response, ethBlockToQuery, sidecarErr, &bridgeParams, blockedAddresses,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MsgIndex: %w", err)
		}

		// Trim events from head to skip the events that were already processed.
		eventTxs, err = msgIndex.TrimEventsFromHead(eventTxs, ethereumEventIndexOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to trim events from head: %w", err)
		}

		// Sanity check: number of event txs is equal to NumInjectedTxs
		if msgIndex.NumInjectedTxs != uint64(len(eventTxs)) {
			return nil, fmt.Errorf(
				"mismatch in number of events; expected: %d, got: %d",
				len(eventTxs), msgIndex.NumInjectedTxs,
			)
		}

		// Trim events from tail to fit the block size allocated for events.
		maxBytesForEvents := uint64(req.MaxTxBytes - supplyDeltaBytesSize)
		maxNumberOfEvents, err := msgIndex.NumberOfEventsWithMaxBytes(eventTxs, maxBytesForEvents)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate number of events for max bytes %d: %w", maxBytesForEvents, err)
		}
		originalNumberOfEvents := len(eventTxs)
		eventTxs, trimmed, err := msgIndex.KeepEventsFromHead(eventTxs, uint64(maxNumberOfEvents))
		if err != nil {
			return nil, fmt.Errorf("failed to trim events from tail: %w", err)
		}
		if trimmed > 0 {
			ctx.Logger().Debug(fmt.Sprintf(
				"Skipped %d/%d of remaining events from block %d because only %d could fit in max bytes %d",
				trimmed, originalNumberOfEvents, msgIndex.BlockNumber, maxNumberOfEvents, maxBytesForEvents,
			))
		}

		msgIndexBz, err := msgIndex.RawTxBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to encode MsgIndex: %w", err)
		}

		// Inject MsgIndex and Ethereum event transactions as the first txs in the block.
		req.Txs = append(append([][]byte{msgIndexBz}, eventTxs...), req.Txs...)

		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		// This is needed to clear variables tracked by the TxSelector
		defer h.txSelector.Clear()

		// Since we are assuming a NoOp mempool we simply return the transactions requested from CometBFT, which, by
		// default, should be in FIFO order. Note, we still need to ensure the transactions returned respect
		// req.MaxTxBytes and blockParams.MaxGas
		for _, txBz := range req.Txs {
			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {

				// We will be assuming that all transactions given to PrepareProposal can be properly decoded.
				// As a result, blocks will get rejected by ProcessProposal if PrepareProposal can't decode a tx.
				return nil, err
			}

			stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)

			// If we are at full capacity stop adding transactions
			if stop {
				break
			}
		}

		// Calculate the minimum number of expected transactions, which is the MsgIndex, the number of event txs, and
		// lastly the MsgSupplyDelta, if we're at the MsgSupplyDelta height.
		minimumExpectedTxs := 1 + msgIndex.NumInjectedTxs
		if injectMsgSupplyDelta {
			minimumExpectedTxs += 1
		}

		// Given that we trimmed the events list earlier on, we expect the MsgSupplyDelta transaction and the event
		// transactions to fit in the block. If this is not the case, there must be something wrong with the size
		// calculations or trimming logic. We want to fail in both of these cases.
		if uint64(len(h.txSelector.SelectedTxs(ctx))) < minimumExpectedTxs {
			req.Txs = [][]byte{}
			return nil, errors.New("failed to add all mandatory messages to block proposal")
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
		// Expect that there is at least one transaction (MsgIndex must be there)
		if len(req.Txs) == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"block proposal doesn't have any transactions: first tx expected to be MsgIndex",
			)
		}

		bridgeParams := h.bridgeKeeper.GetParams(ctx)

		blockedAddresses, err := h.bridgeKeeper.GetAllBlockedAddresses(ctx, bridgeParams.AdditionalBlockedAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to get blocked addresses: %w", err)
		}

		// Expect that the first transaction is always a valid transaction containing MsgIndex.
		injectedMsgIndexUnparsed, err := h.txVerifier.TxDecode(req.Txs[0])
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"first transaction expected to be an valid tx: %w", err,
			)
		}

		// Parse the first transaction into an MsgIndex.
		var injectedMsgIndex bridgetypes.MsgIndex
		err = injectedMsgIndex.FromSdkTx(injectedMsgIndexUnparsed)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"first transaction expected to be MsgIndex: %w", err,
			)
		}

		// Reject the block if it doesn't indicate a sync-up with Ethereum and if we haven't synced up with Ethereum
		// for a while.
		lastEthBlockUpdateTime, found := h.bridgeKeeper.GetLastEthBlockUpdateTime(ctx)
		ethSyncDelayExceeded := found && req.Time.After(lastEthBlockUpdateTime.Add(bridgeParams.MaxEthBlockUpdateDelay))
		if !injectedMsgIndex.NewEthereumBlock && ethSyncDelayExceeded {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"last syncup with Ethereum was at %s; block time: %s; max delay allowed: %s",
				lastEthBlockUpdateTime.String(),
				req.Time.String(),
				bridgeParams.MaxEthBlockUpdateDelay.String(),
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

		// Query the events of the next Ethereum block.
		ethBlockToQuery := lastEthereumBlockSynced + 1
		response, sidecarErr := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: strconv.FormatUint(ethBlockToQuery, 10)},
		)
		if sidecarErr != nil {
			ctx.Logger().Warn("observed sidecar error at ProcessProposal", "err", sidecarErr)
			// This error is also passed to generateMsgIndexAndEventTxs to perform dedicated error handling.
		}

		// Generate the MsgIndex that should be included at index 0 in the block proposal
		msgIndex, eventTxs, err := h.generateMsgIndexAndEventTxs(
			response, ethBlockToQuery, sidecarErr, &bridgeParams, blockedAddresses,
		)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to generate MsgIndex: %w", err,
			)
		}

		// Trim events from head to skip the events that were already processed.
		eventTxs, err = msgIndex.TrimEventsFromHead(eventTxs, ethereumEventIndexOffset)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim events from head: %w", err,
			)
		}

		// Trim events from tail to fit the block size allocated for events. Unlike the PrepareProposal step, here we do
		// not have access to the max block size, so instead we assume that the proposer proposed an optimised block.
		// TODO: consider adding access to max block size instead of assuming the optimal number of events were proposed
		originalNumberOfEvents := msgIndex.NumInjectedTxs
		eventTxs, trimmed, err := msgIndex.KeepEventsFromHead(eventTxs, injectedMsgIndex.NumInjectedTxs)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim events from tail: %w", err,
			)
		}
		if trimmed > 0 {
			ctx.Logger().Debug(fmt.Sprintf(
				"Skipped %d/%d of remaining events from block %d because only %d were received from the proposer",
				trimmed, originalNumberOfEvents, msgIndex.BlockNumber, injectedMsgIndex.NumInjectedTxs,
			))
		}

		// Extract the injected event txs and ensure that we've gotten the right amount of transactions.
		injectedEventTxs := req.Txs[1 : injectedMsgIndex.NumInjectedTxs+1]
		if uint64(len(injectedEventTxs)) != injectedMsgIndex.NumInjectedTxs {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"unexpected number of injected event txs; expected: %d, got: %d",
				injectedMsgIndex.NumInjectedTxs, len(injectedEventTxs),
			)
		}

		// Reject block if injected MsgIndex does not match the one generated by the validator verifying the block
		// proposal
		err = injectedMsgIndex.Equal(msgIndex, injectedEventTxs, eventTxs)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"generated data does not match that from the block proposal: %w", err,
			)
		}

		supplyDeltaPeriod := bridgeParams.SupplyDeltaPeriod
		if supplyDeltaPeriod == 0 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"SupplyDeltaPeriod cannot be zero",
			)
		}

		// Check that MsgSupplyDelta was injected correctly if expected
		expectMsgSupplyDelta := uint64(req.Height)%supplyDeltaPeriod == 0
		if expectMsgSupplyDelta {
			err := h.verifyInjectedMsgSupplyDeltaTx(req.Txs, msgIndex.NumInjectedTxs)
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

		for _, txBytes := range req.Txs {
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

// getNewEthereumBlock returns the value for MsgIndex.NewEthereumBlock. NewEthereumBlock should be true iff the
// Sidecar didn't error.
func (h *FuelSequencerProposalHandler) getNewEthereumBlock(sidecarErr error) bool {
	return sidecarErr == nil
}

// generateMsgIndexAndEventTxs generates MsgIndex and transactions from events based on the response of the sidecar.
// It errors upon invalid events from the sidecar. Returned errors have the capability of halting block production.
func (h *FuelSequencerProposalHandler) generateMsgIndexAndEventTxs(
	sidecarResponse *sidecartypes.QueryBlockEventsResponse,
	blockNumber uint64,
	sidecarErr error,
	params *bridgetypes.Params,
	blockedAddresses map[string]bool,
) (msgIndex *bridgetypes.MsgIndex, eventTxs [][]byte, err error) {

	// Set events to nil by default to avoid a null pointer dereference if the Sidecar errors.
	// Context: Sidecar returns a nil response when it errors.
	var events []*sidecartypes.Event
	if sidecarResponse != nil {
		events = sidecarResponse.Events
	}

	// Identify the deposit and authorize events in the MsgIndex and produce one new valid transaction per event.
	// If an event is not valid, returned errors have the capability of halting block production.
	for _, event := range events {

		err = event.Validate(params.EthereumProxyContractAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("encountered invalid event with err: %s; event: %s", err.Error(), event)
		}

		eventTx, err := event.RawTxBytes(h.cdc, h.bridgeKeeper.GetAuthority())
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get messages with err: %s; event: %s", err.Error(), event)
		}

		authenticated, err := h.authenticateEvent(event, eventTx, params, blockedAddresses)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to check authorization: %s; event: %s", err.Error(), event)
		}

		if authenticated {
			eventTxs = append(eventTxs, eventTx)
		} else {
			h.logger.Warn(fmt.Sprintf("skipping unauthorized event: %s", event))
		}
	}

	// Generate MsgIndex based on the number of injected events.
	msgIndex = &bridgetypes.MsgIndex{
		Authority:        h.bridgeKeeper.GetAuthority(),
		NumInjectedTxs:   uint64(len(eventTxs)),
		NewEthereumBlock: h.getNewEthereumBlock(sidecarErr),
		BlockNumber:      blockNumber,
	}

	return
}

func (h *FuelSequencerProposalHandler) generateMsgSupplyDeltaTx() ([]byte, error) {

	// Construct Any from message.
	msgSupplyDeltaAny, err := codectypes.NewAnyWithValue(&bridgetypes.MsgSupplyDelta{
		Authority: h.bridgeKeeper.GetAuthority(),
	})
	if err != nil {
		return nil, err
	}

	return utils.ValidRawTxBytesFromAnyMsgs([]*codectypes.Any{msgSupplyDeltaAny})
}

// verifyInjectedMsgSupplyDeltaTx is used by ProcessProposal to check whether MsgSupplyDeltaTx was injected properly
func (h *FuelSequencerProposalHandler) verifyInjectedMsgSupplyDeltaTx(txs [][]byte, numInjectedTxs uint64) error {

	// We expect the MsgSupplyDeltaTx to be at the index right after MsgIndex and any injected events.
	// If no events were injected, the index will be 1, i.e. right after the MsgIndex.
	msgSupplyDeltaIndex := 1 + numInjectedTxs

	// Check that there's enough transactions for the above index to make sense.
	if uint64(len(txs)) < 1+msgSupplyDeltaIndex {
		return fmt.Errorf("expected at least %d transactions in block proposal", 1+msgSupplyDeltaIndex)
	}

	// Try to decode transaction at the MsgSupplyDelta index into sdk.Tx.
	tx, err := h.txVerifier.TxDecode(txs[msgSupplyDeltaIndex])
	if err != nil {
		return fmt.Errorf("failed to decode transaction at index %d into sdk.Tx: %w", msgSupplyDeltaIndex, err)
	}

	// MsgSupplyDeltaTx should contain exactly one message, MsgSupplyDelta
	msgs := tx.GetMsgs()
	if len(msgs) != 1 {
		return fmt.Errorf("expected one message in transaction at index %d", msgSupplyDeltaIndex)
	}
	msg := msgs[0]

	// Confirm that the proper message was encoded
	if sdk.MsgTypeURL(msg) != sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}) {
		return fmt.Errorf(
			"incorrect msg type url in transaction at index %d; expected %s got %s",
			msgSupplyDeltaIndex,
			sdk.MsgTypeURL(&bridgetypes.MsgSupplyDelta{}),
			sdk.MsgTypeURL(msg),
		)
	}

	// Check that the message unmarshals successfully to MsgSupplyDelta
	msgSupplyDelta, ok := msg.(*bridgetypes.MsgSupplyDelta)
	if !ok {
		return fmt.Errorf(
			"could not unmarshal message in transaction at index %d to MsgSupplyDelta",
			msgSupplyDeltaIndex,
		)
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
