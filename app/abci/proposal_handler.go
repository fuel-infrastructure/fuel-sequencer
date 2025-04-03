package abci

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sidecarclient "github.com/fuel-infrastructure/fuel-sequencer/sidecar/client"
	sidecartypes "github.com/fuel-infrastructure/fuel-sequencer/sidecar/service/types"
	"github.com/fuel-infrastructure/fuel-sequencer/utils"
	bridgekeeper "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/keeper"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

type FuelSequencerProposalHandler struct {
	cdc                    codec.Codec                     // codec
	valStore               baseapp.ValidatorStore          // to get the current validators' pubkeys
	txVerifier             baseapp.ProposalTxVerifier      // a utility for transaction verification
	sidecar                sidecarclient.AppSidecarClient  // a client to query the Sidecar service
	bridgeKeeper           bridgekeeper.Keeper             // Bridge keeper
	defaultProposalHandler *baseapp.DefaultProposalHandler // gives us access to default proposal handling behaviour
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	cdc codec.Codec,
	valStore baseapp.ValidatorStore,
	txVerifier baseapp.ProposalTxVerifier,
	sidecar sidecarclient.AppSidecarClient,
	bridgeKeeper bridgekeeper.Keeper,
) *FuelSequencerProposalHandler {
	proposalHandler := &FuelSequencerProposalHandler{
		cdc:                    cdc,
		valStore:               valStore,
		txVerifier:             txVerifier,
		sidecar:                sidecar,
		bridgeKeeper:           bridgeKeeper,
		defaultProposalHandler: baseapp.NewDefaultProposalHandler(nil, txVerifier),
	}

	// Set TxSelector to our custom fuelSequencerTxSelector
	proposalHandler.defaultProposalHandler.SetTxSelector(NewFuelSequencerTxSelector())

	return proposalHandler
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
		proposerConsAddress := sdk.ConsAddress(req.ProposerAddress)
		ctx.Logger().Info("preparing proposal", "proposer", proposerConsAddress, "num_txs", len(req.Txs))

		bridgeParams := h.bridgeKeeper.GetParams(ctx)
		injectMsgSupplyDelta := bridgeParams.IsMsgSupplyDeltaBlock(uint64(req.Height))

		// Get the transaction sequences for MsgIndex, MsgSupplyDelta, and the first sequence for event transactions.
		indexSequence, supplyDeltaSequence, firstEventTxsSequence := h.generateTxSequences(ctx, injectMsgSupplyDelta)

		blockedBech32Addresses, err := h.bridgeKeeper.GetAllBlockedBech32Addresses(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get blocked addresses: %w", err)
		}

		// Inject MsgSupplyDeltaTx if expected at current height
		supplyDeltaBytesSize := uint64(0)
		if injectMsgSupplyDelta {
			supplyDeltaBytes, err := bridgetypes.NewMsgSupplyDelta(
				h.bridgeKeeper.GetAuthority(),
			).RawTxBytes(supplyDeltaSequence)
			if err != nil {
				return nil, fmt.Errorf("failed to generate msg supply delta tx: %w", err)
			}

			// Calculate size of supply delta message
			supplyDeltaBytesSize = utils.TxSize(supplyDeltaBytes)

			// Ensure that supply delta fits in the block on its own
			if supplyDeltaBytesSize > uint64(req.MaxTxBytes) {
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
		} else {
			ctx.Logger().Info("received sidecar response at PrepareProposal", "num_events", len(response.Events))
		}

		msgIndex, eventTxs, err := h.generateMsgIndexAndEventTxs(
			ctx, response, ethBlockToQuery, sidecarErr, &bridgeParams, blockedBech32Addresses, firstEventTxsSequence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate MsgIndex and event txs: %w", err)
		}

		// Trim events from head to skip the events that were already processed.
		eventTxs, err = msgIndex.TrimEventsFromHead(eventTxs, ethereumEventIndexOffset)
		if err != nil {
			return nil, fmt.Errorf("failed to trim event txs from head: %w", err)
		}

		// Calculate the block space that should be reserved for event transactions.

		// The SupplyDelta transaction size is deducted because we have already allocated block space for it. We have
		// not deducted the size of MsgIndex because we will check whether it fits the allocated block space when
		// calling msgIndex.NumberOfEventsWithMaxBytes. We should never be in a position where there isn't enough block
		// space for MsgIndex as it is relatively small.
		sequencerTxsSize := utils.TxsSize(req.Txs) - supplyDeltaBytesSize
		maxBlockSpace := uint64(req.MaxTxBytes) - supplyDeltaBytesSize

		// Reserve a percentage of the available block space for Sequencer-native transactions. We are sure that this
		// will not cover the entire block space since there are limits imposed on bridgeParams.SequencerTxsAllocation.
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

			// Otherwise, Sequencer-native transactions are set to occupy at most bridgeParams.SequencerTxsAllocation
			// percent of the available block space, depending on the size of Sequencer-native transactions.
			maxBytesForEvents = max(maxBlockSpace-sequencerTxsSize, maxBlockSpace-sequencerTxsBlockSpace)
		}

		// Trim events from tail to fit the block space allocated for events.
		// NOTE: The TxSelector will be able to fit in more Sequencer-native transactions at the end if there is more
		// space in the block after adjusting the number of event transactions.
		maxNumberOfEvents, err := msgIndex.NumberOfEventsWithMaxBytes(eventTxs, maxBytesForEvents, indexSequence)
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
		// Anything that comes before this point will cause ProcessProposal to error because it expects a MsgIndex tx.

		// Inject MsgIndex and Ethereum event transactions as the first txs in the block.
		msgIndexBz, err := msgIndex.RawTxBytes(indexSequence)
		if err != nil {
			return nil, fmt.Errorf("failed to encode MsgIndex: %w", err)
		}
		req.Txs = append(append([][]byte{msgIndexBz}, eventTxs...), req.Txs...)

		// Use the DefaultProposalHandler. Since we're using a NoOp mempool, it will select the transactions requested
		// by CometBFT which by default should be in FIFO order. The handler ensures that the selected transactions
		// comply with the `req.MaxTxBytes` and `blockParams.MaxGas` limits. Among these transactions are several
		// injected transactions that will consume zero gas.
		resp, err := h.defaultProposalHandler.PrepareProposalHandler()(ctx, req)
		if err != nil {
			req.Txs = [][]byte{}
			return nil, fmt.Errorf("default proposal handler failed with error: %w", err)
		}
		selectedTxs := resp.Txs

		// Check that we've collected the minimum expected transactions
		err = checkMinimumNumTxs(uint64(len(selectedTxs)), msgIndex, injectMsgSupplyDelta)
		if err != nil {
			req.Txs = [][]byte{}
			return nil, err
		}

		ctx.Logger().Debug("prepared proposal", "txs", len(selectedTxs))

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
		expectMsgSupplyDelta := bridgeParams.IsMsgSupplyDeltaBlock(uint64(req.Height))

		// Get the transaction sequences for MsgIndex, MsgSupplyDelta, and the first sequence for event transactions.
		indexSequence, supplyDeltaSequence, firstEventTxsSequence := h.generateTxSequences(ctx, expectMsgSupplyDelta)

		blockedBech32Addresses, err := h.bridgeKeeper.GetAllBlockedBech32Addresses(ctx)
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

		lastEthereumBlockSynced, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		if !found {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"could not get last Ethereum block synced from state",
			)
		}
		ethBlockToQuery := lastEthereumBlockSynced + 1

		// If the proposer fails to synchronize with the Ethereum height LastEthereumBlockSynced + 1, assume their
		// sidecar is out of sync and requires more time to catch up. Therefore, a block proposal is only considered
		// valid if it accurately reflects the absence of event transactions. This is permissible only until
		// MaxEthBlockUpdateDelay is exceeded. Beyond this point, verifiers will expect synchronization with Ethereum.
		//
		// This approach prevents a scenario where consensus among validators cannot be achieved. Some validators may be
		// stuck on a block proposal showing no Ethereum synchronization, while others, whose sidecars indicate they
		// have synchronized with Ethereum, would not agree with that proposal.
		// REF: https://github.com/cometbft/cometbft/blob/8cd4a692d0fbd70f183e28f4507641c799a8f14f/consensus/state.go#L1347-L1352
		var msgIndex *bridgetypes.MsgIndex
		var eventTxs [][]byte
		if injectedMsgIndex.NoEthereumSyncing() {

			// Reject block if MaxEthBlockUpdateDelay is exceeded
			lastEthBlockUpdateTime, found := h.bridgeKeeper.GetLastEthBlockUpdateTime(ctx)
			ethSyncDelayExceeded := found && req.Time.After(
				lastEthBlockUpdateTime.Add(bridgeParams.MaxEthBlockUpdateDelay),
			)
			if ethSyncDelayExceeded {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
					"last syncup with Ethereum was at %s; block time: %s; max delay allowed: %s",
					lastEthBlockUpdateTime.String(),
					req.Time.String(),
					bridgeParams.MaxEthBlockUpdateDelay.String(),
				)
			}

			// Otherwise, generate a MsgIndexTx that indicates no Ethereum syncing
			msgIndex = &bridgetypes.MsgIndex{
				Authority:           h.bridgeKeeper.GetAuthority(),
				NumInjectedEventTxs: 0,
				NewEthereumBlock:    false,
				BlockNumber:         ethBlockToQuery,
			}
			eventTxs = [][]byte{} // No event transactions expected

			ctx.Logger().Info("proposer did not sync with Ethereum; skipping query to sidecar", "height", req.Height)
		} else {

			// Query the events of the next Ethereum block.
			response, sidecarErr := h.sidecar.GetBlockEvents(
				ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: strconv.FormatUint(ethBlockToQuery, 10)},
			)
			if sidecarErr != nil {
				ctx.Logger().Warn("observed sidecar error at ProcessProposal", "err", sidecarErr)
				// This error is also passed to generateMsgIndexAndEventTxs to perform dedicated error handling.
			}

			// Generate the MsgIndex and event transactions based on the queried events of LastEthereumBlockSynced + 1
			msgIndex, eventTxs, err = h.generateMsgIndexAndEventTxs(
				ctx, response, ethBlockToQuery, sidecarErr, &bridgeParams, blockedBech32Addresses, firstEventTxsSequence,
			)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
					"failed to generate MsgIndex and event txs: %w", err,
				)
			}
		}

		ethereumEventIndexOffset, found := h.bridgeKeeper.GetEthereumEventIndexOffset(ctx)
		if !found {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"could not get Ethereum event index offset from state",
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

		// Check that MsgIndex was injected correctly
		err = h.verifyInjectedMsgIndexTx(req.Txs[0], indexSequence, msgIndex)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to verify injected MsgIndexTx: %w", err,
			)
		}

		// Check that event transactions were injected correctly
		err = h.verifyInjectedEventTxs(injectedEventTxs, eventTxs)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to verify injected event txs: %w", err,
			)
		}

		// Check that MsgSupplyDelta was injected correctly if expected
		if expectMsgSupplyDelta {
			err := h.verifyInjectedMsgSupplyDeltaTx(req.Txs, msgIndex.NumInjectedEventTxs, supplyDeltaSequence)
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

		ctx.Logger().Debug("processed proposal", "height", req.Height, "num_txs", len(req.Txs))

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// generateTxSequences returns the transaction sequences that should be used throughout the ProposalHandler:
// - MsgIndex sequence: the next injected txs sequence.
// - MsgSupplyDelta sequence: MsgIndex sequence + 1, if applicable.
// - The sequence for the first event tx is the next available sequence.
func (h *FuelSequencerProposalHandler) generateTxSequences(
	ctx context.Context,
	isMsgSupplyDeltaHeight bool,
) (indexSequence, supplyDeltaSequence, firstEventTxsSequence uint64) {
	nextSequence := h.bridgeKeeper.MustGetNextConsensusTxsSequence(ctx)
	if isMsgSupplyDeltaHeight {
		return nextSequence, nextSequence + 1, nextSequence + 2
	} else {
		return nextSequence, 0, nextSequence + 1
	}
}

// checkMinimumNumTxs ensures that we've collected the minimum number of expected transactions, i.e. MsgIndex, the event
// transactions, and the MsgSupplyDelta if we're at the MsgSupplyDelta height. It returns an error if the check fails.
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

// getNewEthereumBlock returns the value for MsgIndex.NewEthereumBlock, which should be true if the Sidecar didn't error
func (h *FuelSequencerProposalHandler) getNewEthereumBlock(sidecarErr error) bool {
	return sidecarErr == nil
}

// generateMsgIndexAndEventTxs generates MsgIndex and transactions from events based on the response of the sidecar.
// It errors upon invalid events from the sidecar. Returned errors have the capability of halting block production.
func (h *FuelSequencerProposalHandler) generateMsgIndexAndEventTxs(
	ctx sdk.Context,
	sidecarResponse *sidecartypes.QueryBlockEventsResponse,
	blockNumber uint64,
	sidecarErr error,
	params *bridgetypes.Params,
	blockedBech32Addresses map[string]bool,
	eventTxsSequence uint64,
) (msgIndex *bridgetypes.MsgIndex, eventTxs [][]byte, err error) {

	// Set events to nil by default to avoid a null pointer dereference if the Sidecar errors.
	// Context: Sidecar returns a nil response when it errors.
	var events []*sidecartypes.Event
	if sidecarResponse != nil {
		events = sidecarResponse.Events
	}

	// Identify all events in the MsgIndex and produce one new valid transaction per event.
	//
	// Both Deposit and Authorize events are expected to follow a strict structure as they are fully generated by
	// Ethereum smart contracts and proto-encoded by the sidecar. Therefore, we halt block production if any event
	// cannot be encoded as transaction bytes as this would mean that there is a bug either on the Ethereum smart
	// contracts or the sidecar.
	//
	// To protect the Sequencer from attacks induced from invalid payloads we skip an Authorize event if it fails
	// authentication (incl. ValidateBasic and signer check).
	for _, event := range events {

		// Validation checks at this stage are of critical nature and should halt block production upon failure.
		err = event.Validate(params.EthereumProxyContractAddress)
		if err != nil {
			return nil, nil, fmt.Errorf("encountered invalid event with err: %s; event: %s", err.Error(), event)
		}

		eventTx, err := event.RawTxBytes(h.cdc, h.bridgeKeeper.GetAuthority(), eventTxsSequence)
		if err != nil {

			// Return an error because all events are expected to be successfully encoded as transaction bytes. This
			// will halt block production.
			return nil, nil, fmt.Errorf(
				"failed to encode event as raw tx bytes with err: %s; event: %s",
				err.Error(), event,
			)
		}

		authenticated, err := h.authenticateEvent(event, eventTx, blockedBech32Addresses)
		if err != nil {

			// Authentication errors are unexpected and should halt block production.
			return nil, nil, fmt.Errorf("failed to check authorization: %s; event: %s", err.Error(), event)
		}

		var authenticatedTx []byte
		if authenticated {
			authenticatedTx = eventTx
		} else {
			errStr := "unauthorized event"
			ctx.Logger().Warn(fmt.Sprintf("skipping %s", errStr))

			authenticatedTx, err = h.generateSkipTxBytes(errStr, event, blockNumber, eventTxsSequence)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to generate skip tx bytes: %w", err)
			}
		}
		eventTxs = append(eventTxs, authenticatedTx)
		eventTxsSequence += 1 // increment the sequence since we've officially included the eventTx
	}

	// Generate MsgIndex based on the number of injected events.
	msgIndex = &bridgetypes.MsgIndex{
		Authority:           h.bridgeKeeper.GetAuthority(),
		NumInjectedEventTxs: uint64(len(eventTxs)),
		NewEthereumBlock:    h.getNewEthereumBlock(sidecarErr),
		BlockNumber:         blockNumber,
	}

	// Sanity check: number of events equal generated txs
	if len(eventTxs) != len(events) {
		return nil, nil, fmt.Errorf(
			"mismatch between no. of events to be injected and no. of events extracted from Ethereum; had: %d, got: %d",
			len(events), len(eventTxs),
		)
	}

	return
}

