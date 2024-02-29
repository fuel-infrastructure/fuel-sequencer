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

// NewFuelSequencerProposalHandler defines a custom FuelSequencer proposal handler object
func NewFuelSequencerProposalHandler(logger log.Logger) *FuelSequencerProposalHandler {
	return &FuelSequencerProposalHandler{
		logger: logger,
	}
}

// PrepareProposalHandler defines the logic that is executed by the block proposer when they are crafting a block. This
// function needs to obey the following rules:
//
// 1. Can be non-deterministic
// 2. A transaction cannot exceed RequestPrepareProposal.MaxTxBytes
// 3. Block gas cannot exceed BlockParams.MaxGas
// 4. If vote extensions are enabled it will get the vote extensions of the previous block in req.LocalLastCommit
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

		h.logger.Info(
			"prepared proposal",
			"txs", len(resp.Txs),
			"vote_extensions_enabled", voteExtensionsEnabled,
		)

		return &abci.ResponsePrepareProposal{Txs: h.txSelector.SelectedTxs(ctx)}, nil
	}
}
