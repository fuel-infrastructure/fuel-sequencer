package abci

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxSelector defines a helper type that assists in selecting transactions during mempool transaction selection in
// PrepareProposal. It keeps track of the total number of bytes and total gas of the selected transactions. It also
// keeps track of the selected transactions themselves.
// NOTE: This struct was copied over from the Cosmos SDK because we need special handling for vote extensions
// Reference: https://github.com/cosmos/cosmos-sdk/blob/a86a83f761383c1ea434925cddd199cd5a271303/baseapp/abci_utils.go#L390-L454
type TxSelector interface {
	// SelectedTxs should return a copy of the selected transactions.
	SelectedTxs(ctx context.Context) [][]byte

	// Clear should clear the TxSelector, nulling out all relevant fields.
	Clear()

	// SelectTxForProposal should attempt to select a transaction for inclusion in a proposal based on inclusion
	// criteria defined by the TxSelector. It must return <true> if the caller should halt the transaction selection
	// loop (typically over a mempool) or <false> otherwise.
	SelectTxForProposal(ctx context.Context, maxTxBytes, maxBlockGas uint64, memTx sdk.Tx, txBz []byte) bool

	// SelectVETxForProposal should attempt to select a vote extension tx for inclusion in a proposal based on inclusion
	// criteria defined by the TxSelector. It must return <true> if the vote extension was added to the block proposal
	// or <false> otherwise.
	SelectVETxForProposal(_ context.Context, maxTxBytes uint64, veBz []byte) bool
}

type fuelSequencerTxSelector struct {
	totalTxBytes uint64
	totalTxGas   uint64
	selectedTxs  [][]byte
}

func NewFuelSequencerTxSelector() TxSelector {
	return &fuelSequencerTxSelector{}
}

func (ts *fuelSequencerTxSelector) SelectedTxs(_ context.Context) [][]byte {
	txs := make([][]byte, len(ts.selectedTxs))
	copy(txs, ts.selectedTxs)
	return txs
}

func (ts *fuelSequencerTxSelector) Clear() {
	ts.totalTxBytes = 0
	ts.totalTxGas = 0
	ts.selectedTxs = nil
}

func (ts *fuelSequencerTxSelector) SelectTxForProposal(
	_ context.Context, maxTxBytes, maxBlockGas uint64, memTx sdk.Tx, txBz []byte,
) bool {
	txSize := uint64(len(txBz))

	var txGasLimit uint64
	if memTx != nil {
		if gasTx, ok := memTx.(baseapp.GasTx); ok {
			txGasLimit = gasTx.GetGas()
		}
	}

	// only add the transaction to the proposal if we have enough capacity
	if (txSize + ts.totalTxBytes) <= maxTxBytes {
		// If there is a max block gas limit, add the tx only if the limit has
		// not been met.
		if maxBlockGas > 0 {
			if (txGasLimit + ts.totalTxGas) <= maxBlockGas {
				ts.totalTxGas += txGasLimit
				ts.totalTxBytes += txSize
				ts.selectedTxs = append(ts.selectedTxs, txBz)
			}
		} else {
			ts.totalTxBytes += txSize
			ts.selectedTxs = append(ts.selectedTxs, txBz)
		}
	}

	// check if we've reached capacity; if so, we cannot select any more transactions
	return ts.totalTxBytes >= maxTxBytes || (maxBlockGas > 0 && (ts.totalTxGas >= maxBlockGas))
}

func (ts *fuelSequencerTxSelector) SelectVETxForProposal(_ context.Context, maxTxBytes uint64, veTxBz []byte) bool {
	veTxSize := uint64(len(veTxBz))

	// only add the transaction to the proposal if we have enough capacity. Note that vote extension transactions do not
	// consume any gas
	if (veTxSize + ts.totalTxBytes) <= maxTxBytes {
		ts.totalTxBytes += veTxSize
		ts.selectedTxs = append(ts.selectedTxs, veTxBz)
		return true
	}

	// if we reach this point it means that the vote extension transaction was not selected
	return false
}

// VoteExtensionsEnabled determines if vote extensions are enabled for the current block. If vote extensions are enabled
// at height h, then a proposer will receive vote extensions in height h+1. This is primarily utilized by any module
// that needs to make state changes based on whether vote extensions have been included in a proposal.
func VoteExtensionsEnabled(ctx sdk.Context) bool {
	cp := ctx.ConsensusParams()
	if cp.Abci == nil || cp.Abci.VoteExtensionsEnableHeight == 0 {
		return false
	}

	return cp.Abci.VoteExtensionsEnableHeight < ctx.BlockHeight()
}
