package abci

import (
	"errors"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// TODO: Any required objects need to go here
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	logger log.Logger,
	valStore baseapp.ValidatorStore,
	txVerifier baseapp.ProposalTxVerifier,
	sidecar sidecarclient.AppSidecarClient,
	bridgeKeeper bridgekeeper.Keeper,
) *FuelSequencerProposalHandler {
	// TODO: Add any required parameters
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
		blockHeight, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		if !found {
			return nil, errors.New("could not get last Ethereum block synced from state")
		}

		// Query the events of the next Ethereum block
		ethBlockToQuery := blockHeight.Add(sdkmath.OneInt()).String()
		response, err := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: ethBlockToQuery},
		)
		if err != nil {
			// Any error returned from the sidecar will cause the block proposer to return req.Txs and thus the block
			// proposer will panic because it won't find EthEventsTx at index zero. As a result, a new consensus round
			// should be generated. This may occur when the Sidecar is not catching up with Ethereum, Sequencer is too
			// fast or connection issues with the sidecar, among other potential situations not specifically mentioned.
			return nil, fmt.Errorf("failed to query sidecar at block %s: %w", ethBlockToQuery, err)
		}

		// TODO: Add MsgSupplyDelta logic and inject transaction at index 1 if we expect a MsgSupplyDelta

		ethEventsTx, err := h.generateEthEventsTx(response)
		if err != nil {
			return nil, fmt.Errorf("failed to generate eth events tx: %w", err)
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

		// Since we are assuming a NoOp mempool we simply return the transactions requested from CometBFT, which, by
		// default, should be in FIFO order. Note, we still need to ensure the transactions returned respect
		// req.MaxTxBytes and blockParams.MaxGas
		for index, txBz := range req.Txs {

			// EthEventsTx will not satisfy sdk.Tx. Therefore, we cannot handle it the same way as we handle other
			// transactions. Note: Here we are assuming that EthEventsTx is injected at index zero and that we will
			// always have such a tx.
			if index == 0 {
				success := h.txSelector.SelectNonSDKTxForProposal(ctx, uint64(req.MaxTxBytes), txBz)
				if !success {
					return nil, errors.New("failed to add eth events transaction to block proposal")
				}
				continue
			}

			tx, err := h.txVerifier.TxDecode(txBz)
			if err != nil {
				return nil, err
			}

			stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
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
		if len(req.Txs) < 1 {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"block proposal doesn't have any transactions first transaction expected to be an eth events tx",
			)
		}

		// Expect that the first transaction is always the EthEventsTx
		var injectedEthEventsTx bridgetypes.EthEventsTx
		if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"first transaction expected to be an eth events tx: %w", err,
			)
		}

		blockHeight, found := h.bridgeKeeper.GetLastEthereumBlockSynced(ctx)
		if !found {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, errors.New(
				"could not get last Ethereum block synced from state",
			)
		}

		// Query the events of the next Ethereum block
		ethBlockToQuery := blockHeight.Add(sdkmath.OneInt()).String()
		response, err := h.sidecar.GetBlockEvents(
			ctx, &sidecartypes.QueryBlockEventsRequest{BlockNumber: ethBlockToQuery},
		)
		if err != nil {
			// Any error returned from the sidecar should cause the validator to fail in processing the block proposal.
			// A new consensus round is generated if more than 2/3s of the validator set errors. An error at this stage
			// can occur when if Sidecar is not catching up with Ethereum, Sequencer is too fast or connection issues
			// with the sidecar, among other potential situations not specifically mentioned.
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to query sidecar at block %s: %w", ethBlockToQuery, err,
			)
		}

		// Generate the EthEventsTx that should be included at index 0 in the block proposal
		ethEventsTx, err := h.generateEthEventsTx(response)
		if err != nil {
			return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, fmt.Errorf(
				"failed to generate eth events tx: %w", err,
			)
		}

		// Check that the injected EthEventsTx matches the one generated by the validator verifying the block proposal
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

		// TODO: Add MsgSupplyDelta verification logic at index 1 if we expect a MsgSupplyDelta

		var totalTxGas uint64

		var maxBlockGas int64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = b.MaxGas
		}

		for index, txBytes := range req.Txs {
			// Eth events transactions typically don't implement sdk.Tx, therefore, they can be skipped since no gas
			// is consumed. Note: Here we are assuming that eth events txs are always injected at index zero.
			if index == 0 {
				continue
			}

			// There is something wrong with the Tx if it cannot be decoded. Thus reject the block proposal.
			tx, err := h.txVerifier.TxDecode(txBytes)
			if err != nil {
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
) (*bridgetypes.EthEventsTx, error) {
	ethEventsTx := bridgetypes.EthEventsTx{Events: sidecarResponse.Events}
	if err := ethEventsTx.ValidateBasic(); err != nil {
		return nil, err
	}

	return &ethEventsTx, nil
}

// PreBlocker contains logic that should run before any FinalizeBlock logic. FinalizeBlock ignores any byte slices not
// implementing sdk.Tx. As a consequence, any important results originating from PrepareProposal or ProcessProposal not
// implementing sdk.Tx need to be made available to the modules in storage at PreBlocker stage.
func (h *FuelSequencerProposalHandler) PreBlocker(
	_ sdk.Context, req *abci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	// TODO: This should be adapted as per application requirements
	// This check is done for completeness’s sake as we should not expect to run into this scenario
	if len(req.Txs) == 0 {
		return nil, fmt.Errorf("expected eth events transaction to be injected")
	}

	// TODO: Check if certain transactions are expected at this stage ex MsgSupplyDelta at specific epochs

	// TODO: This was done for demonstration purposes and should be adapted as per application requirements.
	var injectedEthEventsTx bridgetypes.EthEventsTx
	if err := injectedEthEventsTx.Unmarshal(req.Txs[0]); err != nil {
		return nil, fmt.Errorf("failed to decode injected eth events tx: %w", err)
	}

	// TODO: Custom logic like storing "special" transactions in state

	h.logger.Debug("finished executing pre-block hook")

	return &sdk.ResponsePreBlock{}, nil
}
