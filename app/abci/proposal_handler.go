package abci

import (
	"errors"
	"fmt"
	"strconv"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
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
	cdc                    codec.Codec // codec
	logger                 log.Logger
	valStore               baseapp.ValidatorStore          // to get the current validators' pubkeys
	txVerifier             baseapp.ProposalTxVerifier      // a utility for transaction verification
	sidecar                sidecarclient.AppSidecarClient  // a client to query the Sidecar service
	bridgeKeeper           bridgekeeper.Keeper             // Bridge keeper
	defaultProposalHandler *baseapp.DefaultProposalHandler // gives us access to default proposal handling behaviour
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
		cdc:                    cdc,
		logger:                 logger,
		valStore:               valStore,
		txVerifier:             txVerifier,
		sidecar:                sidecar,
		bridgeKeeper:           bridgeKeeper,
		defaultProposalHandler: baseapp.NewDefaultProposalHandler(nil, txVerifier),
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

		blockedAddresses, err := h.bridgeKeeper.GetAllBlockedAddresses(ctx, bridgeParams.AdditionalBlockedAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to get blocked addresses: %w", err)
		}

		// Inject MsgSupplyDeltaTx if expected at current height
		injectMsgSupplyDelta := bridgeParams.IsMsgSupplyDeltaBlock(uint64(req.Height))
		supplyDeltaBytesSize := int64(0)
		if injectMsgSupplyDelta {
			supplyDeltaBytes, err := h.generateMsgSupplyDeltaTx()
			if err != nil {
				return nil, fmt.Errorf("failed to generate msg supply delta tx: %w", err)
			}

			// Calculate size of supply delta message
			supplyDeltaBytesSize = int64(len(supplyDeltaBytes))

			// Ensure that supply delta fits in the block on its own
			if supplyDeltaBytesSize > req.MaxTxBytes {
				return nil, fmt.Errorf(
					"could not fit MsgSupplyDelta of size %d in block's max bytes %d",
					supplyDeltaBytesSize, req.MaxTxBytes,
				)
			}

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
			return nil, fmt.Errorf("failed to generate MsgIndex and event txs: %w", err)
		}

		// Trim events from head to skip the events that were already processed.
		eventTxs, err = msgIndex.TrimEventsFromHead(eventTxs, ethereumEventIndexOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to trim event txs from head: %w", err)
		}

		/**
		Calculate the block space that should be reserved for event transactions.
		*/

		// The SupplyDelta tx size is considered in the calculations as we need to make sure that it is always injected
		// when required. Note, SupplyDelta transactions are not considered to be Sequencer-native txs as these are
		// injected by the consensus algorithm.
		sequencerTxsSize := utils.NumberOfBytes(req.Txs) - uint64(supplyDeltaBytesSize)
		maxBlockSpace := uint64(req.MaxTxBytes) - uint64(supplyDeltaBytesSize)

		// Reserve a percentage of the available block space for Sequencer-native transactions. We are sure that this
		// will not cover the entire block space because there are limits imposed on
		// bridgeParams.SequencerTxsAllocation.
		// Note: We might allocate less block space than bridgeParams.SequencerTxsAllocation indicates when supply delta
		// transactions are injected, because, we will take a percentage of req.MaxTxBytes - supplyDeltaBytesSize. This
		// should be fine as long as SupplyDelta tx remains small and the MaxBytes consensus parameter is set to a
		// reasonable value.
		sequencerTxsBlockSpace := uint64(bridgeParams.SequencerTxsAllocation.MulInt(
			sdkmath.NewIntFromUint64(maxBlockSpace),
		).TruncateInt64())

		var maxBytesForEvents uint64
		if sequencerTxsSize > maxBlockSpace {

			// If size of Sequencer-native transactions given by CometBFT is bigger than the available block space,
			// allocate bridgeParams.SequencerTxsAllocation percent of the available block space to Sequencer-native
			// transactions and the rest to event transactions. Note, we need to make this check because if
			// sequencerTxsSize > maxBlockSpace we will run into overflow issues when subtracting two uint64 values.
			maxBytesForEvents = maxBlockSpace - sequencerTxsBlockSpace
		} else {

			// Otherwise, Sequencer-native transactions will be set to occupy at most
			// bridgeParams.SequencerTxsAllocation percent of the available block space, depending on the size of
			// Sequencer-native transactions.
			maxBytesForEvents = max(maxBlockSpace-sequencerTxsSize, maxBlockSpace-sequencerTxsBlockSpace)
		}

		// Trim events from tail to fit the block space allocated for events.
		// NOTE: The TxSelector will be able to fit in more Sequencer-native transactions at the end if there is more
		// space in the block after adjusting the number of event transactions.
		maxNumberOfEvents, err := msgIndex.NumberOfEventsWithMaxBytes(eventTxs, maxBytesForEvents)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate number of events with max bytes %d: %w", maxBytesForEvents, err)
		}
		originalNumberOfEvents := len(eventTxs)
		eventTxs, trimmed, err := msgIndex.KeepEventsFromHead(eventTxs, uint64(maxNumberOfEvents))
		if err != nil {
			return nil, fmt.Errorf("failed to trim event txs from tail: %w", err)
		}
		if trimmed > 0 {
			ctx.Logger().Debug(fmt.Sprintf(
				"Skipped %d/%d of remaining events from block %d because only %d could fit in max bytes %d",
				trimmed, originalNumberOfEvents, msgIndex.BlockNumber, maxNumberOfEvents, maxBytesForEvents,
			))
		}

		// Sanity check: number of event txs is equal to NumInjectedEventTxs
		if msgIndex.NumInjectedEventTxs != uint64(len(eventTxs)) {
			return nil, fmt.Errorf(
				"mismatch between NumInjectedEventTxs in index and actual number of event txs; expected: %d, got: %d",
				msgIndex.NumInjectedEventTxs, len(eventTxs),
			)
		}

		// ----- Beyond this point, any error returned should consider setting req.Txs = [][]byte{},
		// otherwise CometBFT will still use the req.Txs even though we return an error or panic.
		// Anything that comes before this point will cause ProcessProposal to error where a
		// MsgIndex tx is expected

		// Inject MsgIndex and Ethereum event transactions as the first txs in the block.
		msgIndexBz, err := msgIndex.RawTxBytes()
		if err != nil {
			return nil, fmt.Errorf("failed to encode MsgIndex: %w", err)
		}
		req.Txs = append(append([][]byte{msgIndexBz}, eventTxs...), req.Txs...)

		// Use the DefaultProposalHandler which, since we're using a NoOp mempool, will select the txs requested from
		// CometBFT which by default should be in FIFO order. It still ensures the txs returned respect req.MaxTxBytes
		// and blockParams.MaxGas. Amongst these transactions are a number of injected txs which will consume zero gas.
		resp, err := h.defaultProposalHandler.PrepareProposalHandler()(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("default proposal handler failed with error: %w", err)
		}
		selectedTxs := resp.Txs

		// Check that we've collected the minimum expected transactions
		err = checkMinimumNumTxs(uint64(len(selectedTxs)), msgIndex, injectMsgSupplyDelta)
		if err != nil {
			req.Txs = [][]byte{}
			return nil, err
		}

		h.logger.Debug("prepared proposal", "txs", len(selectedTxs))

		return &abci.ResponsePrepareProposal{Txs: selectedTxs}, nil
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
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to get blocked addresses: %w", err,
			)
		}

		// Parse the first transaction into an MsgIndex.
		var injectedMsgIndex bridgetypes.MsgIndex
		err = injectedMsgIndex.FromRawTxBytes(req.Txs[0], h.txVerifier.TxDecode)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"first transaction expected to be a valid MsgIndex: %w", err,
			)
		}

		// Reject the block if it doesn't indicate a sync-up with Ethereum and if we haven't synced up with Ethereum
		// for a while.
		if injectedMsgIndex.NoEthereumSyncing() {
			// No Ethereum syncing, therefore, we only accept this block if MaxEthBlockUpdateDelay is not exceeded
			lastEthBlockUpdateTime, found := h.bridgeKeeper.GetLastEthBlockUpdateTime(ctx)
			ethSyncDelayExceeded := found && req.Time.After(lastEthBlockUpdateTime.Add(bridgeParams.MaxEthBlockUpdateDelay))

			if ethSyncDelayExceeded {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
					"last syncup with Ethereum was at %s; block time: %s; max delay allowed: %s",
					lastEthBlockUpdateTime.String(),
					req.Time.String(),
					bridgeParams.MaxEthBlockUpdateDelay.String(),
				)
			}
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
				"failed to generate MsgIndex and event txs: %w", err,
			)
		}

		// Trim events from head to skip the events that were already processed.
		eventTxs, err = msgIndex.TrimEventsFromHead(eventTxs, ethereumEventIndexOffset)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim event txs from head: %w", err,
			)
		}

		// Trim events from tail to fit the block size allocated for events. Unlike the PrepareProposal step, here we do
		// not have access to the max block size, so instead we assume that the proposer proposed an optimised block.
		// TODO: consider adding access to max block size instead of assuming the optimal number of events were proposed
		originalNumberOfEvents := msgIndex.NumInjectedEventTxs
		eventTxs, trimmed, err := msgIndex.KeepEventsFromHead(eventTxs, injectedMsgIndex.NumInjectedEventTxs)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to trim event txs from tail: %w", err,
			)
		}
		if trimmed > 0 {
			ctx.Logger().Debug(fmt.Sprintf(
				"Skipped %d/%d of remaining events from block %d because only %d were received from the proposer",
				trimmed, originalNumberOfEvents, msgIndex.BlockNumber, injectedMsgIndex.NumInjectedEventTxs,
			))
		}

		// Sanity check: number of event txs is equal to NumInjectedEventTxs
		if msgIndex.NumInjectedEventTxs != uint64(len(eventTxs)) {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"mismatch between NumInjectedEventTxs in index and actual number of event txs; expected: %d, got: %d",
				msgIndex.NumInjectedEventTxs, len(eventTxs),
			)
		}

		// Extract the injected event txs and ensure that we've gotten the right amount of transactions.
		injectedEventTxs := req.Txs[1 : injectedMsgIndex.NumInjectedEventTxs+1]
		if uint64(len(injectedEventTxs)) != injectedMsgIndex.NumInjectedEventTxs {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"unexpected number of injected event txs; expected: %d, got: %d",
				injectedMsgIndex.NumInjectedEventTxs, len(injectedEventTxs),
			)
		}

		// Reject block if injected MsgIndex does not match the one generated by the validator verifying the block
		// proposal
		err = injectedMsgIndex.Equal(msgIndex, injectedEventTxs, eventTxs)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"generated injected txs do not match the ones from the block proposal: %w", err,
			)
		}

		// Check that MsgSupplyDelta was injected correctly if expected
		expectMsgSupplyDelta := bridgeParams.IsMsgSupplyDeltaBlock(uint64(req.Height))
		if expectMsgSupplyDelta {
			err := h.verifyInjectedMsgSupplyDeltaTx(req.Txs, msgIndex.NumInjectedEventTxs)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
					"failed to verify injected MsgSupplyDeltaTx: %w", err,
				)
			}
		}

		// Verify transactions' bytes and gas consumption.
		err = verifyTransactionsInProposal(ctx, req.Txs, h.txVerifier.TxDecode)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		// Check that we've collected the minimum expected transactions
		err = checkMinimumNumTxs(uint64(len(req.Txs)), msgIndex, expectMsgSupplyDelta)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, err
		}

		h.logger.Debug("processing proposal", "height", req.Height, "num_txs", len(req.Txs))

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// checkMinimumNumTxs ensures that we've collected the minimum number of expected transactions, i.e. MsgIndex, the event
// transactions, and the MsgSupplyDelta if we're at the MsgSupplyDelta height, and returns an error otherwise
func checkMinimumNumTxs(numTxs uint64, msgIndex *bridgetypes.MsgIndex, injectMsgSupplyDelta bool) error {

	minimumExpectedTxs := 1 + msgIndex.NumInjectedEventTxs
	if injectMsgSupplyDelta {
		minimumExpectedTxs += 1
	}

	if numTxs < minimumExpectedTxs {
		return fmt.Errorf(
			"incorrect number of messages in block proposal; expected minimum number of txs %d, got %d",
			minimumExpectedTxs, numTxs,
		)
	}

	return nil
}