// verifyInjectedMsgIndexTx is used by ProcessProposal to check whether MsgIndexTx was injected properly.
func (h *FuelSequencerProposalHandler) verifyInjectedMsgIndexTx(
	injectedMsgIndexTx []byte,
	expectedSequence uint64,
	generatedMsgIndex *bridgetypes.MsgIndex,
) error {

	generatedMsgIndexTx, err := generatedMsgIndex.RawTxBytes(expectedSequence)
	if err != nil {
		return fmt.Errorf("failed to encode MsgIndex: %w", err)
	}

	if !bytes.Equal(injectedMsgIndexTx, generatedMsgIndexTx) {
		return fmt.Errorf(
			"generated MsgIndex tx differs from that of the block proposal (injected: %X) (generated: %X)",
			injectedMsgIndexTx, generatedMsgIndexTx,
		)
	}
	return nil
}

// verifyInjectedEventTxs is used by ProcessProposal to check whether the event transactions were injected properly.
// Additional details about the event transactions is included in the error since this might be useful during debugging.
func (h *FuelSequencerProposalHandler) verifyInjectedEventTxs(injected, generated [][]byte) error {

	if !utils.IsEqualBytesSlices(injected, generated) {

		// Return a specific error if the number of injected transactions does not match the number generated by the
		// verifier.
		if len(injected) != len(generated) {
			return fmt.Errorf(
				"generated event txs do not match those from the proposal (num_injected: %d) (num_generated: %d)",
				len(injected), len(generated),
			)
		}

		return fmt.Errorf(
			"generated event txs do not match those from the proposal (size(injected): %d) (size(generated): %d)",
			utils.TxsSize(injected), utils.TxsSize(generated),
		)
	}
	return nil
}

