package abci

import (
	"cosmossdk.io/log"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
)

type FuelSequencerProposalHandler struct {
	logger   log.Logger
	valStore baseapp.ValidatorStore // to get the current validators' pubkeys

	// Any required objects need to go here
}

// NewFuelSequencerProposalHandler defines a new FuelSequencerProposalHandler object
func NewFuelSequencerProposalHandler(logger log.Logger) FuelSequencerProposalHandler {
	return FuelSequencerProposalHandler{
		logger: logger,
	}
}

// PrepareProposalHandler defines the logic that is executed by the block proposer when they are crafting a block. This
// function needs to obey the following rules:
//
// 1. Can be non-deterministic
// 2. A transaction cannot exceed RequestPrepareProposal.MaxTxBytes
// 3. Block gas cannot exceed BlockParams.MaxGas
//
// Note: Here we are assuming that the NoOp mempool is to be used. In the case that a different mempool is implemented
// we must refer to the default implementation below in order to add some mempool specific logic, if required.
// Default Implementation: https://github.com/cosmos/cosmos-sdk/blob/7e6948f50cd4838a0161838a099f74e0b5b0213c/baseapp/abci_utils.go#L198
func (h *FuelSequencerProposalHandler) PrepareProposalHandler() sdk.PrepareProposalHandler {
	return func(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
		if VoteExtensionsEnabled(ctx) {
			err := baseapp.ValidateVoteExtensions(ctx, h.valStore, req.Height, ctx.ChainID(), req.LocalLastCommit)
			if err != nil {
				return nil, err
			}

			// TODO: Define custom logic here
		}

		// TODO: Do NoOp mempool part

		var maxBlockGas uint64
		if b := ctx.ConsensusParams().Block; b != nil {
			maxBlockGas = uint64(b.MaxGas)
		}

		defer h.txSelector.Clear()

		// If the mempool is nil or NoOp we simply return the transactions
		// requested from CometBFT, which, by default, should be in FIFO order.
		//
		// Note, we still need to ensure the transactions returned respect req.MaxTxBytes.
		_, isNoOp := h.mempool.(mempool.NoOpMempool)
		if h.mempool == nil || isNoOp {
			for _, txBz := range req.Txs {
				tx, err := h.txVerifier.TxDecode(txBz)
				if err != nil {
					return nil, err
				}

				stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, tx, txBz)
				if stop {
					break
				}
			}

			return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
		}

		iterator := h.mempool.Select(ctx, req.Txs)
		for iterator != nil {
			memTx := iterator.Tx()

			// NOTE: Since transaction verification was already executed in CheckTx,
			// which calls mempool.Insert, in theory everything in the pool should be
			// valid. But some mempool implementations may insert invalid txs, so we
			// check again.
			txBz, err := h.txVerifier.PrepareProposalVerifyTx(memTx)
			if err != nil {
				err := h.mempool.Remove(memTx)
				if err != nil && !errors.Is(err, mempool.ErrTxNotFound) {
					return nil, err
				}
			} else {
				stop := h.txSelector.SelectTxForProposal(ctx, uint64(req.MaxTxBytes), maxBlockGas, memTx, txBz)
				if stop {
					break
				}
			}

			iterator = iterator.Next()
		}

		h.logger.Info(
			"prepared proposal",
			"txs", len(resp.Txs),
			"vote_extensions_enabled", voteExtensionsEnabled,
		)

		return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
	}
}
