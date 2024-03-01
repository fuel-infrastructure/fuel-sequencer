package abci

import (
	"encoding/json"
	"errors"
	"fmt"

	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type FuelSequencerProposalHandler struct {
	logger     log.Logger
	valStore   baseapp.ValidatorStore     // to get the current validators' pubkeys
	txSelector TxSelector                 // a utility for checking whether a Tx can be included in the proposal
	txVerifier baseapp.ProposalTxVerifier // a utility for transaction verification

	// Any required objects need to go here
}

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(
	logger log.Logger, valStore baseapp.ValidatorStore, txVerifier baseapp.ProposalTxVerifier,
) *FuelSequencerProposalHandler {
	return &FuelSequencerProposalHandler{
		logger:     logger,
		valStore:   valStore,
		txVerifier: txVerifier,
		txSelector: NewFuelSequencerTxSelector(),
	}
}

// PrepareProposalHandler defines the logic that is executed by the block proposer when they are crafting a block. This
// function needs to obey the following rules:
//
// 1. Can be non-deterministic
// 2. Transaction size cannot exceed RequestPrepareProposal.MaxTxBytes
// 3. Block gas cannot exceed BlockParams.MaxGas
//
// Note: Here we are assuming that the NoOp mempool is to be used, meaning that Txs requested from CometBFT will simply
// be returned and not verified. It is also recommended that the ProcessProposalHandler implements any verifications
// that PrepareProposalHandler is implementing, therefore, for the NoOp mempool ProcessProposalHandler shouldn't
// implement any verification checks.
// In the case that a different mempool is implemented we must perform extra Tx verification checks in
// PrepareProposalHandler and ProcessProposalHandler because we are no longer relying on CometBFT, and thus we must
// ensure that we are including valid transactions.
//
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L199-L303
func (h *FuelSequencerProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		if VoteExtensionsEnabled(ctx) {
			// Perform basic checks on the vote extensions
			err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			// TODO: This should be removed as it was implemented for demonstration purposes
			aggregatedOracleData, err := h.aggregateVotesIntoParsedOracleData(ctx, req.LocalLastCommit)
			if err != nil {
				return nil, fmt.Errorf("failed to aggregate votes into parsed oracle data: %w", err)
			}
			voteExtTxBz, err := aggregatedOracleData.Marshal()
			if err != nil {
				return nil, fmt.Errorf("failed to encode injected vote extension tx: %w", err)
			}
			// It is important txs that cannot be decoded into sdk.Tx are appended at entry 0.
			req.Txs = append([][]byte{voteExtTxBz}, req.Txs...)

			// TODO: Define custom logic here
		}

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

			// Vote extension trsanctions typically do not satisfy sdk.Tx. Therefore, we cannot handle them the same way
			// we handle other transactions.
			// Note: Here we are assuming that vote extension transactions are inputted at index zero.
			// TODO: Check if vote extension was expected if done periodically
			if index == 0 {
				success := h.txSelector.SelectVETxForProposal(ctx, uint64(req.MaxTxBytes), txBz)
				if !success {
					return nil, errors.New("failed to add vote extension transaction to block proposal")
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

		h.logger.Info(
			"prepared proposal",
			"txs", len(h.txSelector.SelectedTxs(ctx)),
			"vote_extensions_enabled", VoteExtensionsEnabled(ctx),
		)

		return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
	}
}

// ProcessProposalHandler defines the logic which confirms the validity of a block proposal. This logic is executed
// by all validators, and it needs to obey the following rules:
//
// 1. Must be deterministic
// 2. Block gas cannot exceed BlockParams.MaxGas
//
// Note: Here we are assuming that the NoOp mempool is to be used, meaning that PrepareProposal may have included some
// Txs that might fail verification as it relies on CometBFT. The logic implemented here attempts to perform the exact
// verifications performed by the PrepareProposalHandler as suggested by the Cosmos SDK documentation. Unfortunately, we
// cannot verify whether transaction sizes exceed the limit because MaxTxBytes is not part of
// abci.RequestProcessProposal. That being said, it seems that CometBFT validates the Tx sizes returned by the
// application https://github.com/cometbft/cometbft/blob/719b64156aaa3cb89add29d053439060f8e420dd/spec/consensus/creating-proposal.md?plain=1#L41
// It is also important to highlight that the default implementation of ProcessProposalHandler doesn't perform any
// verifications for NoOp mempools, however, we have included a block gas max limit check to follow recommendations
// stated in the documentation.
//
// Please refer to the following default handler implementation on Cosmos SDK main branch if in doubt:
// https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L305-L352
func (h *FuelSequencerProposalHandler) ProcessProposalHandler() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
		if VoteExtensionsEnabled(ctx) {
			// TODO: This should be removed as it was implemented for demonstration purposes
			var injectedVoteExtTx AggregatedOracleData
			if err := injectedVoteExtTx.Unmarshal(req.Txs[0]); err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// TODO: We must confirm that all vote extensions satisfy the application's logic because validators do not
			//       share the same view of what vote extensions they verify when VerifyVoteExtension is called as this
			//       depends on how vote extensions are propagated.

			// Perform basic checks on vote extensions
			err := baseapp.ValidateVoteExtensions(
				ctx, h.valStore, req.Height, ctx.ChainID(), injectedVoteExtTx.ExtendedCommitInfo,
			)
			if err != nil {
				return nil, err
			}

			// TODO: This should be removed as it was implemented for demonstration purposes
			aggregatedOracleData, err := h.aggregateVotesIntoParsedOracleData(ctx, injectedVoteExtTx.ExtendedCommitInfo)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}
			if injectedVoteExtTx.ParsedOracleData != aggregatedOracleData.ParsedOracleData ||
				injectedVoteExtTx.Msgs != aggregatedOracleData.Msgs {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// TODO: Define custom logic here
		}

		var totalTxGas uint64

		var maxBlockGas int64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = b.MaxGas
		}

		for index, txBytes := range req.Txs {
			// Vote extension transactions are typically not a transaction, therefore, they can be skipped since no gas
			// is consumed. Note: Here we are assuming that vote extension transactions are inputted at index zero.
			// TODO: Check if vote extension was expected if done periodically
			if index == 0 {
				continue
			}

			// There is something wrong with the Tx if it cannot be decoded. Thus reject the block proposal.
			tx, err := h.txVerifier.TxDecode(txBytes)
			if err != nil {
				return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
			}

			// Confirm that the block's max gas limit is not exceeded
			if maxBlockGas > 0 {
				gasTx, ok := tx.(baseapp.GasTx)
				if ok {
					totalTxGas += gasTx.GetGas()
				}

				if totalTxGas > uint64(maxBlockGas) {
					return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_REJECT}, nil
				}
			}
		}

		h.logger.Info(
			"processing proposal",
			"height", req.Height,
			"num_txs", len(req.Txs),
			"vote_extensions_enabled", VoteExtensionsEnabled(ctx),
		)

		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}
}