// verifyTransactionsInProposal checks that all the transactions proposed by the block proposer can be decoded into a
// valid SDK transaction, and that the total gas consumption by the transactions will be less than the max gas.
func verifyTransactionsInProposal(ctx sdk.Context, txs [][]byte, txDecoder sdk.TxDecoder) error {
	var totalTxGas uint64

	var maxBlockGas int64
	if b := ctx.ConsensusParams().Block; b != nil {
		maxBlockGas = b.MaxGas
	}

	// Note: amongst these transactions are a number of injected txs which will consume zero gas.
	for _, txBytes := range txs {
		tx, err := txDecoder(txBytes)
		if err != nil {

			// This should not occur as PrepareProposal should get transactions that can be decoded properly, but,
			// an error is returned just in case.
			return err
		}

		// Confirm that the block's max gas limit is not exceeded
		if maxBlockGas > 0 {
			gasTx, ok := tx.(baseapp.GasTx)
			if ok {
				totalTxGas += gasTx.GetGas()
			}

			if totalTxGas > uint64(maxBlockGas) {
				return errors.New(
					"block gas limit exceeded",
				)
			}
		}
	}

	return nil
}

// getNewEthereumBlock returns the value for MsgIndex.NewEthereumBlock. NewEthereumBlock should be true if the
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
	// If an Authorize event cannot be encoded as bytes tx, is bigger than the allowed max bytes, has more messages than
	// the allowable limit, or is not authenticated we will skip it. This is done to protect the Sequencer from attacks
	// induced from a very large or invalid payload. On the other hand, deposit events are expected to always have the
	// same structure as they are fully generated by the smart contracts. Due to this, we will halt the block production
	// if a deposit event fails to be encoded as a bytes tx or is bigger than the allowed max bytes.
	for _, event := range events {

		err = event.Validate(params.EthereumProxyContractAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("encountered invalid event with err: %s; event: %s", err.Error(), event)
		}

		eventTx, err := event.RawTxBytesWithLimitChecks(
			h.cdc, h.bridgeKeeper.GetAuthority(), params.InjectedEventTxMaxBytes, params.MaxAuthorizeMessages,
		)
		if err != nil {

			// If an event is an authorization it should be skipped.
			if event.EventType == sidecartypes.AuthorizeEventName {
				h.logger.Warn(fmt.Sprintf(
					"skipping event; failed to encode event as raw tx bytes with err: %s; event: %s",
					err.Error(),
					event,
				))
				continue
			}

			// Otherwise, return an error as it means that we have an event that should always be expected to be encoded
			// successfully as a bytes tx. This will halt block production.
			return nil, nil, fmt.Errorf(
				"failed to encode event as raw tx bytes with err: %s; event: %s",
				err.Error(), event,
			)
		}

		authenticated, err := h.authenticateEvent(event, eventTx, params, blockedAddresses)
		if err != nil {

			// At this stage it is safe to assume that garbage payloads sent in Authorize events by users would have
			// already been skipped. Therefore, we can return an error here and halt block production.
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
		Authority:           h.bridgeKeeper.GetAuthority(),
		NumInjectedEventTxs: uint64(len(eventTxs)),
		NewEthereumBlock:    h.getNewEthereumBlock(sidecarErr),
		BlockNumber:         blockNumber,
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

	var msgSupplyDelta bridgetypes.MsgSupplyDelta
	err = msgSupplyDelta.FromSdkTx(tx)
	if err != nil {
		return fmt.Errorf("failed to parse MsgSupplyDelta at index %d with error: %w", msgSupplyDeltaIndex, err)
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