// verifyInjectedMsgSupplyDeltaTx is used by ProcessProposal to check whether MsgSupplyDeltaTx was injected properly
func (h *FuelSequencerProposalHandler) verifyInjectedMsgSupplyDeltaTx(
	txs [][]byte, numInjectedTxs, expectedSequence uint64,
) error {

	// We expect the MsgSupplyDeltaTx to be at the index right after MsgIndex and any injected events.
	// If no events were injected, the index will be 1, i.e. right after the MsgIndex.
	msgSupplyDeltaIndex := 1 + numInjectedTxs

	// Check that there's enough transactions for the above index to make sense.
	if uint64(len(txs)) < 1+msgSupplyDeltaIndex {
		return fmt.Errorf("expected at least %d transactions in block proposal", 1+msgSupplyDeltaIndex)
	}
	injectedMsgSupplyDeltaTx := txs[msgSupplyDeltaIndex]

	generatedMsgSupplyDeltaTx, err := bridgetypes.NewMsgSupplyDelta(
		h.bridgeKeeper.GetAuthority(),
	).RawTxBytes(expectedSequence)
	if err != nil {
		return fmt.Errorf("failed to generate MsgSupplyDelta tx: %w", err)
	}

	if !bytes.Equal(injectedMsgSupplyDeltaTx, generatedMsgSupplyDeltaTx) {
		return fmt.Errorf(
			"generated MsgSupplyDelta tx differs from that of the block proposal (injected: %X) (generated: %X)",
			injectedMsgSupplyDeltaTx, generatedMsgSupplyDeltaTx,
		)
	}

	return nil
}

func (h *FuelSequencerProposalHandler) generateSkipTxBytes(
	errStr string, event *sidecartypes.Event, blockNumber uint64, eventTxsSequence uint64,
) ([]byte, error) {
	msg, trimmed := bridgetypes.NewMsgSkippedEventTx(
		h.bridgeKeeper.GetAuthority(),
		errStr,
		blockNumber,
		event.LogIndex,
		event.TxIndex,
		event.TxHash,
	)

	if trimmed {
		h.bridgeKeeper.Logger().Warn(
			"reason_for_skip exceeded maximum length and was trimmed",
			"original_length", len(errStr),
			"max_length", bridgetypes.MaxReasonLength,
		)
	}

	return msg.RawTxBytes(eventTxsSequence)
}