// TODO: This function should be replaced with our application logic
func (h *FuelSequencerProposalHandler) aggregateVotesIntoParsedOracleData(
	_ sdk.Context, localLastCommit abci.ExtendedCommitInfo,
) (*AggregatedOracleData, error) {
	// Just use first VoteExtension; We could check that all VEs match.
	var voteExtension CustomOracleVoteExtension
	if err := json.Unmarshal(localLastCommit.Votes[0].VoteExtension, &voteExtension); err != nil {
		return nil, fmt.Errorf("failed to json.unmarshal vote extension: %w", err)
	}

	return &AggregatedOracleData{
		ParsedOracleData:   fmt.Sprintf("PARSED { %s }", string(voteExtension.Data.Data)),
		Msgs:               voteExtension.Data.Msgs,
		ExtendedCommitInfo: localLastCommit,
	}, nil
}

// PreBlocker contains logic that should run before any other logic during FinalizeBlock. FinalizeBlock ignores
// any byte slices that don't implement sdk.Tx. As a result, any important results originating from PrepareProposal or
// ProcessProposal that don't implement sdk.Tx need to be made available to the modules in storage at this stage.
func (h *FuelSequencerProposalHandler) PreBlocker(
	ctx sdk.Context, req *abci.RequestFinalizeBlock,
) (*sdk.ResponsePreBlock, error) {
	if len(req.Txs) == 0 {
		return &sdk.ResponsePreBlock{}, nil
	}

	if VoteExtensionsEnabled(ctx) {

		// TODO: Check if certain transactions are expected at this stage

		// TODO: This was done for demonstration purposes and should be adapted as per application requirements
		var injectedVoteExtTx AggregatedOracleData
		if err := json.Unmarshal(req.Txs[0], &injectedVoteExtTx); err != nil {
			return &sdk.ResponsePreBlock{}, fmt.Errorf("failed to decode injected vote extension tx: %w", err)
		}

		// TODO: Custom logic like storing "special" transactions in state
	}

	h.logger.Info("finished executing pre-block hook")

	return &sdk.ResponsePreBlock{}, nil
}
